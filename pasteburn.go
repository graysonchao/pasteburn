package pasteburn

import (
	"github.com/nu7hatch/gouuid"
	"io/ioutil"
	"os"
)

type Note struct {
	Body []byte
}

func (n *Note) save() (string, error) {
	uuid, err := uuid.NewV4()
	filename := uuid.String()
	if err != nil {
		return filename, err
	}

	err = ioutil.WriteFile(filename, n.Body, 0600)
	if err != nil {
		return filename, err
	}

	return filename, nil
}

func loadNote(filename string) (*Note, error) {
	body, err := ioutil.ReadFile(filename)
	os.Remove(filename)
	if err != nil {
		return nil, err
	}
	return &Note{Body: body}, nil
}
