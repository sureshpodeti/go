package iterator

type UserIterator struct {
	Index int
	Users []*User
}

func (ui *UserIterator) HasNext() bool {
	if ui.Index < len(ui.Users) {
		return true
	}
	return false
}

func (ui *UserIterator) Next() *User {
	user := ui.Users[ui.Index]
	ui.Index++
	return user
}
