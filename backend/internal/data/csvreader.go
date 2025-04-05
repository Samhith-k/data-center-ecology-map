package data

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// ReadAllDataCenters reads the “us_datacenters.csv” file
func ReadAllDataCenters(filename string) ([]DataCenter, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var dataCenters []DataCenter
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		dc, parseErr := parseDataCenterLine(line)
		if parseErr != nil {
			continue
		}
		dataCenters = append(dataCenters, *dc)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return dataCenters, nil
}

// parseDataCenterLine splits the line so last two fields are lat/lng
func parseDataCenterLine(line string) (*DataCenter, error) {
	parts := strings.Split(line, ",")
	if len(parts) < 3 {
		return nil, fmt.Errorf("line has fewer than 3 fields")
	}
	longitudeStr := parts[len(parts)-1]
	latitudeStr := parts[len(parts)-2]
	name := strings.Join(parts[:len(parts)-2], ",")

	lat, err := strconv.ParseFloat(latitudeStr, 64)
	if err != nil {
		return nil, err
	}
	lng, err := strconv.ParseFloat(longitudeStr, 64)
	if err != nil {
		return nil, err
	}
	return &DataCenter{Name: name, Latitude: lat, Longitude: lng}, nil
}

// ReadDatacenterLocations reads “us_possible_locations.csv”
func ReadDatacenterLocations(filename string) ([]DatacenterLocation, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	// skip header
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
		if len(record) < 6 {
			continue
		}

		lat, err := strconv.ParseFloat(strings.TrimSpace(record[0]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid latitude value: %s", record[0])
		}
		lng, err := strconv.ParseFloat(strings.TrimSpace(record[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid longitude value: %s", record[1])
		}

		locations = append(locations, DatacenterLocation{
			Latitude:    lat,
			Longitude:   lng,
			Name:        strings.TrimSpace(record[2]),
			LandPrice:   strings.TrimSpace(record[3]),
			Electricity: strings.TrimSpace(record[4]),
			Notes:       parseNotes(strings.TrimSpace(record[5])),
		})
	}
	return locations, nil
}

// ReadExistingDatacenters reads “us_datacenters.csv” again to produce a []DatacenterLocation
func ReadExistingDatacenters(filename string) ([]DatacenterLocation, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open existing datacenters CSV: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var existingDCs []DatacenterLocation

	// skip header
	scanner.Scan()

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

		lat, err := strconv.ParseFloat(latitudeStr, 64)
		if err != nil {
			continue
		}
		lng, err := strconv.ParseFloat(longitudeStr, 64)
		if err != nil {
			continue
		}

		existingDCs = append(existingDCs, DatacenterLocation{
			Name: name, Latitude: lat, Longitude: lng,
		})
	}
	return existingDCs, scanner.Err()
}

// parseNotes (simple pass-through, or do advanced logic)
func parseNotes(notesStr string) string {
	return notesStr
}
