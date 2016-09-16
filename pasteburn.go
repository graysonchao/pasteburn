package pasteburn

import (
	"crypto/sha256"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/nu7hatch/gouuid"
)

// Service is a CRUD interface for Pasteburn documents.
type Service interface {
	PostDocument(ctx context.Context, d *Document) error
	GetDocument(ctx context.Context, id uuid.UUID, key []byte) (*Document, error)
}

// BoltBackedService uses Boltdb to implement Service
type BoltBackedService struct {
	db *BoltDBService
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
	if err := d.Save(*s.db); err != nil {
		return err
	}
	return nil
}

// Save a Document, assigning it a UUID and returning that UUID.
func (d *Document) Save(db DatabaseService) error {

	if err := db.SaveDocument(d); err != nil {
		log.WithFields(log.Fields{}).Fatal("Failed to save document")
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
	}).Debug("Loaded encrypted note")

	if len(d.Contents) == 0 {
		// TODO Document has been deleted how to handle this better?
		return d, nil
	}

	if err := d.DecryptInPlace(key); err != nil {
		log.WithFields(log.Fields{
			"id":      id,
			"keyHash": sha256.Sum256(key),
		}).Fatal("Failed to decrypt note")
		return nil, err
	}

	return d, nil
}
