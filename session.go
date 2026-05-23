package main

import (
	"crypto/rand"
	"encoding/hex"
)

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

// GenerateSessionToken creates a cryptographically secure, random 256-bit session token
// encoded as a 64-character hexadecimal string.
func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
