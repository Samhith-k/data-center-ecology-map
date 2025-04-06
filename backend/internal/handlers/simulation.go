package handlers

import (
	"encoding/json"
	"math"
	"net/http"
	"strings"

	"github.com/Samhith-k/data-center-ecology-map/backend/internal/cart"
	"github.com/Samhith-k/data-center-ecology-map/backend/internal/data"
)

// Minimal constants for the demonstration.
const (
	thresholdSurvivability = 10.0
	startYear              = 2025
	endYear                = 2100
)

// ClimateProjection represents one data point in our simulation.
type ClimateProjection struct {
	Year                   int     `json:"year"`
	BaselineTemperature    float64 `json:"baseline_temperature"`
	DataCenterContribution float64 `json:"data_center_contribution"` // extra °C
	TotalTemperature       float64 `json:"total_temperature"`        // °C
	FossilFuelReserves     float64 `json:"fossil_fuel_reserves"`     // 1.0 -> 0
	Survivability          int     `json:"survivability"`            // 0-100 scale
}

// SimulationResponse is the overall response from the simulation endpoint.
type SimulationResponse struct {
	Username               string              `json:"username"`
	Data                   []ClimateProjection `json:"data"`
	TotalTimeToEnd         int                 `json:"total_time_to_end"`
	TimeDatacentersRemoved int                 `json:"time_datacenters_removed"`
}

// GetUserClimateSimulationHandler handles GET /api/simulation?username=alice
func GetUserClimateSimulationHandler(w http.ResponseWriter, r *http.Request) {
	addCORSHeaders(w)
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

	// 1. Get or create a user cart
	userCart, ok := cart.GetCart(username)
	if !ok {
		// If not found, use an empty cart
		userCart = &cart.Cart{
			Username:  username,
			Items:     []data.DatacenterLocation{},
			MoneyLeft: 1000000,
		}
	}

	// 2. Calculate total data center contribution
	dataCenterContribution := calcDataCenterContribution(userCart.Items)

	var (
		projections     []ClimateProjection
		projectionsNoDC []ClimateProjection
		totalTimeToEnd  int
		totalTimeNoDC   int
	)

	for year := startYear; year <= endYear; year++ {
		// Baseline from 1.2°C (2025) to 3.7°C (2100)
		baselineTemp := getBaselineTemperature(year)

		// Example: fossil fuel fraction from 1.0 down to 0.2
		fossilRes := getFossilFuelFraction(year)

		totalTemp := baselineTemp + dataCenterContribution
		surv := calcSurvivability(totalTemp, fossilRes)

		proj := ClimateProjection{
			Year:                   year,
			BaselineTemperature:    baselineTemp,
			DataCenterContribution: dataCenterContribution,
			TotalTemperature:       totalTemp,
			FossilFuelReserves:     fossilRes,
			Survivability:          int(math.Round(surv)),
		}
		projections = append(projections, proj)

		// Compare scenario with zero data center usage
		noDCtemp := baselineTemp
		noDCsurv := calcSurvivability(noDCtemp, fossilRes)
		projNoDC := ClimateProjection{
			Year:                   year,
			BaselineTemperature:    baselineTemp,
			DataCenterContribution: 0,
			TotalTemperature:       noDCtemp,
			FossilFuelReserves:     fossilRes,
			Survivability:          int(math.Round(noDCsurv)),
		}
		projectionsNoDC = append(projectionsNoDC, projNoDC)

		// Threshold crossing checks
		if totalTimeToEnd == 0 && surv <= thresholdSurvivability {
			totalTimeToEnd = year - startYear
		}
		if totalTimeNoDC == 0 && noDCsurv <= thresholdSurvivability {
			totalTimeNoDC = year - startYear
		}
	}

	if totalTimeToEnd == 0 {
		totalTimeToEnd = endYear - startYear
	}
	if totalTimeNoDC == 0 {
		totalTimeNoDC = endYear - startYear
	}

	resp := SimulationResponse{
		Username:               username,
		Data:                   projections,
		TotalTimeToEnd:         totalTimeToEnd,
		TimeDatacentersRemoved: totalTimeNoDC - totalTimeToEnd,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ----------------------------------------------------------
// Helper functions
// ----------------------------------------------------------

func calcDataCenterContribution(items []data.DatacenterLocation) float64 {
	var total float64
	for _, dc := range items {
		// 1) Identify DC type from name or notes
		dcType := inferDCType(dc.Name, dc.Notes)

		// 2) Identify size from landPrice or notes
		size := inferDCSize(dc.LandPrice)

		// 3) Identify region factor from lat/long
		region := inferRegion(dc.Latitude, dc.Longitude)

		// 4) Calculate final
		emission := dataCenterEmission(dcType, size, region)
		total += emission
	}
	return total
}

// dataCenterEmission returns a small fraction of °C contributed by one DC.
func dataCenterEmission(dcType, size, region string) float64 {
	// Base for Standard: 0.005 °C, HPC: 0.01, Colo: 0.007
	var base float64
	switch dcType {
	case "HPC":
		base = 0.01
	case "Colo":
		base = 0.007
	default:
		base = 0.005
	}

	// Increase if large
	if size == "large" {
		base += 0.003
	}

	// Region factor: "coal" => +0.002, "renewable" => -0.001, "average" => 0
	switch region {
	case "coal":
		base += 0.002
	case "renewable":
		base -= 0.001
	}

	return base
}

// inferDCType checks the name/notes. If it has "hpc", we assume HPC. If "colo", colocation. Else standard.
func inferDCType(name, notes string) string {
	txt := strings.ToLower(name + " " + notes)
	switch {
	case strings.Contains(txt, "hpc"):
		return "HPC"
	case strings.Contains(txt, "colo"):
		return "Colo"
	}
	return "Standard"
}

// inferDCSize checks if the land_price mentions "2.5M", we assume "large"
func inferDCSize(landPrice string) string {
	if strings.Contains(strings.ToLower(landPrice), "2.5m") {
		return "large"
	}
	return "medium"
}

// inferRegion does a naive bounding box. e.g. lat < 30 => "coal", lat > 45 => "renewable", else "average".
func inferRegion(lat, lng float64) string {
	if lat < 30 {
		return "coal"
	} else if lat > 45 {
		return "renewable"
	}
	return "average"
}

// getBaselineTemperature linearly interpolates from 1.2°C in 2025 to 3.7°C in 2100
func getBaselineTemperature(year int) float64 {
	frac := float64(year-startYear) / float64(endYear-startYear)
	return 1.2 + frac*(3.7-1.2)
}

// getFossilFuelFraction linearly decays from 1.0 in 2025 -> 0.2 in 2100
func getFossilFuelFraction(year int) float64 {
	frac := float64(year-startYear) / float64(endYear-startYear)
	return 1.0 - frac*(1.0-0.2)
}

// calcSurvivability example: 100 - temp*20 - (1 - reserves)*40
func calcSurvivability(temp, reserves float64) float64 {
	surv := 100 - (temp * 20) - ((1 - reserves) * 40)
	if surv < 0 {
		surv = 0
	}
	return surv
}
