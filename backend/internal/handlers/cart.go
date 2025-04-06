package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Samhith-k/data-center-ecology-map/backend/internal/cart"
	"github.com/Samhith-k/data-center-ecology-map/backend/internal/data"
)

// AddToCartRequest is the expected JSON payload for adding an item.
type AddToCartRequest struct {
	Username string                  `json:"username"`
	Item     data.DatacenterLocation `json:"item"`
	Cost     float64                 `json:"cost"`
}

func GetCarbonFootprintHandler(w http.ResponseWriter, r *http.Request) {
	addCORSHeaders(w) // if you have a helper for CORS
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Missing username query parameter", http.StatusBadRequest)
		return
	}

	footprint, err := cart.CalculateCarbonFootprint(username)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error calculating footprint: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{
		"carbon_footprint": footprint,
	})
}

// AddToCartHandler handles POST /cart/add.
func AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	addCORSHeaders(w)
	fmt.Println("AddToCartHandler called")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if req.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}
	if err := cart.AddToCart(req.Username, req.Item, req.Cost); err != nil {
		http.Error(w, fmt.Sprintf("Error adding to cart: %v", err), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Item added to cart",
	})
}

// GetCartHandler handles GET /cart?username=...
func GetCartHandler(w http.ResponseWriter, r *http.Request) {
	addCORSHeaders(w)
	fmt.Println("GetCartHandler called")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}
	c, exists := cart.GetCart(username)
	if !exists {
		http.Error(w, "Cart not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

// DeleteCartItemHandler handles DELETE /cart/item?username=alice&index=0
func DeleteCartItemHandler(w http.ResponseWriter, r *http.Request) {
	addCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := r.URL.Query().Get("username")
	indexStr := r.URL.Query().Get("index")
	if username == "" || indexStr == "" {
		http.Error(w, "username and index parameters are required", http.StatusBadRequest)
		return
	}
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Error(w, "Invalid index value", http.StatusBadRequest)
		return
	}

	if err := cart.RemoveItemFromCart(username, index); err != nil {
		http.Error(w, fmt.Sprintf("Error deleting cart item: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Cart item deleted",
	})
}

// DeleteCartHandler handles DELETE /cart?username=alice
func DeleteCartHandler(w http.ResponseWriter, r *http.Request) {
	addCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "username parameter is required", http.StatusBadRequest)
		return
	}
	if err := cart.DeleteCart(username); err != nil {
		http.Error(w, fmt.Sprintf("Error deleting cart: %v", err), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Cart deleted",
	})
}
