package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// API Server
type APIServer struct {
	listenAddr string
	store      Storage
}

// Create New API Server
// listenAddr: The Address To Listen On
// store: The Storage Interface
func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

// handleAccount handles HTTP requests for account-related operations.
// It delegates the request to the appropriate handler based on the HTTP method:
// - GET: Calls handleGetAccounts to retrieve account information.
// - POST: Calls handleCreateAccount to create a new account.
// - DELETE: Calls handleDeleteAccount to delete an existing account.
// If the HTTP method is not supported, it returns an error indicating that the method is not allowed.
func (as *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	// Check The Method and Call The Right Handler
	switch r.Method {
	case "GET":
		return as.handleGetAccounts(w, r)
	case "POST":
		return as.handleCreateAccount(w, r)
	case "DELETE":
		return as.handleDeleteAccount(w, r)
	default:
		return fmt.Errorf("Method Is Not Allowed: %s", r.Method)
	}
}

// handleGetAccounts handles the HTTP request to retrieve all accounts.
// It fetches all accounts from the store and writes them as a JSON response.
// If an error occurs while fetching the accounts, it returns the error.
//
// Parameters:
//   - w: The HTTP response writer to send the response.
//   - r: The HTTP request received from the client.
//
// Returns:
//   - error: An error if there is an issue retrieving the accounts or writing the response.
func (as *APIServer) handleGetAccounts(w http.ResponseWriter, r *http.Request) error {
	// Check The Method
	if r.Method != http.MethodGet {
		return fmt.Errorf("Method Not Allowed: %s", r.Method)
	}
	// Get All Accounts From The Store
	accs, err := as.store.GetAccounts()

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accs)
}

// handleCreateAccount handles the creation of a new account.
// It decodes the request body into a CreateAccountRequest, creates a new account,
// stores it in the database, and writes the created account as a JSON response.
//
// Parameters:
//   - w: http.ResponseWriter to write the response.
//   - r: *http.Request containing the request data.
//
// Returns:
//   - error: An error if the request body cannot be decoded, the account cannot be created,
//     or the account cannot be stored in the database.
func (as *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	// Decode The Request Body To CreateAccountRequest
	accReq := new(CreateAccountRequest)

	if err := json.NewDecoder(r.Body).Decode(accReq); err != nil {
		return err
	}
	defer r.Body.Close()

	// Create The Account
	acc := NewAccount(accReq.FirstName, accReq.LastName)

	// Store The Account in The Database
	if err := as.store.CreateAccount(acc); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusCreated, acc)
}

// handleGetAccountById handles the HTTP request to retrieve an account by its ID.
// It extracts the account ID from the URL, fetches the account details from the store,
// and writes the account information as a JSON response.
//
// Parameters:
//   - w: http.ResponseWriter to write the response.
//   - r: *http.Request containing the request details.
//
// Returns:
//   - error: An error if the account retrieval fails, otherwise nil.
func (as *APIServer) handleGetAccountById(w http.ResponseWriter, r *http.Request) error {
	// Get The ID From The URL
	id := getId(w, r)

	// Get The Account By ID
	acc, err := as.store.GetAccountById(id)

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, acc)
}

// handleDeleteAccount handles the HTTP request for deleting an account.
// It extracts the account ID from the URL, deletes the account from the store,
// and writes a JSON response indicating the deleted account ID.
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response.
//   - r: *http.Request containing the HTTP request details.
//
// Returns:
//   - error: An error if the account deletion fails, otherwise nil.
func (as *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	// Get The ID From The URL
	id := getId(w, r)

	// Delete The Account By ID from The Store
	if err := as.store.DeleteAccount(id); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, map[string]int{
		"deleted": id,
	})
}

// handleTransfer handles the transfer request by decoding the JSON payload
// from the request body into a TransferRequest struct. It returns an error
// if the decoding fails. The function writes the transfer request back as a
// JSON response with an HTTP status of 200 OK.
//
// Parameters:
// - w: http.ResponseWriter to write the response.
// - r: *http.Request containing the transfer request.
//
// Returns:
// - error: An error if the JSON decoding fails.
func (as *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transferReq := TransferRequest{}

	if err := json.NewDecoder(r.Body).Decode(&transferReq); err != nil {
		return err
	}
	defer r.Body.Close()

	// TODO: Implement the transfer logic here
	// For now, just return the transfer request as a response

	return WriteJSON(w, http.StatusOK, transferReq)
}

