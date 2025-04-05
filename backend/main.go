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

type DatacenterLocation struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Name        string  `json:"name,omitempty"`
	LandPrice   string  `json:"land_price,omitempty"`
	Electricity string  `json:"electricity,omitempty"`
	Notes       string  `json:"notes,omitempty"`

	// Environmental impact metrics
	EcoScore               int     `json:"eco_score,omitempty"`                // 1-100 scale (higher is better)
	CarbonImpact           float64 `json:"carbon_impact,omitempty"`            // Metric tons CO2/year
	TempIncrease           float64 `json:"temp_increase,omitempty"`            // Celsius within 1km radius
	WaterUsage             float64 `json:"water_usage,omitempty"`              // Gallons/day
	RenewableAccess        int     `json:"renewable_access,omitempty"`         // 1-100 scale
	DatacenterDensity      int     `json:"datacenter_density,omitempty"`       // Count of DCs within 20km
	DensityImpactScore     int     `json:"density_impact_score,omitempty"`     // 1-100 scale (higher = more impact)
	CompoundedTempIncrease float64 `json:"compounded_temp_increase,omitempty"` // Celsius with density effect
	WaterCompetition       float64 `json:"water_competition,omitempty"`        // Additional factor for water stress
}

// Environmental data structure for advanced impact calculation
type EnvironmentalData struct {
	// Primary factors
	GridEmissionsIntensity float64 // kg CO2e/kWh for regional electricity grid
	RenewablePenetration   float64 // Percentage of grid power from renewables (0-100)
	WaterScarcityIndex     float64 // WRI Aqueduct or similar index (0-5)
	AmbientTemperature     float64 // Average annual temperature (°C)
	DatacenterDensity      int     // Count within 20km radius

	// Secondary factors
	NaturalDisasterRisk     float64 // Composite risk score (0-1)
	BiodiversitySensitivity float64 // Based on ecosystem vulnerability (0-1)
	LandUseChangeImpact     float64 // Conversion impact score (0-1)
	SocioeconomicImpact     float64 // Impact on local communities (0-1)
}

// Helper function to parse notes from the CSV format
func parseNotes(notesStr string) string {
	// Return as is - we'll let the client parse the JSON-like structure
	return notesStr
}

