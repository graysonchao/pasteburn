package pasteburn

import (
	"sync"

	uuid "github.com/nu7hatch/gouuid"
)

type MockDBService struct {
	mem map[string][]byte
	mtx sync.RWMutex
}

func (m *MockDBService) SaveDocument(d *Document) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	key := string(d.ID.String())
	m.mem[key] = d.Contents
	return nil
}

func (m *MockDBService) LoadDocument(id uuid.UUID) (*Document, error) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	key := id.String()
	contents := m.mem[key]
	d := &Document{
		ID:       id,
		Contents: contents,
	}
	return d, nil
}
