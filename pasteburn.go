package pasteburn

import (
	"log"

	"github.com/boltdb/bolt"
	"github.com/nu7hatch/gouuid"
)

// A Note containing arbitrary text.
type Note struct {
	ID   string `json:"id"`
	Body string `json:"body"`
}

// Save a Note, assigning it a UUID and returning that UUID.
func (n *Note) Save() (*Note, error) {
	if len(n.ID) == 0 {
		uuid, err := uuid.NewV4()
		if err != nil {
			return nil, err
		}
		n.ID = uuid.String()
	}

	db, err := bolt.Open("pasteburn.db", 0600, nil)
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Notes"))
		err = b.Put([]byte(n.ID), []byte(n.Body))
		return err
	})

	if err != nil {
		log.Fatal(err)
	}

	return n, nil
}

func LoadNote(id []byte) (*Note, error) {

	db, err := bolt.Open("pasteburn.db", 0600, nil)
	defer db.Close()

	var body []byte

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Notes"))

		value := b.Get([]byte(id))
		body = make([]byte, len(value))
		copy(body, value)

		return b.Delete([]byte(id))
	})

	if err != nil {
		return nil, err
	}

	return &Note{ID: string(id), Body: string(body)}, nil
}
