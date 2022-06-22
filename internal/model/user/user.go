package user

type User struct {
	Name   string
	ApiKey string
}

//var userInstance User

func New() *User {
	return &User{
		Name:   "User",
		ApiKey: "d4cb5a9843b040e8b2e2b7d85794c18d",
	}
}
