package main

// Session store, maps a Token string to a Session
var sessions = make(map[string]*Session)

// User store with default users, with their:
// - role
// - username
// - password
var allUsers = map[string]*User{
	"admin": {"admin", "admin", "password"},
	"user":  {"user", "user", "password"},
}
