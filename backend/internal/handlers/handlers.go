package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/Samhith-k/data-center-ecology-map/backend/internal/data"
	"github.com/Samhith-k/data-center-ecology-map/backend/internal/session"
	"github.com/Samhith-k/data-center-ecology-map/backend/internal/user"
	"golang.org/x/crypto/bcrypt"
)

// RegisterHandler handles POST /register
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	addCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds user.Credentials
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
	if user.Exists(creds.Username) {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Hash the password
	hashed, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Write to users.txt and update in-memory map
	if err := user.AddUser(creds.Username, string(hashed)); err != nil {
		http.Error(w, "Error writing user file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "User registered successfully!",
	})
}

// LoginHandler handles POST /login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	addCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds user.Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Cannot parse request body", http.StatusBadRequest)
		return
	}

	hashedPass, ok := user.GetHashedPassword(creds.Username)
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
	sessionID := session.GenerateSessionID()
	// Store the session -> user mapping
	session.SetUserForSession(sessionID, creds.Username)

	// Set it in a cookie
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, cookie)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "success",
		"message":   "Logged in!",
		"sessionId": sessionID,
	})
}

// ProfileHandler handles GET /profile
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	addCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "No session cookie found, please login", http.StatusUnauthorized)
		return
	}

	username, valid := session.GetUserForSession(cookie.Value)
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

// LogoutHandler handles POST /logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	addCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "No session cookie found", http.StatusUnauthorized)
		return
	}

	session.ClearSession(cookie.Value)

	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Logged out successfully!",
	})
}

// AllDataCentersHandler handles GET /alldatacenters
func AllDataCentersHandler(w http.ResponseWriter, r *http.Request) {
	addCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	dataCenters, err := data.ReadAllDataCenters("us_datacenters.csv")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open file: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dataCenters); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode JSON: %v", err), http.StatusInternalServerError)
		return
	}
}

// PossibleDataCenterHandler handles GET /api/possible-datacenters
func PossibleDataCenterHandler(w http.ResponseWriter, r *http.Request) {
	addCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	locations, err := data.ReadDatacenterLocations("us_possible_locations.csv")
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

// GetPropertyDetailsHandler handles GET /api/property-details?lat=..&lng=..
func GetPropertyDetailsHandler(w http.ResponseWriter, r *http.Request) {
	addCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
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

	const epsilon = 0.0001
	locations, err := data.ReadDatacenterLocations("us_possible_locations.csv")
	if err != nil {
		http.Error(w, "Error reading datacenter locations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	existingDCs, err := data.ReadExistingDatacenters("us_datacenters.csv")
	if err != nil {
		log.Printf("Warning: Could not read existing datacenter data: %v", err)
	}

	// Find the matching location
	var matched *data.DatacenterLocation
	for i := range locations {
		if math.Abs(locations[i].Latitude-lat) < epsilon &&
			math.Abs(locations[i].Longitude-lng) < epsilon {
			matched = &locations[i]

			// Calculate environmental metrics
			allDCs := append(locations, existingDCs...)
			CalculateResearchBasedMetrics(matched, allDCs) // see envcalcs.go
			break
		}
	}

	if matched == nil {
		http.Error(w, "No property found at the specified coordinates", http.StatusNotFound)
		return
	}

	// Return all details including environmental metrics
	response := map[string]interface{}{
		"location_name":            matched.Name,
		"land_price":               matched.LandPrice,
		"electricity":              matched.Electricity,
		"notes":                    matched.Notes,
		"eco_score":                matched.EcoScore,
		"carbon_impact":            matched.CarbonImpact,
		"temp_increase":            matched.TempIncrease,
		"water_usage":              matched.WaterUsage,
		"renewable_access":         matched.RenewableAccess,
		"datacenter_density":       matched.DatacenterDensity,
		"density_impact_score":     matched.DensityImpactScore,
		"compounded_temp_increase": matched.CompoundedTempIncrease,
		"water_competition":        matched.WaterCompetition,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// addCORSHeaders is a helper that adds CORS-related headers
func addCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}
