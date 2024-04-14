package core

type User struct {
	Username string `bson:"username"`
	IP       string `bson:"ip"`
}
