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
		_, err := tx.CreateBucketIfNotExists([]byte("Documents"))
		return err
	})
	db.Close()

	flag.Parse()
	os.Exit(m.Run())
}

func TestDeletionAfterRead(t *testing.T) {
	plaintext := []byte("secret")
	key := []byte("11112222333344445555666677778888")
	d1, err := MakeDocument(plaintext, key)
	if err := d1.Save(); err != nil {
		t.Error(err)
	}

	d2, err := LoadDocument(d1.ID, key)
	if err != nil {
		t.Error(err)
	}
	if d2.Contents == nil {
		t.Error("Couldn't read body the first time")
	}

	d3, err := LoadDocument(d1.ID, key)
	if len(d3.Contents) > 0 {
		t.Error("Shouldn't be able to read note again!")
	}

}

func TestEncryptedOnDisk(t *testing.T) {
	plaintext := []byte("secret")
	key := []byte("11112222333344445555666677778888")
	d1, err := MakeDocument(plaintext, key)
	if err != nil {
		t.Error(err)
	}

	if err = d1.Save(); err != nil {
		t.Error(err)
	}

	id := d1.ID

	ciphertext, err := loadAndDeleteFromDb(id)
	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(ciphertext, plaintext) == 0 {
		t.Error("Document was not encrypted")
	}
}

func TestDecryption(t *testing.T) {
	plaintext := []byte("secret")
	key := []byte("11112222333344445555666677778888")
	d1, err := MakeDocument(plaintext, key)
	if err != nil {
		t.Error(err)
	}

	if err = d1.Save(); err != nil {
		t.Error(err)
	}

	id := d1.ID

	d2, err := LoadDocument(id, key)

	plaintext2 := d2.Contents

	if bytes.Compare(plaintext, plaintext2) != 0 {
		t.Error("Document was not properly decrypted")
	}
}
