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

// Handle The Account Routes
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

// Handle The Get Accounts Route
// Return All Accounts
func (as *APIServer) handleGetAccounts(w http.ResponseWriter, r *http.Request) error {
	// Get All Accounts From The Store
	accs, err := as.store.GetAccounts()

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accs)
}

// Handle The Create Account Route
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

// Handle The Get Account By ID Route
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

// Handle The Delete Account Route
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

// Handle The Transfer Route
func (as *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transferReq := TransferRequest{}

	if err := json.NewDecoder(r.Body).Decode(&transferReq); err != nil {
		return err
	}
	defer r.Body.Close()

	// // Get The Account By ID
	// toAcc, err := as.store.GetAccountById(transferReq.ToAccountID)

	// if err != nil {
	// 	return err
	// }

	return WriteJSON(w, http.StatusOK, transferReq)
}

// Run The API Server
func (as *APIServer) Run() {
	// Create The Router and SubRouter
	router := mux.NewRouter()
	subRouter := router.PathPrefix("/api/v1").Subrouter()

	// Handle The Accounts Routes
	subRouter.HandleFunc("/account", makeHTTPHandlerFunc(as.handleAccount))
	subRouter.HandleFunc("/account/{id:[0-9]+}", makeHTTPHandlerFunc(as.handleGetAccountById)).Methods("GET")
	subRouter.HandleFunc("/account/{id:[0-9]+}", makeHTTPHandlerFunc(as.handleDeleteAccount)).Methods("DELETE")

	// Handle The Transfer Route
	subRouter.HandleFunc("/transfer", makeHTTPHandlerFunc(as.handleTransfer))

	// Run The HTTPServer

	log.Println("Server is Runing On Port: ", as.listenAddr)

	http.ListenAndServe(as.listenAddr, router)
}

/*
API Helper Functions
*/
type apiFunc func(http.ResponseWriter, *http.Request) error

/*
API Error Response
*/
type APIError struct {
	Error string `json:"error"`
}

/*
Make HTTP Handler Function Wrapper
*/
func makeHTTPHandlerFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, APIError{
				Error: err.Error(),
			})
		}
	}
}

/*
Write JSON Response Utility
*/
func WriteJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

/*
Get ID From URL
*/
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
