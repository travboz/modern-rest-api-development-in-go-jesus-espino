package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/mattn/go-sqlite3"
)

type RepositoryInterface interface {
	Init() error
	Empty() error
	AddSession(username string) (*Session, error)
	GetSession(token string) (*Session, error)
	GetUserRoleFromSession(token string) (string, error)
	CreateNewShoppingList(list *ShoppingList) error
	GetAllShoppingLists() ([]*ShoppingList, error)
	GetListByID(id int) (*ShoppingList, error)
	DeleteShoppingList(id int) error
	PatchShoppingList(id int, patch *ShoppingListPatchRequest) error
	UpdateShoppingList(update *ShoppingList) error
}

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
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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

func (r *Repository) Empty() error {
	query := sq.Delete("shopping_lists")
	_, err := query.RunWith(r.db).Exec()
	if err != nil {
		return err
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

// sqlCreateShoppingListTable = `
// 		CREATE TABLE IF NOT EXISTS shopping_lists (
// 			id VARCHAR PRIMARY KEY,
// 			name VARCHAR,
// 			items TEXT
// 		);
// 	`

func (r *Repository) CreateNewShoppingList(list *ShoppingList) error {
	args := []any{list.Name, strings.Join(list.Items, ",")}
	query := sq.Insert("shopping_lists").Columns("name", "items").Values(args...)
	// TODO: Need to return ID of new shopping list
	_, err := query.RunWith(r.db).Exec()
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetAllShoppingLists() ([]*ShoppingList, error) {
	query := sq.Select("id", "name", "items").From("shopping_lists")

	rows, err := query.RunWith(r.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	shoppingLists := make([]*ShoppingList, 0)

	for rows.Next() {
		var list ShoppingList

		var items string

		if err := rows.Scan(&list.ID, &list.Name, &items); err != nil {
			return nil, err
		}

		list.Items = strings.Split(items, ",")

		shoppingLists = append(shoppingLists, &list)
	}

	return shoppingLists, nil
}

var (
	ErrRecordNotFound = errors.New("Record not found")
)

func (r *Repository) GetListByID(id int) (*ShoppingList, error) {
	query := sq.Select("id", "name", "items").From("shopping_lists").Where(sq.Eq{"id": id})

	row := query.RunWith(r.db).QueryRow()

	var shopList ShoppingList
	var listItems string

	if err := row.Scan(&shopList.ID, &shopList.Name, &listItems); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}

		return nil, err
	}

	shopList.Items = strings.Split(listItems, ",")

	return &shopList, nil
}

func (r *Repository) DeleteShoppingList(id int) error {
	query := sq.Delete("shopping_lists").Where(sq.Eq{"id": id})

	result, err := query.RunWith(r.db).Exec()
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (r *Repository) PatchShoppingList(id int, patch *ShoppingListPatchRequest) error {
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

func (r *Repository) UpdateShoppingList(update *ShoppingList) error {
	query := sq.Update("shoppings_lists").Where(sq.Eq{"id": update.ID}).Set("name", update.Name).Set("items", strings.Join(update.Items, ","))

	result, err := query.RunWith(r.db).Exec()
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
