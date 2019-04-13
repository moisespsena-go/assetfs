package local

import (
	"sync"
)

type Sources struct {
	ByName map[string]*SourceDir

	mu sync.RWMutex
}

func (m *Sources) Get(name string) *SourceDir {
	if m.ByName != nil {
		return m.ByName[name]
	}
	return nil
}

func (m *Sources) Register(name string, dir string) (assets *SourceDir) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ByName == nil {
		m.ByName = map[string]*SourceDir{}
	}
	assets = &SourceDir{Dir: dir}
	m.ByName[name] = assets
	return assets
}
