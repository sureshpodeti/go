package prototype

type TemplatedDocument struct {
	Document
	TemplateName string
	Version      int
}

func (t *TemplatedDocument) Clone() Prototype {
	return &TemplatedDocument{
		Document: Document{
			Title:   t.Title,
			Content: t.Content,
			Author:  t.Author,
		},
		TemplateName: t.TemplateName,
		Version:      t.Version,
	}
}
