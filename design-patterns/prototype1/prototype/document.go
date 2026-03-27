package prototype

type Document struct {
	Title    string
	Author   string
	Content  string
	FontSize int
}

func (d *Document) Clone() Prototype {
	return &Document{
		Title:    d.Title,
		Author:   d.Author,
		Content:  d.Content,
		FontSize: d.FontSize,
	}
}
