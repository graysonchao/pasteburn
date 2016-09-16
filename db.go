package pasteburn

import (
	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	uuid "github.com/nu7hatch/gouuid"
)

// DatabaseService defines operations on the backing database.
type DatabaseService interface {
	SaveDocument(*Document) error
	LoadDocument(uuid.UUID) (*Document, error)
}

// BoltDBService implements DatabaseService using Bolt.
// The buckets map contains identifiers for all buckets.
type BoltDBService struct {
	dbPath  string
	buckets map[string][]byte
}

// NewBoltDBService returns an initialized BoltDBService.
func NewBoltDBService(dbPath string) (*BoltDBService, error) {
	b := &BoltDBService{
		dbPath: dbPath,
		buckets: map[string][]byte{
			"documents": []byte("Documents"),
		},
	}
	if err := b.initDb(); err != nil {
		return b, err
	}
	return b, nil
}

// InitDb sets up the database by creating any buckets specified in the service.
// Failing to create a bucket is a fatal error and any uncreated buckets at that point
// will not be created.
func (s BoltDBService) initDb() error {
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
func (s BoltDBService) LoadDocument(id uuid.UUID) (*Document, error) {

	db, err := bolt.Open(s.dbPath, 0600, nil)
	defer db.Close()

	var value []byte

	if err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.buckets["documents"])

		l := b.Get(id[:])
		value = make([]byte, len(l))
		copy(value, l)

		return b.Delete(id[:])
	}); err != nil {
		return nil, err
	}

	d := &Document{
		ID:       id,
		Contents: value,
	}

	return d, nil
}

// SaveDocument saves a document to the database.
func (s BoltDBService) SaveDocument(d *Document) error {
	db, err := bolt.Open(s.dbPath, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	key := make([]byte, len(d.ID))
	copy(key, d.ID[:])

	if err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.buckets["documents"])
		err = b.Put(key, d.Contents)
		return err
	}); err != nil {
		log.WithField("function", "saveToDb").Fatal(err)
		return err
	}

	return nil
}
