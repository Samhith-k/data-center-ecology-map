// src/components/Game.js
import React, { useState, useEffect } from 'react';
import { MapContainer, TileLayer, Marker, Popup, useMap } from 'react-leaflet';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';
import './Game.css';
import ApiService from '../services/api';

// Fix for Leaflet marker icons
delete L.Icon.Default.prototype._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon-2x.png',
  iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png',
});

// Custom marker icons
const availableIcon = new L.DivIcon({
  className: 'custom-map-marker',
  html: `<div class="map-dot blue-dot"></div>`,
  iconSize: [14, 14],
  iconAnchor: [7, 7],
  popupAnchor: [0, -7]
});

const selectedIcon = new L.DivIcon({
  className: 'custom-map-marker',
  html: `<div class="map-dot green-dot"></div>`,
  iconSize: [14, 14],
  iconAnchor: [7, 7],
  popupAnchor: [0, -7]
});

const builtIcon = new L.DivIcon({
  className: 'custom-map-marker',
  html: `<div class="map-dot purple-dot"></div>`,
  iconSize: [14, 14],
  iconAnchor: [7, 7],
  popupAnchor: [0, -7]
});

// New icon for potential data center locations
const potentialLocationIcon = new L.DivIcon({
  className: 'custom-map-marker',
  html: `<div class="map-dot orange-dot"></div>`,
  iconSize: [14, 14],
  iconAnchor: [7, 7],
  popupAnchor: [0, -7]
});

