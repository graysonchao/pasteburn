package pasteburn

import (
	"log"

	"github.com/boltdb/bolt"
	"github.com/nu7hatch/gouuid"
)

// A Note containing arbitrary text.
type Note struct {
	Body []byte
}

// Save a Note, assigning it a UUID and returning that UUID.
func (n *Note) Save() (string, error) {
	uuid, err := uuid.NewV4()
	name := uuid.String()

	db, err := bolt.Open("pasteburn.db", 0600, nil)
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Notes"))
		err = b.Put([]byte(name), n.Body)
		return err
	})

	if err != nil {
		log.Fatal(err)
	}

	return name, nil
}

func LoadNote(name string) (*Note, error) {

	db, err := bolt.Open("pasteburn.db", 0600, nil)
	defer db.Close()

	var body []byte

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Notes"))

		value := b.Get([]byte(name))
		body = make([]byte, len(value))
		copy(body, value)

		return b.Delete([]byte(name))
	})

	if err != nil {
		return nil, err
	}

	return &Note{Body: body}, nil
}
