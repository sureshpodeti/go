package prototype

type TemplatedDocument struct {
	Document
	TemplateName string
	Version      int
}

func (td *TemplatedDocument) Clone() Prototype {
	return &TemplatedDocument{
		Document: Document{
			Title:    td.Title,
			Author:   td.Author,
			Content:  td.Content,
			FontSize: td.FontSize,
		},
		TemplateName: td.TemplateName,
		Version:      td.Version,
	}
}
