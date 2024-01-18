package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Customer struct {
	ID    uuid.UUID `json:"id"`
	Fname string    `json:"fname"`
	Lname string    `json:"lname"`
	Email string    `json:"email"`
}

type Address struct {
	ID          uuid.UUID `json:"id"`
	Customer_id uuid.UUID `json:"customer_id"`
	Nickname    string    `json:"nickname"`
	Street1     string    `json:"street1"`
	Street2     string    `json:"street2"`
	City        string    `json:"city"`
	State       string    `json:"state"`
	Zipcode     string    `json:"zipcode"`
}

func main() {
	//connection to database
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//create customer table if it doesn't exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS customers(id UUID PRIMARY KEY, fname VARCHAR NOT NULL, lname VARCHAR, email VARCHAR UNIQUE);")
	if err != nil {
		log.Fatal(err)
	}

	//create address table if it doesn't exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS addresses(id UUID PRIMARY KEY, customer_id UUID REFERENCES customers(id), nickname VARCHAR, street1 VARCHAR UNIQUE, street2 VARCHAR, city VARCHAR, state VARCHAR, zipcode VARCHAR);")
	if err != nil {
		log.Fatal(err)
	}

	//create router
	router := mux.NewRouter()
	credentials := handlers.AllowCredentials()
	methods := handlers.AllowedMethods([]string{"POST", "GET"})
	//ttl := handlers.MaxAge(3600)
	origins := handlers.AllowedOrigins([]string{"*"})
	router.HandleFunc("/customers", getCustomers(db)).Methods("GET")
	router.HandleFunc("/customer/{id}", getCustomerByID(db)).Methods("GET")
	router.HandleFunc("/customers", createCustomer(db)).Methods("POST")
	router.HandleFunc("/customer/{customer_id}/addresses", getAddresses(db)).Methods("GET")
	router.HandleFunc("/customer/{customer_id}/addresses", createAddress(db)).Methods("POST")

	//start server
	//log.Fatal(http.ListenAndServe(":8000", jsonContentTypeMiddleware(router)))
	log.Fatal(http.ListenAndServe(":8000", handlers.CORS(credentials, methods, origins)(router)))
}

func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// get all customers
func getCustomers(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM customers")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var customers []Customer
		for rows.Next() {
			var customer Customer
			err := rows.Scan(&customer.ID, &customer.Fname, &customer.Lname, &customer.Email)
			if err != nil {
				log.Fatal(err)
			}
			customers = append(customers, customer)
		}

		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(customers)
	}
}

// get customer by id
func getCustomerByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		row := db.QueryRow("SELECT id, fname, lname, email FROM customers WHERE id = $1", id)
		var customer Customer
		err := row.Scan(&customer.ID, &customer.Fname, &customer.Lname, &customer.Email)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(customer)
	}
}

// create customer
func createCustomer(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		myUUID := uuid.New()

		var customer Customer
		err := json.NewDecoder(r.Body).Decode(&customer)
		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec("INSERT INTO customers (id, fname, lname, email) VALUES ($1, $2, $3, $4)", myUUID, customer.Fname, customer.Lname, customer.Email)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(customer)
	}
}

// get all addresses for a customer
func getAddresses(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		customer_id := mux.Vars(r)["customer_id"]

		rows, err := db.Query("SELECT * FROM addresses WHERE customer_id = $1", customer_id)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var addresses []Address
		for rows.Next() {
			var address Address
			err := rows.Scan(&address.ID, &address.Customer_id, &address.Nickname, &address.Street1, &address.Street2, &address.City, &address.State, &address.Zipcode)
			if err != nil {
				log.Fatal(err)
			}
			addresses = append(addresses, address)
		}

		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(addresses)
	}
}

// create address of customer
func createAddress(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		customer_id := mux.Vars(r)["customer_id"]
		myUUID := uuid.New()

		var address Address
		err := json.NewDecoder(r.Body).Decode(&address)
		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec("INSERT INTO addresses (id, customer_id, nickname, street1, street2, city, state, zipcode) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", myUUID, customer_id, address.Nickname, address.Street1, address.Street2, address.City, address.State, address.Zipcode)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(address)
	}
}