// Run initializes the API server, sets up the router and sub-router with the appropriate
// routes for handling account and transfer operations, and starts the HTTP server.
//
// Routes:
// - POST /api/v1/account: Handles account creation.
// - GET /api/v1/account/{id:[0-9]+}: Retrieves account details by ID.
// - DELETE /api/v1/account/{id:[0-9]+}: Deletes an account by ID.
// - POST /api/v1/transfer: Handles money transfers between accounts.
//
// The server listens on the address specified in the APIServer's listenAddr field.
func (as *APIServer) Run() {
	// Create The Router and SubRouter
	router := mux.NewRouter()
	subRouter := router.PathPrefix("/api/v1").Subrouter()

	// Handle The Accounts Routes
	subRouter.HandleFunc("/account", makeHTTPHandlerFunc(as.handleAccount))
	subRouter.HandleFunc("/account/{id:[0-9]+}", makeHTTPHandlerFunc(as.handleGetAccountById)).Methods(http.MethodGet)
	subRouter.HandleFunc("/account/{id:[0-9]+}", makeHTTPHandlerFunc(as.handleDeleteAccount)).Methods(http.MethodDelete)

	// Handle The Transfer Route
	subRouter.HandleFunc("/transfer", makeHTTPHandlerFunc(as.handleTransfer)).Methods(http.MethodPost)

	// Run The HTTPServer

	log.Println("API Server is Runing On Port: ", as.listenAddr)

	http.ListenAndServe(as.listenAddr, router)
}

func (as *APIServer) GracefulShutdown() {
	// TODO: Implement graceful shutdown
}

// apiFunc is a type definition for a function that takes an http.ResponseWriter
// and an *http.Request as parameters and returns an error. This type is used
// to define the signature of API handler functions in the application.
type apiFunc func(http.ResponseWriter, *http.Request) error

// APIError is a struct that represents an error response in the API.
// It contains an "error" field that holds the error message to be returned
// to the client in the response body.
type APIError struct {
	Error string `json:"error"`
}

// makeHTTPHandlerFunc wraps an apiFunc with an http.HandlerFunc.
// It executes the provided apiFunc and handles any errors by writing
// a JSON response with a status code of http.StatusBadRequest and an
// APIError containing the error message.
//
// Parameters:
//   - f: The apiFunc to be wrapped.
//
// Returns:
//   - An http.HandlerFunc that executes the provided apiFunc and handles errors.
func makeHTTPHandlerFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
		}
	}
}

// WriteJSON writes the given data as a JSON response with the specified HTTP status code.
// It sets the "Content-Type" header to "application/json" and encodes the data into the response writer.
//
// Parameters:
//   - w: The http.ResponseWriter to write the response to.
//   - status: The HTTP status code to set for the response.
//   - data: The data to encode and write as JSON.
//
// Returns:
//   - error: An error if encoding the data fails, otherwise nil.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

// getId extracts the "id" variable from the URL path, converts it to an integer,
// and returns the integer value. If the conversion fails, it writes an error
// message to the ResponseWriter and returns -1.
//
// Parameters:
//   - w: http.ResponseWriter to write the error message if the conversion fails.
//   - r: *http.Request from which the "id" variable is extracted.
//
// Returns:
//   - int: The integer value of the "id" variable, or -1 if the conversion fails.
func getId(w http.ResponseWriter, r *http.Request) int {
	id := mux.Vars(r)["id"]

	// Convert The ID To Int
	idInt, err := strconv.Atoi(id)

	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid ID: %s", id), http.StatusBadRequest)
		return -1
	}

	return idInt
}

// WriteError writes an error message as a JSON response with the specified status code.
// The error message is formatted as a map with a single key "error" containing the provided message.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, APIError{Error: message})
}
