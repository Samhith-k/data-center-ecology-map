package main

import (
	"bufio"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// Credentials is used to parse incoming JSON for login/register
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// userPasswords stores "username -> hashedPassword"
// In real apps, store this in a DB, not a file or in-memory map.
var userPasswords = make(map[string]string)

// sessions maps "sessionID -> username"
var sessions = make(map[string]string)

// For thread-safety around sessions/userPasswords
// (since multiple requests can happen concurrently)
var mu sync.RWMutex

func main() {
	// 1. Load existing users from users.txt
	err := loadUserPasswords("users.txt")
	if err != nil {
		log.Fatalf("Error loading users: %v\n", err)
	}

	// 2. Define routes
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/profile", profileHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/alldatacenters", allDataCentersHandler)
	http.HandleFunc("/api/possible-datacenters", possibleDataCenterHandler)
	http.HandleFunc("/api/property-details", getPropertyDetailsHandler)
	// http.HandleFunc("/logout", logoutHandler)

	// If you build your React app into ./frontend/build, you can serve it:
	// fs := http.FileServer(http.Dir("./frontend/build"))
	// http.Handle("/", fs)

	// 3. Start the server
	fmt.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

//==========================//
//   HANDLER IMPLEMENTATION
//==========================//

type DataCenter struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Define a struct for datacenter locations
type DatacenterLocation struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Name        string  `json:"name,omitempty"`
	LandPrice   string  `json:"land_price,omitempty"`
	Electricity string  `json:"electricity,omitempty"`
	Notes       string  `json:"notes,omitempty"`
}

// Helper function to read from CSV file
func readDatacenterLocations() ([]DatacenterLocation, error) {
	file, err := os.Open("us_possible_locations.csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Skip header
	_, err = reader.Read()
	if err != nil {
		return nil, err
	}

	var locations []DatacenterLocation
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		latitude, err := strconv.ParseFloat(record[0], 64)
		if err != nil {
			return nil, err
		}

		longitude, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return nil, err
		}

		location := DatacenterLocation{
			Latitude:    latitude,
			Longitude:   longitude,
			Name:        record[2],
			LandPrice:   record[3],
			Electricity: record[4],
			Notes:       record[5],
		}

		locations = append(locations, location)
	}

	return locations, nil
}

