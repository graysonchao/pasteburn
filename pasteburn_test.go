package pasteburn

import "testing"

func TestDeletion(t *testing.T) {
	p1 := &Note{Body: []byte("helloworld")}
	uuid, err := p1.save()
	if err != nil {
		panic(err)
	}

	_, err = loadNote(uuid)
	if err != nil {
		t.Error(err)
	}

	_, err = loadNote(uuid)
	if err == nil {
		t.Error("Shouldn't be able to read note again!")
	}
}
