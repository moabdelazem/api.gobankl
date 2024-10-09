package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Storage interface
// Represents a storage interface that defines the methods for interacting with the database.
type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccounts() ([]*Account, error)
	GetAccountById(int) (*Account, error)
}

// PostgresStorage struct
// Represents a PostgreSQL storage implementation.
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage initializes a new PostgresStorage instance by loading
// environment variables from a .env file and establishing a connection to
// the PostgreSQL database using the connection string specified in the DB_URL
// environment variable. It returns a pointer to the PostgresStorage instance
// and an error if any occurs during the process.
//
// Returns:
//   - *PostgresStorage: A pointer to the initialized PostgresStorage instance.
//   - error: An error if there is an issue loading the .env file, opening the
//     database connection, or pinging the database.
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

// Init initializes the PostgresStorage by creating the accounts table if it does not already exist.
// The table includes columns for id, first_name, last_name, number, balance, and create_at.
// If there is an error during table creation, the function logs a fatal error.
func (s *PostgresStorage) Init() {
	// Create The Table
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

// GetAccounts retrieves all accounts from the Postgres database.
// It executes a SQL query to select all records from the 'accounts' table,
// scans each row into an Account struct, and returns a slice of Account pointers.
// If an error occurs during the query or scanning process, it returns the error.
//
// Returns:
//   - []*Account: A slice of pointers to Account structs representing the accounts.
//   - error: An error object if an error occurs, otherwise nil.
func (s *PostgresStorage) GetAccounts() ([]*Account, error) {
	// Query The Database
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

// CreateAccount inserts a new account record into the accounts table in the database.
// It takes an Account struct as input and returns an error if the insertion fails.
//
// Parameters:
//   - account: A pointer to an Account struct containing the account details to be inserted.
//
// Returns:
//   - error: An error object if the insertion fails, otherwise nil.
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

// DeleteAccount deletes an account from the database based on the provided account ID.
// It returns an error if the account is not found or if there is an issue with the database query.
//
// Parameters:
//   - id: The ID of the account to be deleted.
//
// Returns:
//   - error: An error object if the account is not found or if there is a database query issue, otherwise nil.
func (s *PostgresStorage) DeleteAccount(id int) error {
	_, err := s.db.Query(`DELETE FROM accounts WHERE id = $1`, id)

	if err != nil {
		return fmt.Errorf("account %d not found", id)
	}

	return nil
}

// TODO: Implement UpdateAccount method
func (s *PostgresStorage) UpdateAccount(account *Account) error {
	return nil
}

// GetAccountById retrieves an account from the database based on the provided account ID.
// It executes a SQL query to select the account with the specified ID, scans the row into an Account struct,
// and returns a pointer to the Account struct if the account is found. If the account is not found, it returns an error.
//
// Parameters:
//   - id: The ID of the account to retrieve.
//
// Returns:
//   - *Account: A pointer to the Account struct representing the account.
//   - error: An error object if the account is not found, otherwise nil.
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

// scanIntoAccount scans the current row of the provided SQL rows object into an Account struct.
// It returns a pointer to the Account struct and an error if the scanning process fails.
//
// Parameters:
//   - row: A pointer to the SQL rows object representing the current row.
//
// Returns:
//   - *Account: A pointer to the Account struct containing the scanned data.
//   - error: An error object if the scanning process fails, otherwise nil.
func scanIntoAccount(row *sql.Rows) (*Account, error) {
	account := &Account{}
	if err := row.Scan(&account.ID, &account.FirstName, &account.LastName, &account.Number, &account.Balance, &account.CreatedAt); err != nil {
		return nil, err
	}
	return account, nil

}