function Game({ username, onLogout }) {
  const [availableLocations, setAvailableLocations] = useState([]);
  const [potentialLocations, setPotentialLocations] = useState([]); // New state for potential locations
  const [selectedLocation, setSelectedLocation] = useState(null);
  const [builtDataCenters, setBuiltDataCenters] = useState([]);
  const [budget, setBudget] = useState(10000000);
  const [score, setScore] = useState(0);
  const [carbonFootprint, setCarbonFootprint] = useState(0);
  const [day, setDay] = useState(1);
  const [isLoading, setIsLoading] = useState(true);
  const [locationLoading, setLocationLoading] = useState(false);
  const [error, setError] = useState(null);
  const [notification, setNotification] = useState(null);

  // Building options with different specifications
  const buildingOptions = [
    {
      id: 1,
      name: 'Standard Data Center',
      cost: 2000000,
      energyEfficiency: 60,
      capacity: 5000,
      carbonImpact: 0.8,
      description: 'Basic facility with standard cooling and power systems.'
    },
    {
      id: 2,
      name: 'Eco Optimized Center',
      cost: 3500000,
      energyEfficiency: 85,
      capacity: 4800,
      carbonImpact: 0.4,
      description: 'Energy-efficient design with improved cooling systems and partial renewable integration.'
    },
    {
      id: 3,
      name: 'Next-Gen Sustainable Facility',
      cost: 5000000,
      energyEfficiency: 95,
      capacity: 5200,
      carbonImpact: 0.1,
      description: 'Cutting-edge facility with advanced liquid cooling, on-site renewables, and intelligent power management.'
    }
  ];

  // Fetch potential data center locations
  const fetchPotentialLocations = async () => {
    try {
      console.log("Starting fetch of potential locations...");
      const response = await fetch('http://localhost:8080/api/possible-datacenters', {
        method: 'GET',
        credentials: 'include',
      });
      
      console.log("Response status:", response.status);
      
      const rawData = await response.json();
      console.log("Potential locations received:", rawData);
      
      // Transform the raw data to the format needed for the map
      const locations = rawData.map((dc, index) => ({
        id: `potential-${index + 1}`,
        position: { 
          lat: dc.latitude, 
          lng: dc.longitude 
        },
        isPotential: true // Flag to identify potential locations
      }));
      
      setPotentialLocations(locations);
    } catch (err) {
      console.error("Error fetching potential locations:", err);
    }
  };

  // Fetch available locations
  useEffect(() => {
    const fetchDataCenters = async () => {
      try {
        setIsLoading(true);
        console.log("Fetching data centers...");
        
        // Get data from the backend
        const response = await fetch('http://localhost:8080/alldatacenters', {
          method: 'GET',
          credentials: 'include',
        });
        
        if (!response.ok) {
          throw new Error(`Failed to fetch data centers: ${response.status}`);
        }
        
        const rawData = await response.json();
        console.log("Data centers received:", rawData);
        
        // Transform the raw data to the format needed for the map
        const locations = rawData.map((dc, index) => ({
          id: index + 1,
          name: dc.name,
          position: { 
            lat: dc.latitude, 
            lng: dc.longitude 
          }
        }));
        
        setAvailableLocations(locations);
        setIsLoading(false);
      } catch (err) {
        console.error("Error fetching data centers:", err);
        setError('Failed to load data center locations. Please try again later.');
        setIsLoading(false);
        
        // Use fallback data if API call fails
        setAvailableLocations([
          {
            id: 1,
            name: "Northern Virginia",
            position: {"lat": 38.8, "lng": -77.2}
          },
          {
            id: 2,
            name: "Oregon",
            position: {"lat": 45.5, "lng": -122.5}
          },
          {
            id: 3,
            name: "Iceland",
            position: {"lat": 64.1, "lng": -21.9}
          },
          {
            id: 4,
            name: "Singapore",
            position: {"lat": 1.3, "lng": 103.8}
          },
          {
            id: 5,
            name: "Northern Sweden",
            position: {"lat": 65.6, "lng": 22.1}
          }
        ]);
      }
    };

    fetchDataCenters();
    fetchPotentialLocations(); // Fetch potential locations
  }, []);

  // Calculate total environmental score for a location
  const calculateLocationScore = (location) => {
    if (!location || !location.climate) return 0;
    return (location.climate + location.renewable + location.grid + location.risk) / 4;
  };
  
  // Handle location selection on the map
  const handleLocationSelect = async (location) => {
    try {
      setLocationLoading(true);
      
      if (location.isPotential) {
        // This is a potential location, fetch details from property-details endpoint
        console.log("Fetching property details for:", location);
        const response = await fetch(
          `http://localhost:8080/api/property-details?lat=${location.position.lat}&lng=${location.position.lng}`,
          {
            method: 'GET',
            credentials: 'include',
          }
        );
        
        if (!response.ok) {
          throw new Error(`Failed to fetch property details: ${response.status}`);
        }
        
        const propertyData = await response.json();
        console.log("Property details:", propertyData);
        
        // Parse land price to extract numeric value
        let landCost = 3000000; // Default fallback
        try {
          const priceText = propertyData.land_price || "";
          const priceMatch = priceText.match(/\$([0-9,]+)/);
          if (priceMatch && priceMatch[1]) {
            landCost = parseInt(priceMatch[1].replace(/,/g, ''));
          }
        } catch (e) {
          console.error("Error parsing land price:", e);
        }
        
        // Create details object from the API response
        const details = {
          name: propertyData.location_name || "Potential Location",
          climate: Math.floor(Math.random() * 30) + 60, // Random climate score
          renewable: Math.floor(Math.random() * 40) + 40, // Random renewable score
          grid: Math.floor(Math.random() * 40) + 40, // Random grid score
          risk: Math.floor(Math.random() * 20) + 70, // Random risk score
          land_cost: landCost,
          electricity_cost: propertyData.electricity || "Unknown",
          description: propertyData.notes || "A potential location for a new data center."
        };
        
        // Create enriched location object
        const enrichedLocation = {
          ...location,
          ...details
        };
        
        setSelectedLocation(enrichedLocation);
      } else {
        // Regular existing location, generate metrics as before
        const details = {
          climate: Math.floor(Math.random() * 30) + 60, // Random value between 60-90
          renewable: Math.floor(Math.random() * 40) + 40, // Random value between 40-80
          grid: Math.floor(Math.random() * 40) + 40, // Random value between 40-80
          risk: Math.floor(Math.random() * 20) + 70, // Random value between 70-90
          land_cost: Math.floor(Math.random() * 3000000) + 2000000, // Random cost
          description: `Data center located in ${location.name} with excellent connectivity to major networks.`
        };
        
        // Create a complete location object with the generated details
        const enrichedLocation = {
          ...location,
          ...details
        };
        
        setSelectedLocation(enrichedLocation);
      }
      
      setNotification(null);
      setLocationLoading(false);
    } catch (error) {
      console.error("Error generating location details:", error);
      setError("Failed to load location details");
      setLocationLoading(false);
    }
  };
  
  // Handle building construction
  const handleBuild = (building) => {
    if (budget >= building.cost + selectedLocation.land_cost) {
      // Calculate environmental impact
      const locationScore = calculateLocationScore(selectedLocation);
      const environmentalImpact = building.carbonImpact * (100 - locationScore) / 100;
      
      // Create new data center object
      const newDataCenter = {
        id: Date.now(),
        location: selectedLocation,
        building: building,
        dayBuilt: day,
        score: locationScore * building.energyEfficiency / 10
      };
      
      // Update game state
      setBudget(prev => prev - (building.cost + selectedLocation.land_cost));
      setCarbonFootprint(prev => prev + environmentalImpact);
      setScore(prev => prev + newDataCenter.score);
      setDay(prev => prev + 30);
      setBuiltDataCenters(prev => [...prev, newDataCenter]);
      
      // Show notification
      setNotification({
        type: 'success',
        message: `Successfully built ${building.name} in ${selectedLocation.name}!`
      });
      
      // Reset selection after building
      setSelectedLocation(null);
    } else {
      setNotification({
        type: 'error',
        message: "Insufficient funds for this construction!"
      });
    }
  };

  // Close notification
  const closeNotification = () => {
    setNotification(null);
  };

  if (isLoading) {
    return (
      <div className="game-loading">
        <div className="loader"></div>
        <p>Loading Data Center Tycoon...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="game-error">
        <h2>Error</h2>
        <p>{error}</p>
        <button onClick={() => window.location.reload()}>Try Again</button>
      </div>
    );
  }

  return (
    <div className="game-container">
      {/* Game Header */}
      <div className="game-header">
        <div className="game-title">Data Center Tycoon</div>
        
        {/* Add the user info and logout button here */}
        <div className="user-controls">
          <span className="username">Welcome, {username}!</span>
          <button onClick={onLogout} className="logout-button">
            Logout
          </button>
        </div>
        
        <div className="game-stats">
          <div className="stat">
            <span className="stat-icon">💰</span>
            <span className="stat-label">Budget:</span>
            <span className="stat-value">${budget.toLocaleString()}</span>
          </div>
          
          <div className="stat">
            <span className="stat-icon">📊</span>
            <span className="stat-label">Score:</span>
            <span className="stat-value">{Math.round(score)}</span>
          </div>
          
          <div className="stat">
            <span className="stat-icon">🏭</span>
            <span className="stat-label">Carbon:</span>
            <span className="stat-value">{carbonFootprint.toFixed(1)} MT</span>
          </div>
          
          <div className="stat">
            <span className="stat-icon">📅</span>
            <span className="stat-label">Day:</span>
            <span className="stat-value">{day}</span>
          </div>
        </div>
      </div>
      
      {/* Main Game Area */}
      <div className="game-content">
        {/* World Map */}
        <div className="world-map">
          <h2>Global Data Center Map</h2>
          <MapContainer center={[20, 0]} zoom={2} style={{ height: "500px", width: "100%" }}>
            <TileLayer
              url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
              attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
            />
            
            {/* Available Locations */}
            {availableLocations.map(location => {
              // Skip if a data center is already built here
              const isBuilt = builtDataCenters.some(dc => dc.location.id === location.id);
              const isSelected = selectedLocation && selectedLocation.id === location.id;
              
              if (!isBuilt) {
                return (
                  <Marker 
                    key={location.id} 
                    position={[location.position.lat, location.position.lng]}
                    icon={isSelected ? selectedIcon : availableIcon}
                    eventHandlers={{
                      click: () => handleLocationSelect(location)
                    }}
                  >
                    <Popup>
                      <div>
                        <h3>{location.name}</h3>
                        <button 
                          className="map-select-btn"
                          onClick={() => handleLocationSelect(location)}
                        >
                          View Details
                        </button>
                      </div>
                    </Popup>
                  </Marker>
                );
              }
              return null;
            })}
            
            {/* Potential Locations */}
            {potentialLocations.map(location => {
              const isSelected = selectedLocation && selectedLocation.id === location.id;
              
              return (
                <Marker 
                  key={location.id} 
                  position={[location.position.lat, location.position.lng]}
                  icon={isSelected ? selectedIcon : potentialLocationIcon}
                  eventHandlers={{
                    click: () => handleLocationSelect(location)
                  }}
                >
                  <Popup>
                    <div>
                      <h3>Potential Location</h3>
                      <button 
                        className="map-select-btn"
                        onClick={() => handleLocationSelect(location)}
                      >
                        View Details
                      </button>
                    </div>
                  </Popup>
                </Marker>
              );
            })}
            
            {/* Built Data Centers */}
            {builtDataCenters.map(dataCenter => (
              <Marker 
                key={dataCenter.id} 
                position={[dataCenter.location.position.lat, dataCenter.location.position.lng]}
                icon={builtIcon}
              >
                <Popup>
                  <div>
                    <h3>{dataCenter.location.name}</h3>
                    <p>Facility: {dataCenter.building.name}</p>
                    <p>Built on day: {dataCenter.dayBuilt}</p>
                    <p>Efficiency: {dataCenter.building.energyEfficiency}%</p>
                    <p>Carbon Impact: {(dataCenter.building.carbonImpact * (100 - calculateLocationScore(dataCenter.location)) / 100).toFixed(2)} MT/day</p>
                  </div>
                </Popup>
              </Marker>
            ))}
          </MapContainer>
        </div>
        
        {/* Info Panel */}
        <div className="info-panel">
          {locationLoading ? (
            <div className="loading-details">
              <div className="loader"></div>
              <p>Loading location details...</p>
            </div>
          ) : selectedLocation ? (
            <>
              <h2>{selectedLocation.name}</h2>
              
              {/* Environment Metrics */}
              <div className="metrics-container">
                <h3>Environmental Analysis</h3>
                
                <div className="metric">
                  <div className="metric-label">
                    <span className="metric-icon climate">🌡️</span>
                    <span>Climate Suitability</span>
                  </div>
                  <div className="metric-bar">
                    <div className="metric-fill" style={{width: `${selectedLocation.climate}%`}}></div>
                  </div>
                  <span className="metric-value">{selectedLocation.climate}%</span>
                </div>
                
                <div className="metric">
                  <div className="metric-label">
                    <span className="metric-icon renewable">🔆</span>
                    <span>Renewable Potential</span>
                  </div>
                  <div className="metric-bar">
                    <div className="metric-fill" style={{width: `${selectedLocation.renewable}%`}}></div>
                  </div>
                  <span className="metric-value">{selectedLocation.renewable}%</span>
                </div>
                
                <div className="metric">
                  <div className="metric-label">
                    <span className="metric-icon grid">⚡</span>
                    <span>Grid Cleanliness</span>
                  </div>
                  <div className="metric-bar">
                    <div className="metric-fill" style={{width: `${selectedLocation.grid}%`}}></div>
                  </div>
                  <span className="metric-value">{selectedLocation.grid}%</span>
                </div>
                
                <div className="metric">
                  <div className="metric-label">
                    <span className="metric-icon risk">⚠️</span>
                    <span>Disaster Safety</span>
                  </div>
                  <div className="metric-bar">
                    <div className="metric-fill" style={{width: `${selectedLocation.risk}%`}}></div>
                  </div>
                  <span className="metric-value">{selectedLocation.risk}%</span>
                </div>
                
                <div className="metric-summary">
                  <div>
                    <span>Overall Rating:</span>
                    <span className="score">{Math.round(calculateLocationScore(selectedLocation))}/100</span>
                  </div>
                  <div>
                    <span>Land Cost:</span>
                    <span className="cost">${selectedLocation.land_cost.toLocaleString()}</span>
                  </div>
                </div>
                
                {selectedLocation.electricity_cost && (
                  <div className="additional-info">
                    <span className="info-label">Electricity Cost:</span>
                    <span className="info-value">{selectedLocation.electricity_cost}</span>
                  </div>
                )}
                
                <p className="location-description">{selectedLocation.description}</p>
              </div>
              
              {/* Building Options */}
              <div className="building-options">
                <h3>Select Facility Type</h3>
                
                {buildingOptions.map(building => {
                  const totalCost = building.cost + selectedLocation.land_cost;
                  const canAfford = budget >= totalCost;
                  
                  return (
                    <div 
                      key={building.id} 
                      className={`building-option ${!canAfford ? 'disabled' : ''}`}
                    >
                      <div className="building-header">
                        <h4>{building.name}</h4>
                        <span className="building-cost">${building.cost.toLocaleString()}</span>
                      </div>
                      
                      <p className="building-description">{building.description}</p>
                      
                      <div className="building-specs">
                        <div className="spec">
                          <span>Efficiency:</span>
                          <span>{building.energyEfficiency}%</span>
                        </div>
                        <div className="spec">
                          <span>Capacity:</span>
                          <span>{building.capacity} servers</span>
                        </div>
                        <div className="spec">
                          <span>Carbon Impact:</span>
                          <span>{building.carbonImpact} MT CO₂/day</span>
                        </div>
                      </div>
                      
                      <div className="total-cost">
                        <span>Total Cost:</span>
                        <span>${totalCost.toLocaleString()}</span>
                      </div>
                      
                      <button 
                        className="build-button"
                        onClick={() => handleBuild(building)}
                        disabled={!canAfford}
                      >
                        {canAfford ? 'Build Facility' : 'Insufficient Funds'}
                      </button>
                    </div>
                  );
                })}
              </div>
            </>
          ) : (
            <div className="empty-state">
              <h3>Select a Location</h3>
              <p>Click on a blue marker on the map to view location details and build a data center.</p>
              
              <div className="game-instructions">
                <h3>How to Play</h3>
                <ol>
                  <li>Select a location on the map by clicking a blue marker</li>
                  <li>Review the environmental metrics for that location</li>
                  <li>Choose a data center type to build based on your budget</li>
                  <li>Optimize your network for low carbon impact and high efficiency</li>
                  <li>Balance costs with environmental impact to maximize your score</li>
                </ol>
                
                <div className="game-goal">
                  <h4>Goal</h4>
                  <p>Build the most efficient and environmentally friendly global data center network while managing your budget wisely.</p>
                  <p>Higher scores are achieved by placing data centers in locations with good environmental metrics and using more efficient facilities.</p>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
      
      {/* Notification */}
      {notification && (
        <div className={`game-notification ${notification.type}`}>
          <p>{notification.message}</p>
          <button onClick={closeNotification} className="notification-close">×</button>
        </div>
      )}
    </div>
  );
}

export default Game;