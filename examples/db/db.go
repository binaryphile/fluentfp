package db

func Users() []User {
	return []User{
		{
			name: "User One",
		},
		{
			name: "User Two",
		},
	}
}

type User struct {
	name string
}

func (u User) Name() string {
	return u.name
}