// Handler for getting all possible datacenter locations
func possibleDataCenterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	locations, err := readDatacenterLocations()
	if err != nil {
		http.Error(w, "Error reading datacenter locations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a simplified response with only lat/long
	var response []map[string]float64
	for _, loc := range locations {
		response = append(response, map[string]float64{
			"latitude":  loc.Latitude,
			"longitude": loc.Longitude,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Handler for getting property details based on lat/long
func getPropertyDetailsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse latitude and longitude from query parameters
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")

	if latStr == "" || lngStr == "" {
		http.Error(w, "Missing latitude or longitude parameters", http.StatusBadRequest)
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		http.Error(w, "Invalid latitude format", http.StatusBadRequest)
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		http.Error(w, "Invalid longitude format", http.StatusBadRequest)
		return
	}

	// Define a small epsilon for floating-point comparison
	const epsilon = 0.0001

	locations, err := readDatacenterLocations()
	if err != nil {
		http.Error(w, "Error reading datacenter locations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Find the matching location
	var matchedLocation *DatacenterLocation
	for _, loc := range locations {
		if math.Abs(loc.Latitude-lat) < epsilon && math.Abs(loc.Longitude-lng) < epsilon {
			matchedLocation = &loc
			break
		}
	}

	if matchedLocation == nil {
		http.Error(w, "No property found at the specified coordinates", http.StatusNotFound)
		return
	}

	// Return all details except lat/long
	response := map[string]string{
		"location_name": matchedLocation.Name,
		"land_price":    matchedLocation.LandPrice,
		"electricity":   matchedLocation.Electricity,
		"notes":         matchedLocation.Notes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func allDataCentersHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Open the CSV file
	file, err := os.Open("us_datacenters.csv")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open file: %v", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// 2. Read line by line using a scanner
	scanner := bufio.NewScanner(file)
	var dataCenters []DataCenter

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			// Skip any empty or whitespace-only lines
			continue
		}

		// Parse the line from the end to allow commas in the name
		dc, parseErr := parseDataCenterLine(line)
		if parseErr != nil {
			// You could handle the error differently (e.g., log and skip line)
			log.Printf("Skipping malformed line: %v", parseErr)
			continue
		}
		dataCenters = append(dataCenters, *dc)
	}

	if err := scanner.Err(); err != nil {
		http.Error(w, fmt.Sprintf("Error reading file: %v", err), http.StatusInternalServerError)
		return
	}

	// 3. Encode the slice to JSON and return the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dataCenters); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode JSON: %v", err), http.StatusInternalServerError)
		return
	}
}

// parseDataCenterLine splits a line on commas, assuming the last two fields
// are latitude and longitude, and the rest (possibly containing commas) is 'name'.
func parseDataCenterLine(line string) (*DataCenter, error) {
	parts := strings.Split(line, ",")
	if len(parts) < 3 {
		return nil, fmt.Errorf("line has fewer than 3 comma-separated fields: %s", line)
	}

	// Last token is longitude, second-to-last is latitude
	longitudeStr := parts[len(parts)-1]
	latitudeStr := parts[len(parts)-2]

	// Everything before that is the name (re-join if it had commas)
	name := strings.Join(parts[:len(parts)-2], ",")

	// Convert lat/lon to float64
	lat, err := strconv.ParseFloat(latitudeStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid latitude '%s': %v", latitudeStr, err)
	}
	lng, err := strconv.ParseFloat(longitudeStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid longitude '%s': %v", longitudeStr, err)
	}

	return &DataCenter{
		Name:      name,
		Latitude:  lat,
		Longitude: lng,
	}, nil
}

// registerHandler adds a new user to users.txt with a bcrypt-hashed password
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Cannot parse request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if creds.Username == "" || creds.Password == "" {
		http.Error(w, "Username and password required", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	mu.RLock()
	_, exists := userPasswords[creds.Username]
	mu.RUnlock()
	if exists {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Hash the password
	hashed, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Write to users.txt
	// In a real system, you'd do this in a DB transaction. We keep it simple here.
	f, err := os.OpenFile("users.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		http.Error(w, "Cannot open user file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	line := fmt.Sprintf("%s:%s\n", creds.Username, string(hashed))
	if _, err := f.WriteString(line); err != nil {
		http.Error(w, "Error writing user file", http.StatusInternalServerError)
		return
	}

	// Update our in-memory map as well
	mu.Lock()
	userPasswords[creds.Username] = string(hashed)
	mu.Unlock()

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "User registered successfully!",
	})
}

// loginHandler verifies user credentials, creates a session ID, sets a cookie, and returns it
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Cannot parse request body", http.StatusBadRequest)
		return
	}

	mu.RLock()
	hashedPass, ok := userPasswords[creds.Username]
	mu.RUnlock()
	if !ok {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Compare provided password with stored hash
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(creds.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Create a session ID
	sessionID := generateSessionID()

	// Store the session -> user mapping
	mu.Lock()
	sessions[sessionID] = creds.Username
	mu.Unlock()

	// Option 1: return session ID as JSON
	// Option 2: set it in a cookie
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HttpOnly: true,
		Path:     "/",
		// Secure: true, // should be set in production with HTTPS
		// SameSite: http.SameSiteStrictMode, // recommended for better security
	}
	http.SetCookie(w, cookie)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "success",
		"message":   "Logged in!",
		"sessionId": sessionID, // returning also in the JSON (if the client wants to store it)
	})
}

// profileHandler checks if the user has a valid session, returns "protected" data
func profileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check session cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "No session cookie found, please login", http.StatusUnauthorized)
		return
	}

	mu.RLock()
	username, valid := sessions[cookie.Value]
	mu.RUnlock()
	if !valid {
		http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"profile": fmt.Sprintf("Hello %s! This is protected profile data!", username),
	})
}

// logoutHandler invalidates the user's session
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "No session cookie found", http.StatusUnauthorized)
		return
	}

	mu.Lock()
	delete(sessions, cookie.Value)
	mu.Unlock()

	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1, // immediate expiration
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Logged out successfully!",
	})
}

//==========================//
//   HELPER FUNCTIONS
//==========================//

// loadUserPasswords reads the file line by line, storing "username -> hashedPassword" in a map
func loadUserPasswords(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// If the file doesn't exist, it's okay. We'll create it on register.
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
			hashedPassword := parts[1]
			userPasswords[username] = hashedPassword
		}
	}
	return scanner.Err()
}

// generateSessionID creates a random 16-byte hex string
func generateSessionID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback: just use a timestamp or something else if rand fails
		return fmt.Sprintf("session_%d", len(sessions)+1)
	}
	return hex.EncodeToString(b)
}
