// src/services/api.js
const API_URL = 'http://localhost:8080'; // Your Go server address

const ApiService = {
  register: async (username, password) => {
    try {
      const response = await fetch(`${API_URL}/register`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password }),
        credentials: 'include', // For cookies
      });
      
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Registration failed');
      }
      
      return await response.json();
    } catch (error) {
      console.error('Registration error:', error);
      throw error;
    }
  },
  
  login: async (username, password) => {
    try {
      const response = await fetch(`${API_URL}/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password }),
        credentials: 'include', // For cookies
      });
      
      if (!response.ok) {
        throw new Error('Login failed. Please check your credentials.');
      }
      
      return await response.json();
    } catch (error) {
      console.error('Login error:', error);
      throw error;
    }
  },
  
  getProfile: async () => {
    try {
      const response = await fetch(`${API_URL}/profile`, {
        method: 'GET',
        credentials: 'include', // For cookies
      });
      
      if (!response.ok) {
        throw new Error('Failed to fetch profile');
      }
      
      return await response.json();
    } catch (error) {
      console.error('Profile error:', error);
      throw error;
    }
  },
  
  logout: async () => {
    try {
      const response = await fetch(`${API_URL}/logout`, {
        method: 'POST',
        credentials: 'include', // For cookies
      });
      
      if (!response.ok) {
        throw new Error('Logout failed');
      }
      
      return await response.json();
    } catch (error) {
      console.error('Logout error:', error);
      throw error;
    }
  },
  
  // In src/services/api.js
getAllDataCenters: async () => {
    try {
      const response = await fetch(`${API_URL}/alldatacenters`, {
        method: 'GET',
        credentials: 'include',
      });
      
      if (!response.ok) {
        throw new Error(`Failed to fetch data centers: ${response.status} ${response.statusText}`);
      }
      
      const dataCenters = await response.json();
      
      // Only transform the basic map data
      return dataCenters.map((dc, index) => ({
        id: index + 1,
        name: dc.name,
        position: { lat: dc.latitude, lng: dc.longitude }
      }));
    } catch (error) {
      console.error('Data centers error:', error);
      throw error;
    }
  },
  
  // Add a new function to get detailed data for a specific location
  getLocationDetails: async (locationId) => {
    try {
      // In a real app, you'd fetch from backend, but here we'll generate data
      return {
        climate: Math.floor(Math.random() * 30) + 60,
        renewable: Math.floor(Math.random() * 40) + 40,
        grid: Math.floor(Math.random() * 40) + 40,
        risk: Math.floor(Math.random() * 20) + 70,
        land_cost: Math.floor(Math.random() * 3000000) + 2000000,
        description: `Data center with advanced facilities providing reliable services.`
      };
    } catch (error) {
      console.error('Location details error:', error);
      throw error;
    }
  }
};

export default ApiService;