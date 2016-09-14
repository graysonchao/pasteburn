package pasteburn

import (
	"bytes"
	"flag"
	"os"
	"testing"

	"github.com/boltdb/bolt"
)

func TestMain(m *testing.M) {
	db, _ := bolt.Open("pasteburn.db", 0600, nil)

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Notes"))
		return err
	})
	db.Close()

	flag.Parse()
	os.Exit(m.Run())
}

func TestDeletionAfterRead(t *testing.T) {
	plaintext := []byte("secret")
	key := []byte("11112222333344445555666677778888")
	n1, err := MakeNote(plaintext, key)
	if err := n1.Save(); err != nil {
		t.Error(err)
	}

	n2, err := LoadNote(n1.ID, key)
	if err != nil {
		t.Error(err)
	}
	if n2.Body == nil {
		t.Error("Couldn't read body the first time")
	}

	n3, err := LoadNote(n1.ID, key)
	if len(n3.Body) > 0 {
		t.Error("Shouldn't be able to read note again!")
	}

}

func TestEncryptedOnDisk(t *testing.T) {
	plaintext := []byte("secret")
	key := []byte("11112222333344445555666677778888")
	n1, err := MakeNote(plaintext, key)
	if err != nil {
		t.Error(err)
	}

	if err = n1.Save(); err != nil {
		t.Error(err)
	}

	id := n1.ID

	ciphertext, err := loadAndDeleteFromDb(id)
	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(ciphertext, plaintext) == 0 {
		t.Error("Note was not encrypted")
	}
}

func TestDecryption(t *testing.T) {
	plaintext := []byte("secret")
	key := []byte("11112222333344445555666677778888")
	n1, err := MakeNote(plaintext, key)
	if err != nil {
		t.Error(err)
	}

	if err = n1.Save(); err != nil {
		t.Error(err)
	}

	id := n1.ID

	n2, err := LoadNote(id, key)

	plaintext2 := n2.Body

	if bytes.Compare(plaintext, plaintext2) != 0 {
		t.Error("Note was not properly decrypted")
	}
}
