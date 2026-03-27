package registry

import "designpatterns/prototype1/prototype"

type DocumentRegistry struct {
	Templates map[string]prototype.Prototype
}

func NewDocumentRegistry() *DocumentRegistry {
	return &DocumentRegistry{
		Templates: make(map[string]prototype.Prototype),
	}
}

func (d *DocumentRegistry) Register(name string, proto prototype.Prototype) {
	d.Templates[name] = proto
}
func (d *DocumentRegistry) Get(docType string) prototype.Prototype {
	if temp, ok := d.Templates[docType]; ok {
		return temp.Clone()
	}
	return nil
}
