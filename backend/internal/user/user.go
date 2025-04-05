package user

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Credentials holds the incoming JSON fields
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// in-memory map for user data
var userPasswords = make(map[string]string)
var mu sync.RWMutex

// LoadUserPasswords reads the file line by line, storing "username -> hashedPassword"
func LoadUserPasswords(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// If no file, that’s okay
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			username := parts[0]
			hashed := parts[1]
			userPasswords[username] = hashed
		}
	}
	return scanner.Err()
}

// AddUser adds a new user with hashed password to userPasswords and to file
func AddUser(username, hashedPass string) error {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := userPasswords[username]; exists {
		return fmt.Errorf("user %s already exists", username)
	}

	// append to file
	f, err := os.OpenFile("users.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	line := fmt.Sprintf("%s:%s\n", username, hashedPass)
	if _, err := f.WriteString(line); err != nil {
		return err
	}

	// update in-memory
	userPasswords[username] = hashedPass
	return nil
}

// Exists checks if user already in the map
func Exists(username string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := userPasswords[username]
	return ok
}

// GetHashedPassword returns the hashed password for a user, or false if not found
func GetHashedPassword(username string) (string, bool) {
	mu.RLock()
	defer mu.RUnlock()
	hp, ok := userPasswords[username]
	return hp, ok
}
