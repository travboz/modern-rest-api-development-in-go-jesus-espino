package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/mattn/go-sqlite3"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(dsn string) (*Repository, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	return &Repository{db: db}, nil
}

// Queries used in Init method
const (
	sqlCreateUsersTable = `
		CREATE TABLE IF NOT EXISTS users (
			role VARCHAR,
			username VARCHAR PRIMARY KEY,
			password VARCHAR
		);
	`

	sqlCreateSessionsTable = `
		CREATE TABLE IF NOT EXISTS sessions (
			token VARCHAR PRIMARY KEY,
			expires TIMESTAMP,
			username VARCHAR
		);
	`

	sqlCreateShoppingListTable = `
		CREATE TABLE IF NOT EXISTS shopping_lists (
			id VARCHAR PRIMARY KEY,
			name VARCHAR,
			items TEXT
		);
	`
)

func (r *Repository) Init() error {
	tables := []string{sqlCreateUsersTable, sqlCreateSessionsTable, sqlCreateShoppingListTable}

	for _, createTableQuery := range tables {
		if _, err := r.db.Exec(createTableQuery); err != nil {
			return err
		}
	}

	return nil
}

const (
	EXPIRY_TIME_IN_DAYS = 2
)

func (r *Repository) AddSession(username string) (*Session, error) {
	token, err := GenerateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("error generating token: %w", err)
	}

	session := Session{Token: token, Expires: time.Now().Add(EXPIRY_TIME_IN_DAYS * 24 * time.Hour), Username: username}

	args := []any{session.Token, session.Expires, session.Username}

	query := sq.Insert("sessions").Columns("token", "expires", "username").Values(args...)
	_, err = query.RunWith(r.db).Exec()
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *Repository) GetSession(token string) (*Session, error) {
	query := sq.
		Select("token", "expires, username").
		From("sessions").
		Where(
			sq.Eq{"token": token},
			sq.Gt{"expires": time.Now()},
		)

	row := query.RunWith(r.db).QueryRow()
	var session Session
	if err := row.Scan(&session.Token, &session.Expires, &session.Username); err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *Repository) GetUserRoleFromSession(token string) (string, error) {
	query := sq.
		Select("token", "expires, username").
		From("sessions").
		Where(
			sq.Eq{"token": token},
			sq.Gt{"expires": time.Now()},
		)

	row := query.RunWith(r.db).QueryRow()
	var session Session
	if err := row.Scan(&session.Token, &session.Expires, &session.Username); err != nil {
		return "", err
	}

	// Fetch the user using the username in the session
	query = sq.Select("role, username, password").From("users").Where(sq.Eq{"username": session.Username})
	var user User
	row = query.RunWith(r.db).QueryRow()
	if err := row.Scan(&user.Role, &user.Username, &user.Password); err != nil {
		return "", err
	}

	return user.Role, nil
}

func (r *Repository) PatchShoppingList(id string, patch *ShoppingListPatch) error {
	query := sq.Update("shopping_lists").Where(sq.Eq{"id": id})

	if patch.Name != nil {
		query = query.Set("name", *patch.Name)
	}
	if patch.Items != nil {
		query = query.Set("items", strings.Join(patch.Items, ","))
	}

	_, err := query.RunWith(r.db).Exec()
	if err != nil {
		return err
	}

	return nil

}