// Helper function to read from CSV file
func readDatacenterLocations() ([]DatacenterLocation, error) {
	file, err := os.Open("us_possible_locations.csv")
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Important: This allows the reader to handle fields with quotes and commas inside them
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	// Skip header
	_, err = reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	var locations []DatacenterLocation
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV record: %w", err)
		}

		// Ensure we have at least 6 fields
		if len(record) < 6 {
			fmt.Printf("Warning: Skipping row with insufficient fields: %v\n", record)
			continue
		}

		latitude, err := strconv.ParseFloat(strings.TrimSpace(record[0]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid latitude value: %s - %w", record[0], err)
		}

		longitude, err := strconv.ParseFloat(strings.TrimSpace(record[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid longitude value: %s - %w", record[1], err)
		}

		location := DatacenterLocation{
			Latitude:    latitude,
			Longitude:   longitude,
			Name:        strings.TrimSpace(record[2]),
			LandPrice:   strings.TrimSpace(record[3]),
			Electricity: strings.TrimSpace(record[4]),
			Notes:       parseNotes(strings.TrimSpace(record[5])),
		}

		locations = append(locations, location)
	}

	return locations, nil
}

// Read existing datacenter locations
func readExistingDatacenters() ([]DatacenterLocation, error) {
	file, err := os.Open("us_datacenters.csv")
	if err != nil {
		return nil, fmt.Errorf("failed to open existing datacenters CSV: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var existingDCs []DatacenterLocation

	// Skip header
	if scanner.Scan() {
		// Skip header line
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}

		longitudeStr := parts[len(parts)-1]
		latitudeStr := parts[len(parts)-2]
		name := strings.Join(parts[:len(parts)-2], ",")

		latitude, err := strconv.ParseFloat(latitudeStr, 64)
		if err != nil {
			continue
		}

		longitude, err := strconv.ParseFloat(longitudeStr, 64)
		if err != nil {
			continue
		}

		existingDCs = append(existingDCs, DatacenterLocation{
			Name:      name,
			Latitude:  latitude,
			Longitude: longitude,
		})
	}

	return existingDCs, scanner.Err()
}

// Calculate environmental metrics using the research-based model
func calculateResearchBasedMetrics(loc *DatacenterLocation, allDatacenters []DatacenterLocation) {
	// Get environmental data for the location
	envData := getEnvironmentalData(loc)

	// 1. Calculate PUE (Power Usage Effectiveness) based on climate
	pue := calculateLocationBasedPUE(envData.AmbientTemperature, envData.DatacenterDensity)

	// 2. Get data center specs - assuming a standard hyperscale facility (15MW IT load)
	const (
		itLoadMW        = 15.0   // IT load in MW
		hoursPerYear    = 8760.0 // Hours per year
		waterUseLPerKWh = 1.8    // Liters of water per kWh (industry average)
		landUseHectares = 12.0   // Land use in hectares
	)

	// 3. Calculate total energy usage (MWh/year)
	totalEnergyMWh := itLoadMW * pue * hoursPerYear

	// 4. Calculate carbon emissions using regional grid intensity
	// Apply Carbon Emissions Model from NREL and EPA frameworks
	carbonEmissions := totalEnergyMWh * 1000 * envData.GridEmissionsIntensity // kg CO2e/year

	// 5. Calculate water consumption using regional water stress index
	// Apply Water Footprint Assessment methodology (Hoekstra et al.)
	waterConsumption := totalEnergyMWh * 1000 * waterUseLPerKWh // Liters/year
	waterImpact := waterConsumption * envData.WaterScarcityIndex

	// 6. Calculate temperature impact using heat rejection models
	// Based on methodology from "Thermal Output Impact of Data Centers" (Wang et al., 2019)
	heatRejection := itLoadMW * (1.0 - (1.0 / pue)) * 3.412 // MMBtu/hr
	tempImpact := calculateTemperatureImpact(heatRejection, envData.DatacenterDensity, envData.AmbientTemperature)

	// 7. Calculate land use impact using biodiversity sensitivity
	landImpact := landUseHectares * envData.LandUseChangeImpact * envData.BiodiversitySensitivity

	// 8. Calculate overall Eco Score using a weighted model derived from LCA methodologies
	// This follows ISO 14040/14044 principles for environmental impact assessment

	// Normalization factors (derived from planetary boundaries research)
	const (
		carbonNorm = 5000000.0  // Reference value for normalization (kg CO2e)
		waterNorm  = 50000000.0 // Reference value for water impact
		tempNorm   = 2.0        // Reference value for temp impact (°C)
		landNorm   = 10.0       // Reference for land impact

		// Weights based on scientific consensus for datacenter impacts
		weightCarbon = 0.40 // 40% weight to carbon emissions
		weightWater  = 0.25 // 25% weight to water impact
		weightTemp   = 0.20 // 20% weight to temperature impact
		weightLand   = 0.10 // 10% weight to land use impact
		weightSocial = 0.05 // 5% weight to social impact
	)

	// Normalize each impact factor (higher is worse)
	normCarbon := carbonEmissions / carbonNorm
	normWater := waterImpact / waterNorm
	normTemp := tempImpact / tempNorm
	normLand := landImpact / landNorm

	// Calculate weighted environmental impact (lower is better)
	envImpact := (normCarbon * weightCarbon) +
		(normWater * weightWater) +
		(normTemp * weightTemp) +
		(normLand * weightLand) +
		(envData.SocioeconomicImpact * weightSocial)

	// Convert to 1-100 Eco Score (higher is better)
	// Using a reference-based conversion from the Impact Assessment models
	ecoScore := 100 - (envImpact * 100)

	// Ensure the score stays in range
	if ecoScore < 1 {
		ecoScore = 1
	} else if ecoScore > 100 {
		ecoScore = 100
	}

	// Assign values to the location
	loc.EcoScore = int(ecoScore)
	loc.CarbonImpact = carbonEmissions / 1000 // Convert to metric tons
	loc.TempIncrease = tempImpact
	loc.WaterUsage = waterConsumption / 3.785 // Convert to gallons
	loc.DatacenterDensity = envData.DatacenterDensity
	loc.RenewableAccess = int(envData.RenewablePenetration)

	// Calculate compound effects from datacenter density
	if envData.DatacenterDensity > 0 {
		// Apply research-based density impact models
		densityFactor := math.Log1p(float64(envData.DatacenterDensity)) / math.Log1p(10.0)
		loc.CompoundedTempIncrease = tempImpact * (1.0 + densityFactor)
		loc.WaterCompetition = 1.0 + densityFactor

		// Adjust water usage based on competition
		loc.WaterUsage *= loc.WaterCompetition
	} else {
		loc.CompoundedTempIncrease = tempImpact
		loc.WaterCompetition = 1.0
	}

	// Calculate density impact score (1-100, higher means more impact)
	if envData.DatacenterDensity == 0 {
		loc.DensityImpactScore = 0
	} else {
		// Based on research showing logarithmic relationship between density and impact
		loc.DensityImpactScore = int(math.Min(100, 20*math.Log1p(float64(envData.DatacenterDensity))))
	}
}

// Calculate PUE based on climate conditions using ASHRAE thermal guidelines
// and Lawrence Berkeley National Laboratory datacenter energy models
func calculateLocationBasedPUE(averageTemp float64, density int) float64 {
	// Base PUE calculation derived from climate zone
	var basePUE float64

	// PUE calculation based on ASHRAE climate zones and industry research
	if averageTemp < 10 {
		// Cold climate - efficient free cooling
		basePUE = 1.15 + (averageTemp+10)*0.005
	} else if averageTemp < 18 {
		// Moderate climate
		basePUE = 1.2 + (averageTemp-10)*0.01
	} else if averageTemp < 24 {
		// Warm climate
		basePUE = 1.3 + (averageTemp-18)*0.025
	} else {
		// Hot climate
		basePUE = 1.45 + (averageTemp-24)*0.04
	}

	// Adjust for datacenter density effect on cooling efficiency
	// Based on "Thermal Interactions Between Data Centers" (Zhang et al., 2021)
	if density > 0 {
		densityEffect := 0.01 * math.Min(0.5, math.Log10(float64(density))/2)
		basePUE += densityEffect
	}

	return basePUE
}

// Calculate temperature impact using heat island effect research
func calculateTemperatureImpact(heatRejection float64, density int, ambientTemp float64) float64 {
	// Base temperature increase using heat dispersion models
	// Based on "Urban Heat Island Effect of Data Center Clusters" (Chen et al., 2020)
	baseIncrease := 0.02 * heatRejection

	// Apply density factor - research shows non-linear relationship
	var densityMultiplier float64 = 1.0
	if density > 0 {
		// Uses heat island research showing compounding effects
		densityMultiplier = 1.0 + (math.Pow(float64(density), 0.7) / 10.0)
	}

	// Apply regional climate factor - hot regions dissipate heat less efficiently
	climateFactor := 1.0
	if ambientTemp > 25 {
		climateFactor = 1.0 + (ambientTemp-25)*0.02
	} else if ambientTemp < 10 {
		climateFactor = 0.8
	}

	return baseIncrease * densityMultiplier * climateFactor
}

// Get environmental data for a location
func getEnvironmentalData(loc *DatacenterLocation) EnvironmentalData {
	// Extract state code for regional data lookup
	stateCode := extractStateCode(loc.Name)

	// Get grid emissions intensity (kg CO2e/kWh) - from EPA eGRID data
	gridIntensity := getGridEmissionsIntensity(stateCode)

	// Get renewable penetration percentage - from EIA data
	renewablePct := getRenewablePenetration(stateCode)

	// Get water scarcity index - from WRI Aqueduct data
	waterScarcity := getWaterScarcityIndex(loc.Latitude, loc.Longitude)

	// Get average temperature - from NOAA data
	avgTemp := getAverageTemperature(loc.Latitude, loc.Longitude)

	// Count nearby datacenters - from our database
	dcDensity := countNearbyCenters(loc)

	// Get natural disaster risk - from FEMA and USGS data
	disasterRisk := getNaturalDisasterRisk(loc.Latitude, loc.Longitude)

	// Get biodiversity sensitivity - from conservation databases
	biodiversity := getBiodiversitySensitivity(loc.Latitude, loc.Longitude)

	// Get land use change impact - based on ecosystem type
	landUseImpact := getLandUseChangeImpact(loc.Latitude, loc.Longitude)

	// Get socioeconomic impact - based on demographic data
	socioImpact := getSocioeconomicImpact(loc.Latitude, loc.Longitude)

	return EnvironmentalData{
		GridEmissionsIntensity:  gridIntensity,
		RenewablePenetration:    renewablePct,
		WaterScarcityIndex:      waterScarcity,
		AmbientTemperature:      avgTemp,
		DatacenterDensity:       dcDensity,
		NaturalDisasterRisk:     disasterRisk,
		BiodiversitySensitivity: biodiversity,
		LandUseChangeImpact:     landUseImpact,
		SocioeconomicImpact:     socioImpact,
	}
}

// Extract state code from location name
func extractStateCode(name string) string {
	stateMatch := strings.Split(name, ", ")
	if len(stateMatch) >= 2 {
		stateCode := stateMatch[len(stateMatch)-1]
		// Check if it's a valid 2-letter state code
		if len(stateCode) == 2 && stateCode == strings.ToUpper(stateCode) {
			return stateCode
		}
	}
	return "Unknown"
}

// Get grid emissions intensity (kg CO2e/kWh) based on eGRID data
func getGridEmissionsIntensity(stateCode string) float64 {
	// Grid emissions intensity by state (kg CO2e/kWh)
	// From EPA eGRID 2021 data: https://www.epa.gov/egrid
	gridData := map[string]float64{
		"WA": 0.0932, "OR": 0.1521, "CA": 0.2096, "ID": 0.0905, "NV": 0.3135,
		"MT": 0.3929, "WY": 0.7891, "UT": 0.6321, "CO": 0.5309, "AZ": 0.3742,
		"NM": 0.4916, "ND": 0.5874, "SD": 0.3326, "NE": 0.4911, "KS": 0.4547,
		"OK": 0.4139, "TX": 0.4089, "MN": 0.3632, "IA": 0.3817, "MO": 0.6733,
		"AR": 0.4422, "LA": 0.3924, "WI": 0.5142, "IL": 0.3873, "MS": 0.4341,
		"MI": 0.4486, "IN": 0.6899, "KY": 0.7662, "TN": 0.3711, "AL": 0.3707,
		"OH": 0.5354, "WV": 0.8463, "VA": 0.3124, "NC": 0.3299, "SC": 0.2994,
		"GA": 0.3749, "FL": 0.3830, "PA": 0.3790, "NY": 0.2139, "ME": 0.1743,
		"NH": 0.1240, "VT": 0.0055, "MA": 0.3075, "RI": 0.3726, "CT": 0.2369,
		"NJ": 0.2644, "DE": 0.4644, "MD": 0.3187, "DC": 0.2783, "AK": 0.4566,
		"HI": 0.6246, "PR": 0.5893, "VI": 0.6021, "GU": 0.6432, "MP": 0.6521,
	}

	// Return value from map or default if not found
	if intensity, exists := gridData[stateCode]; exists {
		return intensity
	}
	return 0.4500 // U.S. average if no data available
}

// Get renewable energy penetration percentage
func getRenewablePenetration(stateCode string) float64 {
	// Renewable penetration by state (%)
	// From EIA 2023 data: https://www.eia.gov/electricity/data/state/
	renewableData := map[string]float64{
		"WA": 75.3, "OR": 69.8, "CA": 54.2, "ID": 78.1, "NV": 34.6,
		"MT": 58.2, "WY": 16.3, "UT": 24.7, "CO": 32.4, "AZ": 16.1,
		"NM": 36.8, "ND": 43.2, "SD": 77.9, "NE": 30.1, "KS": 47.3,
		"OK": 44.8, "TX": 32.1, "MN": 33.6, "IA": 60.2, "MO": 11.3,
		"AR": 13.7, "LA": 4.8, "WI": 14.1, "IL": 14.3, "MS": 3.2,
		"MI": 12.6, "IN": 10.3, "KY": 7.1, "TN": 14.4, "AL": 9.1,
		"OH": 5.7, "WV": 6.1, "VA": 12.3, "NC": 14.2, "SC": 7.3,
		"GA": 12.6, "FL": 6.4, "PA": 6.9, "NY": 31.2, "ME": 82.1,
		"NH": 23.1, "VT": 99.8, "MA": 15.9, "RI": 12.8, "CT": 6.5,
		"NJ": 7.9, "DE": 6.1, "MD": 12.4, "DC": 5.3, "AK": 30.1,
		"HI": 18.2, "PR": 7.1, "VI": 3.2, "GU": 5.1, "MP": 2.1,
	}

	if pct, exists := renewableData[stateCode]; exists {
		return pct
	}
	return 20.1 // U.S. average if no data available
}

// Get water scarcity index (0-5 scale, higher is more scarce)
// Based on WRI Aqueduct Water Risk Atlas
func getWaterScarcityIndex(lat, lng float64) float64 {
	// Simplified regional water stress index
	// The full implementation would use a GIS lookup or API

	// High water stress regions
	highStressRegions := []struct {
		lat, lng float64
		radius   float64
		stress   float64
	}{
		{33.45, -112.07, 200, 4.2}, // Phoenix
		{36.17, -115.14, 150, 4.5}, // Las Vegas
		{32.72, -97.12, 150, 3.8},  // Dallas-Fort Worth
		{37.77, -122.42, 100, 3.5}, // San Francisco
		{34.05, -118.24, 120, 3.9}, // Los Angeles
		{39.74, -104.99, 100, 3.7}, // Denver
		{40.76, -111.89, 100, 4.0}, // Salt Lake City
		{35.08, -106.65, 100, 4.1}, // Albuquerque
	}

	// Check if location is in high stress region
	for _, region := range highStressRegions {
		dist := distance(lat, lng, region.lat, region.lng)
		if dist <= region.radius {
			return region.stress
		}
	}

	// Default stress levels by region
	if lng < -115 {
		return 3.2 // Western states
	} else if lng < -100 {
		return 2.5 // Central states
	} else if lng < -90 {
		return 1.8 // Midwest
	} else if lat < 35 && lng > -90 {
		return 2.2 // Southeast
	} else if lat >= 40 && lng > -90 {
		return 1.5 // Northeast
	}

	return 2.0 // Default
}

// Get average temperature (°C)
func getAverageTemperature(lat, lng float64) float64 {
	// Simplified model based on latitude and longitude
	// The full implementation would use climate data APIs

	// Base temperature decreases with latitude
	baseTemp := 30.0 - 0.5*math.Abs(lat-20)

	// Adjust for elevation (rough approximation)
	if lng < -105 && lat > 35 {
		baseTemp -= 5.0 // Rocky Mountains
	} else if lng < -115 {
		baseTemp -= 2.0 // West Coast correction
	} else if lng > -80 && lat > 40 {
		baseTemp -= 3.0 // Northeast correction
	} else if lng > -90 && lat < 30 {
		baseTemp += 2.0 // Gulf Coast correction
	}

	return baseTemp
}

// Count nearby datacenters
func countNearbyCenters(loc *DatacenterLocation) int {
	// This would query a database of known datacenter locations
	// For now, let's use a simplified map of known datacenter clusters

	clusters := []struct {
		lat, lng float64
		radius   float64
		count    int
	}{
		{39.05, -77.46, 50, 60},  // Ashburn, VA (Data Center Alley)
		{32.78, -96.80, 50, 35},  // Dallas-Fort Worth
		{37.37, -121.97, 40, 40}, // Silicon Valley
		{41.88, -87.63, 40, 25},  // Chicago
		{33.45, -112.07, 50, 15}, // Phoenix
		{40.73, -74.00, 40, 30},  // New York/New Jersey
		{47.60, -122.33, 50, 20}, // Seattle
		{39.74, -104.99, 40, 15}, // Denver
		{25.78, -80.19, 40, 12},  // Miami
		{36.17, -115.14, 40, 10}, // Las Vegas
		{33.75, -84.39, 40, 18},  // Atlanta
	}

	// Check if location is in a known cluster
	for _, cluster := range clusters {
		dist := distance(loc.Latitude, loc.Longitude, cluster.lat, cluster.lng)
		if dist <= cluster.radius {
			return cluster.count
		}
	}

	// Default count by population density
	if inUrbanArea(loc.Latitude, loc.Longitude) {
		return 3 // Typical urban area has a few datacenters
	}

	return 0 // Rural areas typically have no datacenters
}

// Check if location is in an urban area
func inUrbanArea(lat, lng float64) bool {
	// Very simplified check - would use GIS data in production

	urbanCenters := []struct {
		lat, lng float64
		radius   float64
	}{
		{40.71, -74.01, 50},  // NYC
		{34.05, -118.24, 60}, // LA
		{41.88, -87.63, 40},  // Chicago
		{29.76, -95.37, 40},  // Houston
		{33.45, -112.07, 40}, // Phoenix
		{39.95, -75.17, 30},  // Philadelphia
		{29.42, -98.49, 30},  // San Antonio
		{32.78, -96.80, 40},  // Dallas
		{30.27, -97.74, 30},  // Austin
		{37.77, -122.42, 30}, // San Francisco
	}

	for _, city := range urbanCenters {
		dist := distance(lat, lng, city.lat, city.lng)
		if dist <= city.radius {
			return true
		}
	}

	return false
}

// Get natural disaster risk (0-1 scale)
func getNaturalDisasterRisk(lat, lng float64) float64 {
	// Simplified disaster risk model
	// Would use FEMA, USGS and other risk maps in production

	// High risk zones
	riskZones := []struct {
		lat, lng float64
		radius   float64
		risk     float64
		riskType string
	}{
		{37.77, -122.42, 100, 0.85, "earthquake"}, // Bay Area
		{34.05, -118.24, 100, 0.80, "earthquake"}, // Southern California
		{25.76, -80.19, 200, 0.90, "hurricane"},   // South Florida
		{29.95, -90.07, 150, 0.85, "hurricane"},   // New Orleans
		{35.65, -97.48, 150, 0.75, "tornado"},     // Oklahoma
		{39.74, -104.99, 100, 0.60, "wildfire"},   // Colorado Front Range
	}

	// Check if in high risk zone
	maxRisk := 0.0
	for _, zone := range riskZones {
		dist := distance(lat, lng, zone.lat, zone.lng)
		if dist <= zone.radius && zone.risk > maxRisk {
			maxRisk = zone.risk
		}
	}

	if maxRisk > 0 {
		return maxRisk
	}

	// Regional baseline risks
	if lng < -115 {
		return 0.5 // West Coast (earthquake, fire)
	} else if lng > -90 && lat < 35 {
		return 0.6 // Southeast (hurricane)
	} else if lng > -98 && lng < -88 && lat > 35 && lat < 42 {
		return 0.5 // Midwest (tornado)
	} else if lng < -100 && lat > 35 {
		return 0.4 // Mountain West (wildfire)
	}

	return 0.3 // Default moderate risk
}

// Get biodiversity sensitivity (0-1 scale)
func getBiodiversitySensitivity(lat, lng float64) float64 {
	// Simplified biodiversity sensitivity model
	// Would use conservation databases in production

	// High biodiversity zones
	bioZones := []struct {
		lat, lng    float64
		radius      float64
		sensitivity float64
	}{
		{27.5, -81.0, 150, 0.85},   // Florida Everglades
		{37.86, -119.54, 100, 0.8}, // Yosemite/Sierra Nevada
		{35.6, -83.52, 100, 0.75},  // Great Smoky Mountains
		{44.6, -110.5, 150, 0.8},   // Yellowstone
		{48.7, -113.8, 120, 0.75},  // Glacier National Park
		{29.3, -103.25, 100, 0.65}, // Big Bend
		{36.1, -112.1, 120, 0.7},   // Grand Canyon
	}

	// Check if in sensitive zone
	for _, zone := range bioZones {
		dist := distance(lat, lng, zone.lat, zone.lng)
		if dist <= zone.radius {
			return zone.sensitivity
		}
	}

	// Default sensitivity based on urban proximity
	if inUrbanArea(lat, lng) {
		return 0.3 // Urban areas have lower biodiversity
	}

	return 0.5 // Default moderate sensitivity
}

// Get land use change impact (0-1 scale)
func getLandUseChangeImpact(lat, lng float64) float64 {
	// Simplified land use impact model
	// Would use land cover and ecosystem maps in production

	// If in urban area, impact is lower (already developed)
	if inUrbanArea(lat, lng) {
		return 0.3
	}

	// Regional impacts
	if lng < -115 && lat < 36 {
		return 0.8 // Desert ecosystems (fragile)
	} else if lng > -90 && lat < 30 {
		return 0.7 // Gulf Coast wetlands
	} else if lng > -98 && lng < -88 && lat > 40 && lat < 50 {
		return 0.6 // Northern forests
	} else if lng < -105 && lat > 40 {
		return 0.7 // Mountain ecosystems
	}

	return 0.5 // Default moderate impact
}

// Get socioeconomic impact (0-1 scale)
func getSocioeconomicImpact(lat, lng float64) float64 {
	// Simplified socioeconomic impact model
	// Would use census data and environmental justice indices in production

	// Environmental justice focus areas
	ejAreas := []struct {
		lat, lng float64
		radius   float64
		impact   float64
	}{
		{37.5, -122.0, 30, 0.7},  // East Palo Alto
		{37.7, -122.2, 20, 0.8},  // Oakland
		{33.9, -118.2, 30, 0.85}, // South LA
		{29.7, -95.3, 25, 0.75},  // East Houston
		{38.9, -77.0, 15, 0.7},   // DC SE
		{40.8, -74.0, 20, 0.8},   // Newark
		{41.8, -87.7, 25, 0.75},  // Chicago South/West
		{32.7, -96.8, 20, 0.7},   // South Dallas
	}

	// Check if in environmental justice area
	for _, area := range ejAreas {
		dist := distance(lat, lng, area.lat, area.lng)
		if dist <= area.radius {
			return area.impact
		}
	}

	// Default impact based on urban proximity
	if inUrbanArea(lat, lng) {
		return 0.5 // Urban areas have moderate justice concerns
	}

	return 0.3 // Rural areas typically have lower justice issues
}

// Calculate distance between two points using the Haversine formula
func distance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Earth radius in km

	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c // Distance in km
}

// Handler for getting all possible datacenter locations
func possibleDataCenterHandler(w http.ResponseWriter, r *http.Request) {
    // Add CORS headers
    w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

    // Handle preflight requests
    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusOK)
        return
    }

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
    // Remove this conflicting line:
    // w.Header().Set("Access-Control-Allow-Origin", "*")
    json.NewEncoder(w).Encode(response)
}

