package pasteburn

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/nu7hatch/gouuid"
)

// A Note containing arbitrary text.
type Note struct {
	ID   string `json:"id"`
	Body string `json:"body"`
}

// Save a Note, assigning it a UUID and returning that UUID.
func (n *Note) Save(key []byte) (*Note, error) {
	if len(n.ID) == 0 {
		uuid, err := uuid.NewV4()
		if err != nil {
			return nil, err
		}
		n.ID = uuid.String()
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Input to AES cipher must be padded to a multiple of aes.BlockSize
	plaintext := make([]byte, (len(n.Body)%aes.BlockSize)+aes.BlockSize)
	copy(plaintext, n.Body)

	// Add an additional block worth of data to insert the IV at the beginning
	ciphertext := make([]byte, (len(n.Body)%aes.BlockSize)+(2*aes.BlockSize))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	if err := saveToDb([]byte(n.ID), ciphertext); err != nil {
		return nil, err
	}

	return n, nil
}

func saveToDb(key []byte, value []byte) error {
	db, err := bolt.Open("pasteburn.db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	if err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Notes"))
		err = b.Put(key, value)
		return err
	}); err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

// LoadNote returns the note with the given id, decrypted using the given key
func LoadNote(id []byte, key []byte) (*Note, error) {

	db, err := bolt.Open("pasteburn.db", 0600, nil)
	defer db.Close()

	var body []byte

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Notes"))

		ciphertext := b.Get([]byte(id))

		block, err := aes.NewCipher(key)
		iv := ciphertext[:aes.BlockSize]

		mode := cipher.NewCBCDecrypter(block, iv)
		mode.CryptBlocks(ciphertext, ciphertext)

		if err != nil {
			return err
		}

		body = make([]byte, len(ciphertext))
		copy(body, ciphertext)

		return b.Delete([]byte(id))
	})

	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"id":   id,
		"body": body,
	}).Debug("Loaded encrypted note")

	return &Note{ID: string(id), Body: string(body)}, nil
}
