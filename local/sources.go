package local

import "github.com/moisespsena-go/assetfs/assetfsapi"

type Sources struct {
	ByName map[string]assetfsapi.LocalSource
}

func NewSources(byName ...map[string]assetfsapi.LocalSource) *Sources {
	s := &Sources{ByName: map[string]assetfsapi.LocalSource{}}
	for _, m := range byName {
		for name, src := range m {
			s.ByName[name] = src
		}
	}
	return s
}

func (m *Sources) Get(name string) assetfsapi.LocalSource {
	if m.ByName != nil {
		return m.ByName[name]
	}
	return nil
}

func (m *Sources) Register(name string, src assetfsapi.LocalSource) {
	if m.ByName == nil {
		m.ByName = map[string]assetfsapi.LocalSource{}
	}
	m.ByName[name] = src
}

type LocalSourcesAttribute struct {
	localSources assetfsapi.LocalSourceRegister
}

func (a *LocalSourcesAttribute) LocalSources() assetfsapi.LocalSourceRegister {
	return a.localSources
}

func (a *LocalSourcesAttribute) SetLocalSources(localSources assetfsapi.LocalSourceRegister) {
	a.localSources = localSources
}
