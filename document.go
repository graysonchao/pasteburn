package pasteburn

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"

	log "github.com/Sirupsen/logrus"
	uuid "github.com/nu7hatch/gouuid"
)

// A Document contains arbitrary data.
// Encrypted may be nil if it's not known whether the data is encrypted.
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

// MakeDocumentRandomID makes a document with a random ID.
func MakeDocumentRandomID(body []byte, key []byte) (*Document, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	return MakeDocument(id, body, key)
}

// MakeDocument returns a *Document whose Body is the given body encrypted with the given key.
func MakeDocument(id *uuid.UUID, body []byte, key []byte) (*Document, error) {

	if len(key) != AES256KeySizeBytes {
		err := errors.New("Tried to make a note with AES256 key of the wrong length")
		log.WithFields(log.Fields{
			"keyLength": len(key),
			"keyHash":   sha256.Sum256(key),
			"key":       string(key),
		}).Fatal(err)
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
		ID:       *id,
		Contents: body,
	}

	if err := d.EncryptInPlace(key); err != nil {
		return d, err
	}

	return d, nil
}

// EncryptInPlace returns an error if the note could not be encrypted.
// It encrypts d.Contents using AES256 with the given key.
func (d *Document) EncryptInPlace(key []byte) error {

	if d.Encrypted {
		err := errors.New("Tried to encrypt an already-encrypted document.")
		log.WithFields(log.Fields{
			"id": d.ID,
		}).Fatal("Failed encrypting document:", err)
	}

	if len(d.Contents)%aes.BlockSize != 0 {
		err := errors.New("Tried to encrypt a plaintext where length % aes.BlockSize != 0")
		log.WithFields(log.Fields{
			"id":     d.ID,
			"body":   d.Contents,
			"length": len(d.Contents),
		}).Fatal("Failed encrypting document:", err)
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
		}).Fatal("Failed encrypting document:", err)
		return err
	}

	cb, err := aes.NewCipher(key)
	if err != nil {
		log.WithFields(log.Fields{
			"key_sha256": sha256.Sum256(key),
		}).Fatal("Failed encrypting document:", err)
		return err
	}

	mode := cipher.NewCBCEncrypter(cb, iv)
	mode.CryptBlocks(d.Contents[aes.BlockSize:], d.Contents[aes.BlockSize:])

	d.Encrypted = true

	return nil
}

// DecryptInPlace returns an error if the note could not be encrypted.
// It decrypts d.Contents using AES256 with the given key.
func (d *Document) DecryptInPlace(key []byte) error {
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
