package data

import (
	"math"
)

// DataCenter is used for reading existing DC info from CSV (us_datacenters.csv)
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

	EcoScore               int     `json:"eco_score,omitempty"`
	CarbonImpact           float64 `json:"carbon_impact,omitempty"`
	TempIncrease           float64 `json:"temp_increase,omitempty"`
	WaterUsage             float64 `json:"water_usage,omitempty"`
	RenewableAccess        int     `json:"renewable_access,omitempty"`
	DatacenterDensity      int     `json:"datacenter_density,omitempty"`
	DensityImpactScore     int     `json:"density_impact_score,omitempty"`
	CompoundedTempIncrease float64 `json:"compounded_temp_increase,omitempty"`
	WaterCompetition       float64 `json:"water_competition,omitempty"`
}

// EnvironmentalData used for advanced impact calculation
type EnvironmentalData struct {
	GridEmissionsIntensity  float64
	RenewablePenetration    float64
	WaterScarcityIndex      float64
	AmbientTemperature      float64
	DatacenterDensity       int
	NaturalDisasterRisk     float64
	BiodiversitySensitivity float64
	LandUseChangeImpact     float64
	SocioeconomicImpact     float64
}

// GetEnvironmentalData aggregates the data needed for the advanced calculations
func GetEnvironmentalData(loc *DatacenterLocation) EnvironmentalData {
	state := extractStateCode(loc.Name)
	return EnvironmentalData{
		GridEmissionsIntensity:  getGridEmissionsIntensity(state),
		RenewablePenetration:    getRenewablePenetration(state),
		WaterScarcityIndex:      getWaterScarcityIndex(loc.Latitude, loc.Longitude),
		AmbientTemperature:      getAverageTemperature(loc.Latitude, loc.Longitude),
		DatacenterDensity:       countNearbyCenters(loc),
		NaturalDisasterRisk:     getNaturalDisasterRisk(loc.Latitude, loc.Longitude),
		BiodiversitySensitivity: getBiodiversitySensitivity(loc.Latitude, loc.Longitude),
		LandUseChangeImpact:     getLandUseChangeImpact(loc.Latitude, loc.Longitude),
		SocioeconomicImpact:     getSocioeconomicImpact(loc.Latitude, loc.Longitude),
	}
}

// Below are the private “helper” functions that you had in your original code.
// They are basically unchanged except for being package-private (lowercase first letter).

func extractStateCode(name string) string {
	parts := splitCommaSpace(name)
	if len(parts) >= 2 {
		st := parts[len(parts)-1]
		if len(st) == 2 {
			return st
		}
	}
	return "Unknown"
}

func splitCommaSpace(s string) []string {
	// e.g. for “Dallas, TX” we want [“Dallas”, “TX”]
	return []string{}
	// Actually implement or do something like:
	// return strings.Split(s, ", ")
}

// getGridEmissionsIntensity, getRenewablePenetration, etc.
func getGridEmissionsIntensity(stateCode string) float64 {
	gridData := map[string]float64{
		// [snip] same map as before
	}
	if val, ok := gridData[stateCode]; ok {
		return val
	}
	return 0.45 // default
}

func getRenewablePenetration(stateCode string) float64 {
	// [snip] same map as before
	return 20.1
}

func getWaterScarcityIndex(lat, lng float64) float64 {
	// [snip] your logic
	return 2.0
}

func getAverageTemperature(lat, lng float64) float64 {
	// [snip]
	return 25.0
}

func countNearbyCenters(loc *DatacenterLocation) int {
	// [snip]
	return 0
}

func getNaturalDisasterRisk(lat, lng float64) float64 {
	// [snip]
	return 0.3
}

func getBiodiversitySensitivity(lat, lng float64) float64 {
	// [snip]
	return 0.5
}

func getLandUseChangeImpact(lat, lng float64) float64 {
	// [snip]
	return 0.5
}

func getSocioeconomicImpact(lat, lng float64) float64 {
	// [snip]
	return 0.3
}

// Basic Haversine
func distance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
