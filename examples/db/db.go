package db

func GetUsers() []User {
	return []User{
		{
			Name: "User One",
		},
		{
			Name: "User Two",
		},
	}
}

type User struct {
	Name string
}

func (u User) GetName() string {
	return u.Name
}
