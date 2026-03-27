package prototype

type Document struct {
	Title   string
	Content string
	Author  string
}

func (d *Document) Clone() Prototype {
	return &Document{
		Title:   d.Title,
		Content: d.Content,
		Author:  d.Author,
	}
}
