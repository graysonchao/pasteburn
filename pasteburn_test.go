package pasteburn

import "testing"

func TestDeletionAfterRead(t *testing.T) {
	p1 := &Note{Body: []byte("helloworld")}
	n, err := p1.Save()
	if err != nil {
		panic(err)
	}

	n2, err := LoadNote(n.ID)
	if err != nil {
		t.Error(err)
	}
	if n2.Body == nil {
		t.Error("Couldn't read body the first time")
	}

	n3, err := LoadNote(n.ID)
	if len(n3.Body) > 0 {
		t.Error("Shouldn't be able to read note again!")
	}
}
