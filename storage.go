package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccounts() ([]*Account, error)
	GetAccountById(int) (*Account, error)
}

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage() (*PostgresStorage, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	connString := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", connString)

	if err != nil {
		return nil, err
	}

	// Test The Connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStorage{
		db: db,
	}, nil
}

func (s *PostgresStorage) Init() {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS accounts (
		id SERIAL PRIMARY KEY,
		first_name TEXT,
		last_name TEXT,
		number BIGINT,
		balance BIGINT,
		create_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)

	if err != nil {
		log.Fatalf("Error Creating Table: %s", err)
	}
}

func (s *PostgresStorage) GetAccounts() ([]*Account, error) {
	rows, err := s.db.Query(`SELECT * FROM accounts`)

	if err != nil {
		return nil, err
	}

	accounts := []*Account{}
	for rows.Next() {
		account, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (s *PostgresStorage) CreateAccount(account *Account) error {
	_, err := s.db.Exec(`INSERT INTO accounts (
	first_name,
	last_name,
	number,
	balance
	) VALUES ($1, $2, $3, $4)`, account.FirstName, account.LastName, account.Number, account.Balance)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) DeleteAccount(id int) error {
	_, err := s.db.Query(`DELETE FROM accounts WHERE id = $1`, id)

	if err != nil {
		return fmt.Errorf("account %d not found", id)
	}

	return nil
}

func (s *PostgresStorage) UpdateAccount(account *Account) error {
	return nil
}

func (s *PostgresStorage) GetAccountById(id int) (*Account, error) {
	rows, err := s.db.Query(`SELECT * FROM accounts WHERE id = $1`, id)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("account %d not found", id)
}

func scanIntoAccount(row *sql.Rows) (*Account, error) {
	account := &Account{}
	if err := row.Scan(&account.ID, &account.FirstName, &account.LastName, &account.Number, &account.Balance, &account.CreatedAt); err != nil {
		return nil, err
	}
	return account, nil

}