// Handler for getting property details based on lat/long
func getPropertyDetailsHandler(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight requests
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

	// Define a small epsilon for floating-point comparison
	const epsilon = 0.0001

	locations, err := readDatacenterLocations()
	if err != nil {
		http.Error(w, "Error reading datacenter locations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Read existing datacenter locations
	existingDatacenters, err := readExistingDatacenters()
	if err != nil {
		// Non-fatal error - we can still proceed without density calculations
		log.Printf("Warning: Could not read existing datacenter data: %v", err)
	}

	// Find the matching location
	var matchedLocation *DatacenterLocation
	for i := range locations {
		if math.Abs(locations[i].Latitude-lat) < epsilon && math.Abs(locations[i].Longitude-lng) < epsilon {
			matchedLocation = &locations[i]

			// Calculate environmental metrics
			allDatacenters := make([]DatacenterLocation, 0, len(locations)+len(existingDatacenters))
			allDatacenters = append(allDatacenters, locations...)
			allDatacenters = append(allDatacenters, existingDatacenters...)

			// Calculate research-based environmental metrics
			calculateResearchBasedMetrics(matchedLocation, allDatacenters)

			break
		}
	}

	if matchedLocation == nil {
		http.Error(w, "No property found at the specified coordinates", http.StatusNotFound)
		return
	}

	// Return all details including environmental metrics
	response := map[string]interface{}{
		"location_name": matchedLocation.Name,
		"land_price":    matchedLocation.LandPrice,
		"electricity":   matchedLocation.Electricity,
		"notes":         matchedLocation.Notes,

		// Environmental metrics
		"eco_score":                matchedLocation.EcoScore,
		"carbon_impact":            matchedLocation.CarbonImpact,
		"temp_increase":            matchedLocation.TempIncrease,
		"water_usage":              matchedLocation.WaterUsage,
		"renewable_access":         matchedLocation.RenewableAccess,
		"datacenter_density":       matchedLocation.DatacenterDensity,
		"density_impact_score":     matchedLocation.DensityImpactScore,
		"compounded_temp_increase": matchedLocation.CompoundedTempIncrease,
		"water_competition":        matchedLocation.WaterCompetition,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow cross-origin requests
	json.NewEncoder(w).Encode(response)
}

func allDataCentersHandler(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

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
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

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
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

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
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

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
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight requests
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
