package pasteburn

import (
	"bytes"
	"testing"
)

func TestEncryptionOnCreate(t *testing.T) {
	plaintext := []byte("secret")
	key := []byte("11112222333344445555666677778888")

	d, err := MakeDocumentRandomID(plaintext, key)
	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(plaintext, d.Contents) == 0 {
		t.Error("Document wasn't encrypted on creation")
	}

}

// Two documents encrypted with the same key should have different ciphertext.
func TestEncryptionNonce(t *testing.T) {
	plaintext := []byte("secret")
	key := []byte("11112222333344445555666677778888")
	d1, err := MakeDocumentRandomID(plaintext, key)
	d2, err := MakeDocumentRandomID(plaintext, key)
	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(d1.Contents, d2.Contents) == 0 {
		t.Error("Two encrypted documents have the same ciphertext")
	}
}

// Two documents with the same plaintext encrypted with different keys
// should decrypt to have the same plaintext Contents.
func TestDecryptionConsistency(t *testing.T) {
	plaintext := []byte("secret")
	key1 := []byte("11112222333344445555666677778888")
	key2 := []byte("AAAABBBBCCCCDDDDEEEEFFFFGGGGHHHH")

	d1, err := MakeDocumentRandomID(plaintext, key1)
	d2, err := MakeDocumentRandomID(plaintext, key2)
	if err != nil {
		t.Error(err)
	}

	d1.DecryptInPlace(key1)
	d2.DecryptInPlace(key2)

	if bytes.Compare(d1.Contents, d2.Contents) != 0 {
		t.Error("Document was not properly decrypted")
	}
}
