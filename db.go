package pasteburn

import (
	"github.com/boltdb/bolt"
	uuid "github.com/nu7hatch/gouuid"
)

// DatabaseService defines operations on the backing database.
type DatabaseService interface {
	InitDb() error
	SaveDocument(uuid.UUID, Document) error
	LoadDocument(uuid.UUID) (Document, error)
}

// BoltService implements DatabaseService using Bolt.
// The buckets array contains identifiers for all buckets.
type BoltService struct {
	dbPath  string
	buckets map[string][]byte
}

// InitDb sets up the database by creating any buckets specified in the service.
// Failing to create a bucket is a fatal error and any uncreated buckets at that point
// will not be created.
func (s *BoltService) InitDb() error {
	db, err := bolt.Open(s.dbPath, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Update(func(tx *bolt.Tx) error {
		for _, bn := range s.buckets {
			_, err := tx.CreateBucketIfNotExists(bn)
			if err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// LoadDocument returns a Document object loaded from Bolt.
func (s *BoltService) LoadDocument(key uuid.UUID) ([]byte, error) {
	db, err := bolt.Open(s.dbPath, 0600, nil)
	defer db.Close()

	var value []byte

	if err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.buckets["documents"])

		l := b.Get(key[:])
		value = make([]byte, len(l))
		copy(value, l)

		return b.Delete(key[:])
	}); err != nil {
		return nil, err
	}

	return value, nil
}

// SaveDocument saves a document to the database.
func (s *BoltService) SaveDocument(uuid uuid.UUID, value []byte) error {
	db, err := bolt.Open(s.dbPath, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	key := make([]byte, len(uuid))
	copy(key, uuid[:])

	if err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.buckets["documents"])
		err = b.Put(key, value)
		return err
	}); err != nil {
		s.log.WithField("function", "saveToDb").Fatal(err)
		return err
	}

	return nil
}
