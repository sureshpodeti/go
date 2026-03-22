package main

import "designpatterns/iterator"

func main() {
	user1 := &iterator.User{Name: "Tom", Age: 20}
	user2 := &iterator.User{Name: "Jack", Age: 22}

	userCollection := &iterator.UserCollection{Users: []*iterator.User{user1, user2}}

	iterator := userCollection.CreateIterator()

	for iterator.HasNext() {
		user := iterator.Next()
		println(user.Name, user.Age)
	}
}
