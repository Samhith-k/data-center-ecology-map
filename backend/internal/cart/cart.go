package cart

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Samhith-k/data-center-ecology-map/backend/internal/data"
)

var (
	// carts holds the in‑memory mapping from username to Cart.
	carts   = make(map[string]*Cart)
	cartMu  sync.RWMutex
	cartDir = "./carts" // directory where cart files are stored
)

type CartItem = data.DatacenterLocation

// Cart represents a user's shopping cart.
type Cart struct {
	Username  string     `json:"username"`
	Items     []CartItem `json:"items"`
	MoneyLeft float64    `json:"money_left"`
}

// LoadAllCarts loads all cart files from disk when the app starts.
func LoadAllCarts() error {
	// Ensure the cart directory exists.
	if err := os.MkdirAll(cartDir, 0755); err != nil {
		return err
	}
	files, err := ioutil.ReadDir(cartDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) != ".cart" {
			continue
		}
		username := file.Name()[0 : len(file.Name())-len(".cart")]
		path := filepath.Join(cartDir, file.Name())
		content, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading cart file %s: %v\n", path, err)
			continue
		}
		var c Cart
		if err := json.Unmarshal(content, &c); err != nil {
			fmt.Printf("Error unmarshaling cart file %s: %v\n", path, err)
			continue
		}
		cartMu.Lock()
		carts[username] = &c
		cartMu.Unlock()
	}
	return nil
}

// SaveCartNoLock saves the given cart assuming the lock is already held.
func SaveCartNoLock(username string, c *Cart) error {
	// Ensure the cart directory exists
	if err := os.MkdirAll(cartDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(cartDir, username+".cart")
	return ioutil.WriteFile(path, data, 0644)
}

// GetCart returns the cart for a given user.
func GetCart(username string) (*Cart, bool) {
	cartMu.RLock()
	defer cartMu.RUnlock()
	c, ok := carts[username]
	return c, ok
}

// AddToCart adds a datacenter item to the user's cart and deducts the cost.
func AddToCart(username string, item data.DatacenterLocation, cost float64) error {
	cartMu.Lock()
	defer cartMu.Unlock()
	c, exists := carts[username]
	if !exists {
		// If no cart exists, create a new one with a default money value.
		c = &Cart{
			Username:  username,
			Items:     []data.DatacenterLocation{},
			MoneyLeft: 1000000, // starting funds (e.g., $1,000,000)
		}
		carts[username] = c
	}
	if c.MoneyLeft < cost {
		return fmt.Errorf("insufficient funds: available %f, cost %f", c.MoneyLeft, cost)
	}
	c.Items = append(c.Items, item)
	c.MoneyLeft -= cost
	// Use the no-lock version since the write lock is held.
	return SaveCartNoLock(username, c)
}

// RemoveItemFromCart removes an item at the given index from the user's cart.
func RemoveItemFromCart(username string, index int) error {
	cartMu.Lock()
	defer cartMu.Unlock()

	c, exists := carts[username]
	if !exists {
		return fmt.Errorf("cart not found for user %s", username)
	}

	if index < 0 || index >= len(c.Items) {
		return fmt.Errorf("invalid index %d", index)
	}

	// Remove the item by slicing it out
	c.Items = append(c.Items[:index], c.Items[index+1:]...)
	return SaveCartNoLock(username, c)
}

// DeleteCart deletes the entire cart for a user.
func DeleteCart(username string) error {
	cartMu.Lock()
	defer cartMu.Unlock()

	// Remove from in-memory map
	delete(carts, username)

	// Remove the file on disk.
	path := filepath.Join(cartDir, username+".cart")
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete cart file: %v", err)
	}
	return nil
}

func CalculateCarbonFootprint(username string) (float64, error) {
	cartMu.RLock()
	c, exists := carts[username]
	cartMu.RUnlock()
	if !exists {
		// if no cart, zero footprint
		return 0, nil
	}

	var totalCarbon float64
	for _, item := range c.Items {
		totalCarbon += computeCarbonForItem(item)
	}
	return totalCarbon, nil
}

// Simple example: parse item notes or name to guess carbon
func computeCarbonForItem(item data.DatacenterLocation) float64 {
	// Basic example logic:
	//   0.8 for "Standard"
	//   0.4 for "Eco"
	//   0.1 for "Next-Gen"
	base := 1.0
	nameLower := strings.ToLower(item.Name)
	notesLower := strings.ToLower(item.Notes)

	switch {
	case strings.Contains(nameLower, "eco") || strings.Contains(notesLower, "eco"):
		base = 0.4
	case strings.Contains(nameLower, "next-gen") || strings.Contains(notesLower, "next-gen"):
		base = 0.1
	case strings.Contains(nameLower, "standard") || strings.Contains(notesLower, "standard"):
		base = 0.8
	default:
		base = 1.0 // fallback
	}

	// Possibly adjust for electricity, e.g. if item.Electricity includes "renewable"
	if strings.Contains(strings.ToLower(item.Electricity), "renewable") {
		base *= 0.5
	}

	// Return that as "metric tons (MT) per day" or whatever your convention
	return base
}
