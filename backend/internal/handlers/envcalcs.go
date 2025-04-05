package handlers

import (
	"math"

	"github.com/Samhith-k/data-center-ecology-map/backend/internal/data"
)

// CalculateResearchBasedMetrics applies your research-based env. calculations
func CalculateResearchBasedMetrics(loc *data.DatacenterLocation, allDatacenters []data.DatacenterLocation) {
	envData := data.GetEnvironmentalData(loc)

	// 1. Calculate PUE (Power Usage Effectiveness) based on climate
	pue := calculateLocationBasedPUE(envData.AmbientTemperature, envData.DatacenterDensity)

	// 2. Some constants
	const (
		itLoadMW        = 15.0
		hoursPerYear    = 8760.0
		waterUseLPerKWh = 1.8
		landUseHectares = 12.0
	)

	// 3. Calculate total energy usage (MWh/year)
	totalEnergyMWh := itLoadMW * pue * hoursPerYear

	// 4. Calculate carbon emissions using regional grid intensity (kg CO2e/year)
	carbonEmissions := totalEnergyMWh * 1000 * envData.GridEmissionsIntensity

	// 5. Calculate water consumption
	waterConsumption := totalEnergyMWh * 1000 * waterUseLPerKWh
	waterImpact := waterConsumption * envData.WaterScarcityIndex

	// 6. Temperature impact
	heatRejection := itLoadMW * (1.0 - (1.0 / pue)) * 3.412
	tempImpact := calculateTemperatureImpact(heatRejection, envData.DatacenterDensity, envData.AmbientTemperature)

	// 7. Land use impact
	landImpact := landUseHectares * envData.LandUseChangeImpact * envData.BiodiversitySensitivity

	// 8. Overall Eco Score
	ecoScore := calcEcoScore(carbonEmissions, waterImpact, tempImpact, landImpact, envData.SocioeconomicImpact)
	if ecoScore < 1 {
		ecoScore = 1
	} else if ecoScore > 100 {
		ecoScore = 100
	}

	// Assign values
	loc.EcoScore = int(ecoScore)
	loc.CarbonImpact = carbonEmissions / 1000 // metric tons
	loc.TempIncrease = tempImpact
	loc.WaterUsage = waterConsumption / 3.785 // gallons
	loc.DatacenterDensity = envData.DatacenterDensity
	loc.RenewableAccess = int(envData.RenewablePenetration)

	// Calculate compound effects
	if envData.DatacenterDensity > 0 {
		densityFactor := math.Log1p(float64(envData.DatacenterDensity)) / math.Log1p(10.0)
		loc.CompoundedTempIncrease = tempImpact * (1.0 + densityFactor)
		loc.WaterCompetition = 1.0 + densityFactor
		loc.WaterUsage *= loc.WaterCompetition
	} else {
		loc.CompoundedTempIncrease = tempImpact
		loc.WaterCompetition = 1.0
	}

	// Density impact score
	if envData.DatacenterDensity == 0 {
		loc.DensityImpactScore = 0
	} else {
		loc.DensityImpactScore = int(math.Min(100, 20*math.Log1p(float64(envData.DatacenterDensity))))
	}
}

func calculateLocationBasedPUE(averageTemp float64, density int) float64 {
	var basePUE float64

	if averageTemp < 10 {
		basePUE = 1.15 + (averageTemp+10)*0.005
	} else if averageTemp < 18 {
		basePUE = 1.2 + (averageTemp-10)*0.01
	} else if averageTemp < 24 {
		basePUE = 1.3 + (averageTemp-18)*0.025
	} else {
		basePUE = 1.45 + (averageTemp-24)*0.04
	}

	if density > 0 {
		densityEffect := 0.01 * math.Min(0.5, math.Log10(float64(density))/2)
		basePUE += densityEffect
	}
	return basePUE
}

func calculateTemperatureImpact(heatRejection float64, density int, ambientTemp float64) float64 {
	baseIncrease := 0.02 * heatRejection

	densityMultiplier := 1.0
	if density > 0 {
		densityMultiplier = 1.0 + (math.Pow(float64(density), 0.7) / 10.0)
	}

	climateFactor := 1.0
	if ambientTemp > 25 {
		climateFactor = 1.0 + (ambientTemp-25)*0.02
	} else if ambientTemp < 10 {
		climateFactor = 0.8
	}

	return baseIncrease * densityMultiplier * climateFactor
}

func calcEcoScore(carbonEmissions, waterImpact, tempImpact, landImpact, socioImpact float64) float64 {
	const (
		carbonNorm = 5000000.0
		waterNorm  = 50000000.0
		tempNorm   = 2.0
		landNorm   = 10.0

		weightCarbon = 0.40
		weightWater  = 0.25
		weightTemp   = 0.20
		weightLand   = 0.10
		weightSocial = 0.05
	)

	normCarbon := carbonEmissions / carbonNorm
	normWater := waterImpact / waterNorm
	normTemp := tempImpact / tempNorm
	normLand := landImpact / landNorm

	envImpact := (normCarbon * weightCarbon) +
		(normWater * weightWater) +
		(normTemp * weightTemp) +
		(normLand * weightLand) +
		(socioImpact * weightSocial)

	return 100 - (envImpact * 100)
}
