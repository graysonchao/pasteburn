package pasteburn

import (
	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/nu7hatch/gouuid"
)

// Service is a CRUD interface for Pasteburn documents.
type Service interface {
	PostDocument(ctx context.Context, d *Document) error
	GetDocument(ctx context.Context, id uuid.UUID, key []byte) (*Document, error)
	PostMultiDoc(ctx context.Context, d *MultiDoc) error
	GetMultiDoc(ctx context.Context, id uuid.UUID, key []byte) (*Document, error)
}

// BoltBackedService uses Boltdb to implement Service
type BoltBackedService struct {
	db *BoltDBService
}

// GenerateKey returns a random AES256 key.
func GenerateKey() ([]byte, error) {
	key := make([]byte, AES256KeySizeBytes)
	_, err := rand.Read(key)
	if err != nil {
		return key, err
	}
	return key, nil
}

// NewBoltBackedService returns an initialized Service
func NewBoltBackedService(dbPath string) (*BoltBackedService, error) {
	dbSvc, err := NewBoltDBService(dbPath)
	if err != nil {
		return nil, err
	}

	return &BoltBackedService{
		db: dbSvc,
	}, nil
}

// PostDocument handles posting a document to the DB
func (s *BoltBackedService) PostDocument(ctx context.Context, d *Document) error {
	if err := d.SaveDoc(*s.db); err != nil {
		return err
	}
	return nil
}

// GetDocument returns the note with the given id, decrypted using key
func (s *BoltBackedService) GetDocument(ctx context.Context, id uuid.UUID, key []byte) (*Document, error) {

	d, err := (*s.db).LoadDocument(id)
	if err != nil {
		log.WithFields(log.Fields{
			"id": "id",
		}).Fatal("Failed to load document:", err)
		return d, err
	}

	log.WithFields(log.Fields{
		"id": id,
	}).Debug("Loaded encrypted document")

	if len(d.Contents) == 0 {
		// TODO Document has been deleted how to handle this better?
		return d, nil
	}

	if err := d.DecryptInPlace(key); err != nil {
		log.WithFields(log.Fields{
			"id":      id,
			"keyHash": sha256.Sum256(key),
		}).Fatal("Failed to decrypt document")
		return nil, err
	}

	return d, nil
}

// PostMultiDoc handles posting a document to the DB
func (s *BoltBackedService) PostMultiDoc(ctx context.Context, d *MultiDoc) error {
	if err := d.SaveMD(*s.db); err != nil {
		return err
	}
	return nil
}

// GetMultiDoc loads a single instance of a document given a key.
// The key's first byte which specifies which copy of the doc to load
func (s *BoltBackedService) GetMultiDoc(ctx context.Context, id uuid.UUID, key []byte) (*Document, error) {

	idx := key[0]
	encKey := make([]byte, len(key)-1)
	copy(encKey, key[1:])

	d, err := (*s.db).LoadMultiDoc(id, idx)
	if err != nil {
		return d, err
	}

	if len(d.Contents) == 0 {
		// TODO Document has been deleted how to handle this better?
		return d, nil
	}

	if err = d.DecryptInPlace(encKey); err != nil {
		return d, err
	}

	return d, nil
}
