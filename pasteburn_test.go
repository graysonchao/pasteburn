package pasteburn

import "testing"

func TestDeletionAfterRead(t *testing.T) {
	p1 := &Note{Body: []byte("helloworld")}
	uuid, err := p1.Save()
	if err != nil {
		panic(err)
	}

	body, err := LoadNote(uuid)
	if err != nil {
		t.Error(err)
	}
	if body == nil {
		t.Error("Couldn't read body the first time")
	}

	note, err := LoadNote(uuid)
	if len(note.Body) > 0 {
		t.Error("Shouldn't be able to read note again!")
	}
}
