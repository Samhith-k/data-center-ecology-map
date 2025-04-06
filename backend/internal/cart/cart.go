package cart

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/Samhith-k/data-center-ecology-map/backend/internal/data"
)

var (
	// carts holds the in‑memory mapping from username to Cart.
	carts   = make(map[string]*Cart)
	cartMu  sync.RWMutex
	cartDir = "./carts" // directory where cart files are stored
)

// Cart represents a user's shopping cart.
type Cart struct {
	Username  string                    `json:"username"`
	Items     []data.DatacenterLocation `json:"items"`
	MoneyLeft float64                   `json:"money_left"`
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
