// src/components/Game.js
import React, { useState, useEffect } from 'react';
import { MapContainer, TileLayer, Marker, Popup } from 'react-leaflet';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';
import './Game.css'; // <-- The CSS you will enhance below
// Add this line after your existing imports:



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



const potentialLocationIcon = new L.DivIcon({
  className: 'custom-map-marker',
  html: `<div class="map-dot orange-dot"></div>`,
  iconSize: [14, 14],
  iconAnchor: [7, 7],
  popupAnchor: [0, -7]
});

// A different icon if the item is in the cart
const cartIcon = new L.DivIcon({
  className: 'custom-map-marker',
  html: `<div class="map-dot cart-dot"></div>`,
  iconSize: [14, 14],
  iconAnchor: [7, 7],
  popupAnchor: [0, -7]
});

function Game({ username, onLogout }) {
  // Cart state
  const [cartItems, setCartItems] = useState([]);
  const [showCart, setShowCart] = useState(false);
  // New state for simulation modal and simulation data
  const [showSimulation, setShowSimulation] = useState(false);
  const [simulationData, setSimulationData] = useState({
    data: [],
    total_time_to_end: 0,
    time_datacenters_removed: 0
  });

  // Standard states
  const [availableLocations, setAvailableLocations] = useState([]);
  const [potentialLocations, setPotentialLocations] = useState([]);
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
  const [showFullDetails, setShowFullDetails] = useState(false);
  const [recentlyViewedLocations, setRecentlyViewedLocations] = useState([]);


  const fetchCarbonFootprint = async () => {
    try {
      const res = await fetch(`http://localhost:8080/cart/carbon-footprint?username=${username}`, {
        method: 'GET',
        credentials: 'include'
      });
      if (!res.ok) {
        throw new Error(`Failed to fetch carbon footprint: ${res.status}`);
      }
      const data = await res.json();
      setCarbonFootprint(data.carbon_footprint || 0);
    } catch (error) {
      console.error("Error fetching carbon footprint:", error);
    }
  };


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

  //----------------------------------------------------
  // CART FUNCTIONS
  //----------------------------------------------------
  // 1) Load the cart from backend
  const fetchCart = async () => {
    try {
      const res = await fetch(`http://localhost:8080/cart?username=${username}`, {
        method: 'GET',
        credentials: 'include'
      });
      if (!res.ok) {
        // Possibly 404 if no cart found, so just skip
        console.log("No existing cart found, or error fetching cart.");
        return;
      }
      const cartData = await res.json();
      // cartData may have .items array
      if (cartData && cartData.items) {
        setCartItems(cartData.items);
      } else {
        // If the cart structure is "Items" capitalized, or something else, adapt here
        setCartItems(cartData.Items || []);
      }
    } catch (err) {
      console.error("Error loading cart:", err);
    }
  };

  // 2) Add item to cart
  const addToCart = async (location) => {
    try {
      let cost = location.land_cost || 100000;
      const itemPayload = {
        username,
        item: {
          latitude: location.position.lat,
          longitude: location.position.lng,
          name: location.name || "Untitled",
          land_price: `$${cost.toLocaleString()}`,
          electricity: location.electricity_cost || "$0.07/kWh",
          notes: location.description || "Data center location"
        },
        cost
      };

      const res = await fetch("http://localhost:8080/cart/add", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify(itemPayload)
      });
      if (!res.ok) {
        throw new Error(`Failed to add item: ${res.status}`);
      }
      console.log("Item added to cart:", location);
      // Reload the cart
      fetchCart();
      fetchCarbonFootprint();
    } catch (err) {
      console.error("Error adding to cart:", err);
    }
  };

  // 3) Remove item from cart
  const removeCartItem = async (index) => {
    try {
      // Call /cart/item?username=XYZ&index=N
      const res = await fetch(`http://localhost:8080/cart/item?username=${username}&index=${index}`, {
        method: "DELETE",
        credentials: "include"
      });
      if (!res.ok) {
        throw new Error(`Failed to remove cart item: ${res.status}`);
      }
      console.log("Item removed from cart at index:", index);
      // Reload cart
      fetchCart();
      fetchCarbonFootprint();
    } catch (err) {
      console.error("Error removing cart item:", err);
    }
  };

  //----------------------------------------------------
  // SIMULATION FUNCTION
  //----------------------------------------------------
  const handleSimulate = async () => {
    try {
      const res = await fetch(`http://localhost:8080/api/simulation?username=${username}`, {
        method: 'GET',
        credentials: 'include'
      });
      if (!res.ok) throw new Error(`Simulation API error: ${res.status}`);
      const data = await res.json();
      setSimulationData(data);
      setShowSimulation(true);
    } catch (error) {
      console.error("Simulation error:", error);
    }
  };


  //----------------------------------------------------
  // Existing fetch of data centers & potential locations
  //----------------------------------------------------
  const fetchPotentialLocations = async () => {
    try {
      console.log("Starting fetch of potential locations...");
      const response = await fetch('http://localhost:8080/api/possible-datacenters', {
        method: 'GET',
        credentials: 'same-origin',
      });

      console.log("Response status:", response.status);

      if (!response.ok) {
        throw new Error(`Failed to fetch potential locations: ${response.status}`);
      }

      const rawData = await response.json();
      console.log("Potential locations received:", rawData);

      if (!rawData) {
        throw new Error("No data received from server");
      }
      const dataArray = Array.isArray(rawData) ? rawData : [rawData];
      if (dataArray.length === 0) {
        throw new Error("Empty data array received");
      }

      const locations = dataArray.map((dc, index) => ({
        id: `potential-${index + 1}`,
        position: {
          lat: dc.latitude || dc.Latitude || 0,
          lng: dc.longitude || dc.Longitude || 0
        },
        isPotential: true
      }));

      setPotentialLocations(locations);
    } catch (err) {
      console.error("Error fetching potential locations:", err);
      // fallback
      setPotentialLocations([
        {
          id: "potential-1",
          position: { lat: 37.7749, lng: -122.4194 },
          isPotential: true
        },
        {
          id: "potential-2",
          position: { lat: 52.5200, lng: 13.4050 },
          isPotential: true
        },
        {
          id: "potential-3",
          position: { lat: -33.8688, lng: 151.2093 },
          isPotential: true
        }
      ]);
    }
  };

  useEffect(() => {
    const fetchDataCenters = async () => {
      try {
        setIsLoading(true);
        console.log("Fetching data centers...");

        const response = await fetch('http://localhost:8080/alldatacenters', {
          method: 'GET',
          credentials: 'same-origin',
        });

        if (!response.ok) {
          throw new Error(`Failed to fetch data centers: ${response.status}`);
        }

        const rawData = await response.json();
        console.log("Data centers received:", rawData);

        if (!rawData) {
          throw new Error("No data received from server");
        }
        let dataArray = Array.isArray(rawData) ? rawData : [];
        if (dataArray.length === 0) {
          throw new Error("Empty data array received");
        }

        const firstItem = dataArray[0];
        if (typeof firstItem.latitude === 'undefined' || typeof firstItem.longitude === 'undefined') {
          console.warn("Data center format is different than expected, attempting to adapt");
          if (typeof firstItem.Latitude !== 'undefined' && typeof firstItem.Longitude !== 'undefined') {
            dataArray = dataArray.map(dc => ({
              id: dc.ID || Math.random().toString(36).substr(2, 9),
              name: dc.Name || "Unknown Location",
              latitude: dc.Latitude,
              longitude: dc.Longitude
            }));
          } else if (typeof firstItem.position !== 'undefined') {
            dataArray = dataArray.map(dc => ({
              id: dc.id || Math.random().toString(36).substr(2, 9),
              name: dc.name || "Unknown Location",
              latitude: dc.position.lat,
              longitude: dc.position.lng
            }));
          } else {
            throw new Error("Unrecognized data center format");
          }
        }

        const locations = dataArray.map((dc, index) => ({
          id: dc.id || index + 1,
          name: dc.name || dc.Name || `Location ${index + 1}`,
          position: {
            lat: dc.latitude || dc.Latitude || 0,
            lng: dc.longitude || dc.Longitude || 0
          }
        }));

        setAvailableLocations(locations);
        setIsLoading(false);
      } catch (err) {
        console.error("Error fetching data centers:", err);
        // fallback
        setAvailableLocations([
          {
            id: 1,
            name: "Northern Virginia",
            position: { lat: 38.8, lng: -77.2 }
          },
          {
            id: 2,
            name: "Oregon",
            position: { lat: 45.5, lng: -122.5 }
          },
          {
            id: 3,
            name: "Iceland",
            position: { lat: 64.1, lng: -21.9 }
          },
          {
            id: 4,
            name: "Singapore",
            position: { lat: 1.3, lng: 103.8 }
          },
          {
            id: 5,
            name: "Northern Sweden",
            position: { lat: 65.6, lng: 22.1 }
          }
        ]);

        setIsLoading(false);
        setError('Using fallback data while connecting to server...');
        setTimeout(() => setError(null), 3000);
      }
    };

    fetchDataCenters();
    fetchPotentialLocations();
  }, []);

  // Once user logs in, load their cart
  useEffect(() => {
    fetchCart();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [username]);

  // Helpers
  const calculateLocationScore = (location) => {
    if (!location || !location.climate) return 0;
    return (location.climate + location.renewable + location.grid + location.risk) / 4;
  };

  const handleLocationSelect = async (location) => {
    try {
      setLocationLoading(true);

      if (location.isPotential) {
        // This is a potential location, fetch details from property-details endpoint
        const response = await fetch(
          `http://localhost:8080/api/property-details?lat=${location.position.lat}&lng=${location.position.lng}`,
          {
            method: 'GET',
            credentials: 'same-origin',
          }
        );

        if (!response.ok) {
          throw new Error(`Failed to fetch property details: ${response.status}`);
        }

        let propertyData;
        try {
          propertyData = await response.json();
        } catch (error) {
          console.error("Error parsing JSON response:", error);
          throw new Error("Invalid JSON response from server");
        }

        let landCost = 3000000;
        try {
          const priceText = propertyData.land_price || "$3,000,000";
          const priceMatch = priceText.match(/\$([0-9,]+)/);
          if (priceMatch && priceMatch[1]) {
            landCost = parseInt(priceMatch[1].replace(/,/g, ''));
          }
        } catch (e) {
          console.error("Error parsing land price:", e);
        }

        const locationName = propertyData.location_name || "Potential Location";

        const details = {
          name: locationName,
          climate: Math.floor(Math.random() * 30) + 60,
          renewable: Math.floor(Math.random() * 40) + 40,
          grid: Math.floor(Math.random() * 40) + 40,
          risk: Math.floor(Math.random() * 20) + 70,
          land_cost: landCost,
          electricity_cost: propertyData.electricity || "$0.07/kWh",
          connectivity: propertyData.connectivity || "Standard",
          water_availability: propertyData.water_availability || "Adequate",
          tax_incentives: propertyData.tax_incentives || "None",
          zone_type: propertyData.zone_type || "Industrial",
          description: propertyData.notes || "A potential location for a new data center."
        };

        const enrichedLocation = { ...location, ...details };

        if (selectedLocation && selectedLocation.id !== location.id) {
          setRecentlyViewedLocations(prev => {
            const newList = [selectedLocation, ...prev.filter(loc => loc.id !== selectedLocation.id)].slice(0, 3);
            return newList;
          });
        }

        setSelectedLocation(enrichedLocation);
      } else {
        // Regular existing location
        const details = {
          climate: Math.floor(Math.random() * 30) + 60,
          renewable: Math.floor(Math.random() * 40) + 40,
          grid: Math.floor(Math.random() * 40) + 40,
          risk: Math.floor(Math.random() * 20) + 70,
          land_cost: Math.floor(Math.random() * 3000000) + 2000000,
          description: `Data center located in ${location.name} with excellent connectivity to major networks.`
        };

        const enrichedLocation = { ...location, ...details };

        if (selectedLocation && selectedLocation.id !== location.id) {
          setRecentlyViewedLocations(prev => {
            const newList = [selectedLocation, ...prev.filter(loc => loc.id !== selectedLocation.id)].slice(0, 3);
            return newList;
          });
        }

        setSelectedLocation(enrichedLocation);
      }

      setNotification(null);
      setLocationLoading(false);
    } catch (error) {
      console.error("Error generating location details:", error);

      const fallbackDetails = {
        name: location.name || "Unknown Location",
        climate: 70,
        renewable: 60,
        grid: 65,
        risk: 80,
        land_cost: 3000000,
        description: "Data about this location could not be loaded. Using fallback data."
      };

      setSelectedLocation({ ...location, ...fallbackDetails });
      setNotification({
        type: 'error',
        message: "Could not load full location details. Using fallback data."
      });
      setLocationLoading(false);
    }
  };

  const handleBuild = (building) => {
    if (!selectedLocation) return;
    let totalCost = building.cost + selectedLocation.land_cost;
    if (budget >= totalCost) {
      // existing build logic ...
      setBudget(prev => prev - totalCost);
      // If you want to track built data centers more thoroughly, you can do:
      // setBuiltDataCenters([...builtDataCenters, { building, location: selectedLocation }]);
      // setDay(day + 30); etc.

      // Then ALSO add to cart
      addToCart(selectedLocation);

      setNotification({
        type: 'success',
        message: `Successfully built ${building.name} in ${selectedLocation.name}!`
      });
    } else {
      setNotification({
        type: 'error',
        message: "Insufficient funds!"
      });
    }
  };

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

        <div className="game-stats">
          <div className="stat">
            <i className="bi bi-cash-stack stat-icon"></i>
            <span className="stat-label">Budget:</span>
            <span className="stat-value">${budget.toLocaleString()}</span>
          </div>

          <div className="stat">
            <i className="bi bi-bar-chart stat-icon"></i>
            <span className="stat-label">Score:</span>
            <span className="stat-value">{Math.round(score)}</span>
          </div>

          <div className="stat">
            <i className="bi bi-cloud-fog2 stat-icon"></i>
            <span className="stat-label">Carbon:</span>
            <span className="stat-value">{carbonFootprint.toFixed(1)} MT</span>
          </div>

          <div className="stat">
            <i className="bi bi-calendar3 stat-icon"></i>
            <span className="stat-label">Day:</span>
            <span className="stat-value">{day}</span>
          </div>
        </div>

        <div className="user-controls">
          <span className="username">Welcome, {username}!</span>
          <button onClick={onLogout} className="logout-button btn btn-danger">
            Logout
          </button>
        </div>
      </div>

      {/* Main Game Area */}
      <div className="game-content">
        {/* World Map */}
        <div className="world-map">
          <h2>Global Data Center Map</h2>
          <MapContainer center={[37.0902, -95.7129]} zoom={4} style={{ height: "500px", width: "100%" }}>
            <TileLayer
              url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
              attribution='&copy; OpenStreetMap contributors'
            />

            {availableLocations.map(location => {
              // Check if built
              const isBuilt = builtDataCenters.some(dc => dc.location.id === location.id);
              // Check if selected
              const isSelected = selectedLocation && selectedLocation.id === location.id;
              // Check if in cart
              const isInCart = cartItems.some(
                ci => parseFloat(ci.latitude) === location.position.lat &&
                  parseFloat(ci.longitude) === location.position.lng
              );

              if (!isBuilt) {
                let markerIcon = availableIcon;
                if (isSelected) {
                  markerIcon = selectedIcon;
                } else if (isInCart) {
                  markerIcon = cartIcon; // a special color for cart
                }

                return (
                  <Marker
                    key={location.id}
                    position={[location.position.lat, location.position.lng]}
                    icon={markerIcon}
                    eventHandlers={{
                      click: () => handleLocationSelect(location)
                    }}
                  >
                    <Popup>
                      <div>
                        <h3>{location.name}</h3>
                        <button
                          className="map-select-btn btn btn-sm btn-info"
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

            {potentialLocations.map(location => {
              const isSelected = selectedLocation && selectedLocation.id === location.id;
              const isInCart = cartItems.some(
                ci => parseFloat(ci.latitude) === location.position.lat &&
                  parseFloat(ci.longitude) === location.position.lng
              );

              let markerIcon = potentialLocationIcon;
              if (isSelected) {
                markerIcon = selectedIcon;
              } else if (isInCart) {
                markerIcon = cartIcon;
              }

              return (
                <Marker
                  key={location.id}
                  position={[location.position.lat, location.position.lng]}
                  icon={markerIcon}
                  eventHandlers={{
                    click: () => handleLocationSelect(location)
                  }}
                >
                  <Popup>
                    <div>
                      <h3>Potential Location</h3>
                      <button
                        className="map-select-btn btn btn-sm btn-info"
                        onClick={() => handleLocationSelect(location)}
                      >
                        View Details
                      </button>
                    </div>
                  </Popup>
                </Marker>
              );
            })}

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
                    <p>Built on day: {dataCenter.dayBuilt || 'N/A'}</p>
                    <p>Efficiency: {dataCenter.building.energyEfficiency}%</p>
                    <p>
                      Carbon Impact: {
                        (dataCenter.building.carbonImpact *
                          (100 - calculateLocationScore(dataCenter.location)) / 100
                        ).toFixed(2)
                      } MT/day
                    </p>
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
                    <span className="metric-icon climate">
                      <i className="bi bi-thermometer-sun"></i>
                    </span>
                    <span>Climate Suitability</span>
                  </div>
                  <div className="metric-bar">
                    <div
                      className="metric-fill"
                      style={{ width: `${selectedLocation.climate}%` }}
                    ></div>
                  </div>
                  <span className="metric-value">{selectedLocation.climate}%</span>
                </div>

                <div className="metric">
                  <div className="metric-label">
                    <span className="metric-icon renewable">
                      <i className="bi bi-sun"></i>
                    </span>
                    <span>Renewable Potential</span>
                  </div>
                  <div className="metric-bar">
                    <div
                      className="metric-fill"
                      style={{ width: `${selectedLocation.renewable}%` }}
                    ></div>
                  </div>
                  <span className="metric-value">{selectedLocation.renewable}%</span>
                </div>

                <div className="metric">
                  <div className="metric-label">
                    <span className="metric-icon grid">
                      <i className="bi bi-lightning-charge"></i>
                    </span>
                    <span>Grid Cleanliness</span>
                  </div>
                  <div className="metric-bar">
                    <div
                      className="metric-fill"
                      style={{ width: `${selectedLocation.grid}%` }}
                    ></div>
                  </div>
                  <span className="metric-value">{selectedLocation.grid}%</span>
                </div>

                <div className="metric">
                  <div className="metric-label">
                    <span className="metric-icon risk">
                      <i className="bi bi-exclamation-triangle"></i>
                    </span>
                    <span>Disaster Safety</span>
                  </div>
                  <div className="metric-bar">
                    <div
                      className="metric-fill"
                      style={{ width: `${selectedLocation.risk}%` }}
                    ></div>
                  </div>
                  <span className="metric-value">{selectedLocation.risk}%</span>
                </div>

                <div className="metric-summary">
                  <div>
                    <span>Overall Rating:</span>
                    <span className="score">
                      {Math.round(calculateLocationScore(selectedLocation))}/100
                    </span>
                  </div>
                  <div>
                    <span>Land Cost:</span>
                    <span className="cost">
                      ${selectedLocation.land_cost.toLocaleString()}
                    </span>
                  </div>
                </div>

                <p className="location-description">
                  {selectedLocation.description}
                </p>
              </div>

              {selectedLocation.isPotential && (
                <div className="property-details">
                  <h3>Property Details</h3>
                  <div className="detail-item">
                    <span className="detail-icon">
                      <i className="bi bi-building"></i>
                    </span>
                    <span className="detail-label">Location Name:</span>
                    <span className="detail-value">{selectedLocation.name}</span>
                  </div>
                  {selectedLocation.electricity_cost && (
                    <div className="detail-item">
                      <span className="detail-icon">
                        <i className="bi bi-lightning"></i>
                      </span>
                      <span className="detail-label">Electricity Cost:</span>
                      <span className="detail-value">{selectedLocation.electricity_cost}</span>
                    </div>
                  )}
                  {selectedLocation.connectivity && (
                    <div className="detail-item">
                      <span className="detail-icon">
                        <i className="bi bi-globe"></i>
                      </span>
                      <span className="detail-label">Network Connectivity:</span>
                      <span className="detail-value">{selectedLocation.connectivity}</span>
                    </div>
                  )}
                  <button
                    className="toggle-details-btn btn btn-sm btn-outline-secondary"
                    onClick={() => setShowFullDetails(prev => !prev)}
                  >
                    {showFullDetails ? 'Show Less Details' : 'Show More Details'}
                  </button>
                  {showFullDetails && (
                    <>
                      {selectedLocation.water_availability && (
                        <div className="detail-item">
                          <span className="detail-icon">
                            <i className="bi bi-droplet-half"></i>
                          </span>
                          <span className="detail-label">Water Availability:</span>
                          <span className="detail-value">
                            {selectedLocation.water_availability}
                          </span>
                        </div>
                      )}
                      {selectedLocation.tax_incentives && (
                        <div className="detail-item">
                          <span className="detail-icon">
                            <i className="bi bi-cash"></i>
                          </span>
                          <span className="detail-label">Tax Incentives:</span>
                          <span className="detail-value">
                            {selectedLocation.tax_incentives}
                          </span>
                        </div>
                      )}
                      {selectedLocation.zone_type && (
                        <div className="detail-item">
                          <span className="detail-icon">
                            <i className="bi bi-signpost-2"></i>
                          </span>
                          <span className="detail-label">Zone Type:</span>
                          <span className="detail-value">{selectedLocation.zone_type}</span>
                        </div>
                      )}
                    </>
                  )}
                </div>
              )}

              {/* Recently Viewed Comparison */}
              {recentlyViewedLocations.length > 0 && (
                <div className="location-comparison">
                  <h3>Location Comparison</h3>
                  <div className="comparison-chart table-responsive">
                    <table className="table table-sm table-bordered">
                      <thead className="thead-light">
                        <tr>
                          <th>Location</th>
                          <th>Climate</th>
                          <th>Renewable</th>
                          <th>Grid</th>
                          <th>Risk</th>
                          <th>Land Cost</th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr className="current-location">
                          <td>{selectedLocation.name}</td>
                          <td>{selectedLocation.climate}%</td>
                          <td>{selectedLocation.renewable}%</td>
                          <td>{selectedLocation.grid}%</td>
                          <td>{selectedLocation.risk}%</td>
                          <td>${selectedLocation.land_cost.toLocaleString()}</td>
                        </tr>
                        {recentlyViewedLocations.map(location => (
                          <tr key={location.id}>
                            <td>{location.name}</td>
                            <td>{location.climate}%</td>
                            <td>{location.renewable}%</td>
                            <td>{location.grid}%</td>
                            <td>{location.risk}%</td>
                            <td>${location.land_cost.toLocaleString()}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}

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
                      <div className="building-header d-flex justify-content-between align-items-center">
                        <h4>{building.name}</h4>
                        <span className="building-cost">
                          ${building.cost.toLocaleString()}
                        </span>
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
                        className="build-button btn btn-primary"
                        onClick={() => handleBuild(building)}
                        disabled={!canAfford}
                      >
                        {canAfford ? 'Build Facility' : 'Insufficient Funds'}
                      </button>
                    </div>
                  );
                })}
              </div>

              {/* Add to Cart Button */}
              <div className="my-3">
                <button
                  className="btn btn-sm btn-success"
                  onClick={() => addToCart(selectedLocation)}
                >
                  Add to Cart
                </button>
              </div>
              <div className="my-3">
                <button className="btn btn-warning" onClick={handleSimulate}>
                  <i className="bi bi-graph-up"></i> Simulate
                </button>
              </div>

            </>
          ) : (
            <div className="empty-state">
              <h3>Select a Location</h3>
              <p>
                Click on a blue marker on the map to view location details and build a data
                center.
              </p>
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
                  <p>
                    Build the most efficient and environmentally friendly global data center
                    network while managing your budget wisely.
                  </p>
                  <p>
                    Higher scores are achieved by placing data centers in locations with good
                    environmental metrics and using more efficient facilities.
                  </p>
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

      {/* CART BUTTON (bottom-right) */}
      <button
        className="btn btn-primary cart-toggle-button"
        onClick={() => setShowCart(true)}
      >
        <i className="bi bi-cart-fill"></i> Cart
      </button>

      {/* CART SIDEBAR */}
      {showCart && (
        <div className="cart-sidebar">
          <div className="cart-header d-flex justify-content-between align-items-center">
            <h4>Your Cart</h4>
            <button
              className="close-cart-btn btn btn-sm btn-danger"
              onClick={() => setShowCart(false)}
            >
              &times;
            </button>
          </div>
          <div className="cart-body mt-3">
            {cartItems.length === 0 ? (
              <p>No items in cart.</p>
            ) : (
              cartItems.map((item, idx) => (
                <div
                  key={idx}
                  className="cart-item d-flex justify-content-between align-items-center mb-2"
                >
                  <div>
                    <strong>{item.name}</strong>
                    <div className="text-muted">
                      {item.land_price} | {item.electricity}
                    </div>
                  </div>
                  <button
                    className="btn btn-outline-danger btn-sm"
                    onClick={() => removeCartItem(idx)}
                  >
                    Remove
                  </button>
                </div>
              ))
            )}
          </div>
        </div>
      )}
      {showCart && <div className="cart-backdrop" onClick={() => setShowCart(false)}></div>}
      {/* Simulation Modal */}
      <SimulationModal
        show={showSimulation}
        handleClose={() => setShowSimulation(false)}
        simulationData={simulationData}
      />

    </div>
  );
}

export default Game;
