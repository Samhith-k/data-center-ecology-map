package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Samhith-k/data-center-ecology-map/backend/internal/handlers"
	"github.com/Samhith-k/data-center-ecology-map/backend/internal/user"
)

func main() {
	// Example usage of your “load users, define routes, start server” logic
	err := user.LoadUserPasswords("users.txt")
	if err != nil {
		log.Fatalf("Error loading users: %v\n", err)
	}

	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/profile", handlers.ProfileHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/alldatacenters", handlers.AllDataCentersHandler)
	http.HandleFunc("/api/possible-datacenters", handlers.PossibleDataCenterHandler)
	http.HandleFunc("/api/property-details", handlers.GetPropertyDetailsHandler)

	fmt.Println("Starting server on :8080 ...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
