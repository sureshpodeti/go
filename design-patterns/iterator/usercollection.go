package iterator

type User struct {
	Name string
	Age  int
}

type Iterator interface {
	Next() *User
	HasNext() bool
}

type UserCollection struct {
	Users []*User
}

func (u *UserCollection) CreateIterator() Iterator {
	return &UserIterator{Users: u.Users, Index: 0}
}
