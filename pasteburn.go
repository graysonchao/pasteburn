package pasteburn

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/nu7hatch/gouuid"
)

// A Note contains arbitrary data.
// It does not know how ot
type Note struct {
	ID   uuid.UUID `json:"id"`
	Body []byte    `json:"body"`
}

// MakeNote returns a *Note whose Body is the given body encrypted with the given key.
// The ID of the note is a UUID generated at the time of creation.
func MakeNote(body []byte, key []byte) (*Note, error) {

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

	n := &Note{
		ID:   *uuid,
		Body: body,
	}

	err = n.encryptInPlace(key)
	if err != nil {
		log.WithFields(log.Fields{
			"note":       n,
			"key_sha256": sha256.Sum256(key),
		}).Fatal("Failed to encrypt note")
	}

	return n, nil
}

// encryptInPlace returns an error if the note could not be encrypted.
// *** WARNING! ***
// This directly changes the values of its own fields - in other words,
// this is a destructive/irreversible operation.
func (n *Note) encryptInPlace(key []byte) error {
	if len(n.Body)%aes.BlockSize != 0 {
		err := errors.New("Tried to encrypt a plaintext where length % aes.BlockSize != 0")
		log.WithFields(log.Fields{
			"id":     n.ID,
			"body":   n.Body,
			"length": len(n.Body),
		}).Fatal("encryptInPlace", err)
		return err
	}

	// Add an extra <blocksize> bytes to prepend IV
	n.Body = append(make([]byte, aes.BlockSize), n.Body...)

	// Fill iv block with random data (iv must be unique but not secure)
	// (The iv block will not be encrypted)
	iv := n.Body[:aes.BlockSize]
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
	mode.CryptBlocks(n.Body[aes.BlockSize:], n.Body[aes.BlockSize:])
	return nil
}

func (n *Note) decryptInPlace(key []byte) error {
	cb, err := aes.NewCipher(key)
	if err != nil {
		log.WithFields(log.Fields{
			"key_sha256": sha256.Sum256(key),
		}).Fatal("Failed to create cipher")
		return err
	}

	// Cut off leading iv
	iv := n.Body[:aes.BlockSize]
	ciphertext := n.Body[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(cb, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// Cut off padding. Go has no way to shrink a slice, so we have to make a new one.
	paddingLength := int(ciphertext[len(ciphertext)-1])
	cutBody := make([]byte, len(ciphertext)-paddingLength)
	copy(cutBody, ciphertext[:len(ciphertext)-paddingLength])
	n.Body = cutBody

	return nil
}

// Save a Note, assigning it a UUID and returning that UUID.
func (n *Note) Save() error {

	if err := saveToDb(n.ID, n.Body); err != nil {
		log.WithFields(log.Fields{
			"note": n,
		}).Fatal("Failed to save note")
		return err
	}

	return nil
}

func saveToDb(uuid uuid.UUID, value []byte) error {
	db, err := bolt.Open("pasteburn.db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	key := make([]byte, len(uuid))
	copy(key, uuid[:])

	if err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Notes"))
		err = b.Put(key, value)
		return err
	}); err != nil {
		log.WithField("function", "saveToDb").Fatal(err)
		return err
	}

	return nil
}

// LoadNote returns the note with the given id, encrypted
func LoadNote(id uuid.UUID, key []byte) (*Note, error) {

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

	n := &Note{
		ID:   id,
		Body: body,
	}

	if len(n.Body) == 0 {
		// TODO Note has been deleted how to handle this better?
		return n, nil
	}

	if err := n.decryptInPlace(key); err != nil {
		log.WithFields(log.Fields{
			"id":      id,
			"keyHash": sha256.Sum256(key),
		}).Fatal("Failed to decrypt note")
		return nil, err
	}

	return n, nil
}

func loadAndDeleteFromDb(key uuid.UUID) ([]byte, error) {
	db, err := bolt.Open("pasteburn.db", 0600, nil)
	defer db.Close()

	var value []byte

	if err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Notes"))

		l := b.Get(key[:])
		value = make([]byte, len(l))
		copy(value, l)

		return b.Delete(key[:])
	}); err != nil {
		return nil, err
	}

	return value, nil
}
