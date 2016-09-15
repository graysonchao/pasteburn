package pasteburn

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/nu7hatch/gouuid"
)

// Service is a CRUD interface for Pasteburn documents.
type Service interface {
	PostDocument(ctx context.Context, d Document) error
	GetDocument(ctx context.Context, d Document) (Document, error)
}

// A Document contains arbitrary data.
type Document struct {
	ID        uuid.UUID `json:"id"`
	Contents  []byte    `json:"body"`
	Encrypted bool
}

// AES256KeySizeBytes is the appropriate size for an AES256 encryption key
// See https://golang.org/pkg/crypto/aes/
const AES256KeySizeBytes int = 32

// MarshalJSON customizes how a Document is marshalled in JSON.
func (d *Document) MarshalJSON() ([]byte, error) {
	if d.Encrypted {
		return json.Marshal(&struct {
			ID   string
			Body []byte
		}{
			ID:   d.ID.String(),
			Body: d.Contents,
		})
	}
	return json.Marshal(&struct {
		ID   string
		Body string
	}{
		ID:   d.ID.String(),
		Body: string(d.Contents),
	})
}

// MakeDocument returns a *Document whose Body is the given body encrypted with the given key.
// The ID of the note is a UUID generated at the time of creation.
func MakeDocument(body []byte, key []byte) (*Document, error) {

	if len(key) != AES256KeySizeBytes {
		err := errors.New("Tried to make a note with AES256 key of the wrong length")
		log.WithFields(log.Fields{
			"keyLength": len(key),
			"keyHash":   sha256.Sum256(key),
			"key":       string(key),
		}).Fatal(err)
		return nil, err
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	// Pad body to a multiple of aes.BlockSize length for encryption.
	paddingLength := aes.BlockSize - (len(body) % aes.BlockSize)
	if len(body)%aes.BlockSize == 0 {
		// Padding must be added because we rely the value of padding bytes
		// to know how much to chop off after we're done decrypting (RFC 5246 6.2.3.2)
		paddingLength = aes.BlockSize
	}
	padding := make([]byte, paddingLength)

	// We can afford to cast paddingLength to a byte because we know the AES block size is small.
	// If it were >127 then we'd end up with nonsense padding data and not know how much to chop off.
	for i := range padding {
		padding[i] = byte(paddingLength)
	}

	body = append(body, padding...)

	d := &Document{
		ID:       *uuid,
		Contents: body,
	}

	err = d.encryptInPlace(key)
	if err != nil {
		log.WithFields(log.Fields{
			"note":       d,
			"key_sha256": sha256.Sum256(key),
		}).Fatal("Failed to encrypt note")
	}

	return d, nil
}

// encryptInPlace returns an error if the note could not be encrypted.
// *** WARNING! ***
// This directly changes the values of its own fields - in other words,
// this is a destructive/irreversible operation.
func (d *Document) encryptInPlace(key []byte) error {
	if len(d.Contents)%aes.BlockSize != 0 {
		err := errors.New("Tried to encrypt a plaintext where length % aes.BlockSize != 0")
		log.WithFields(log.Fields{
			"id":     d.ID,
			"body":   d.Contents,
			"length": len(d.Contents),
		}).Fatal("encryptInPlace", err)
		return err
	}

	// Add an extra <blocksize> bytes to prepend IV
	d.Contents = append(make([]byte, aes.BlockSize), d.Contents...)

	// Fill iv block with random data (iv must be unique but not secure)
	// (The iv block will not be encrypted)
	iv := d.Contents[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.WithFields(log.Fields{
			"iv": iv,
		}).Fatal("Failed to fill iv with random data")
		return err
	}

	cb, err := aes.NewCipher(key)
	if err != nil {
		log.WithFields(log.Fields{
			"key_sha256": sha256.Sum256(key),
		}).Fatal("Creating new cipher", err)
		return err
	}

	mode := cipher.NewCBCEncrypter(cb, iv)
	mode.CryptBlocks(d.Contents[aes.BlockSize:], d.Contents[aes.BlockSize:])

	d.Encrypted = true

	return nil
}

func (d *Document) decryptInPlace(key []byte) error {
	cb, err := aes.NewCipher(key)
	if err != nil {
		log.WithFields(log.Fields{
			"key_sha256": sha256.Sum256(key),
			"key":        string(key),
		}).Fatal("Failed to create cipher")
		return err
	}

	// Cut off leading iv
	iv := d.Contents[:aes.BlockSize]
	ciphertext := d.Contents[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(cb, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// Cut off padding. Go has no way to shrink a slice, so we have to make a new one.
	paddingLength := int(ciphertext[len(ciphertext)-1])
	cutBody := make([]byte, len(ciphertext)-paddingLength)
	copy(cutBody, ciphertext[:len(ciphertext)-paddingLength])
	d.Contents = cutBody

	return nil
}

// Save a Document, assigning it a UUID and returning that UUID.
func (d *Document) Save() error {

	if err := saveToDb(d.ID, d.Contents); err != nil {
		log.WithFields(log.Fields{
			"note": d,
		}).Fatal("Failed to save note")
		return err
	}

	return nil
}


// LoadDocument returns the note with the given id, encrypted
func LoadDocument(id uuid.UUID, key []byte) (*Document, error) {

	body, err := loadAndDeleteFromDb(id)
	if err != nil {
		log.WithFields(log.Fields{
			"id": "id",
		}).Fatal("Failed to load note")
		return nil, err
	}

	log.WithFields(log.Fields{
		"id":   id,
		"body": body,
	}).Debug("Loaded encrypted note")

	d := &Document{
		ID:       id,
		Contents: body,
	}

	if len(d.Contents) == 0 {
		// TODO Document has been deleted how to handle this better?
		return d, nil
	}

	if err := d.decryptInPlace(key); err != nil {
		log.WithFields(log.Fields{
			"id":      id,
			"keyHash": sha256.Sum256(key),
		}).Fatal("Failed to decrypt note")
		return nil, err
	}

	d.Encrypted = false

	return d, nil
}

