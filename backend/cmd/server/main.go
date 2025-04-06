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
	http.HandleFunc("/cart/add", handlers.AddToCartHandler)
	http.HandleFunc("/cart/item", handlers.DeleteCartItemHandler)
	http.HandleFunc("/cart", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			handlers.DeleteCartHandler(w, r)
		} else if r.Method == http.MethodGet {
			handlers.GetCartHandler(w, r) // your existing GET handler for /cart
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/api/simulation", handlers.GetUserClimateSimulationHandler)
	http.HandleFunc("/cart/carbon-footprint", handlers.GetCarbonFootprintHandler)

	fmt.Println("Starting server on :8080 ...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
