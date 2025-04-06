// src/App.js
import React, { useState, useEffect } from 'react';
import './App.css';
import 'bootstrap/dist/css/bootstrap.min.css'; // <-- Import Bootstrap
import Login from './components/Login';
import Register from './components/Register';
import Game from './components/Game';
import ApiService from './services/api';
import logo from './logo.svg';

function App() {
  const [user, setUser] = useState(null);
  const [showRegister, setShowRegister] = useState(false);
  const [loading, setLoading] = useState(true);

  // Check if user is already logged in when app loads
  useEffect(() => {
    const checkLoginStatus = async () => {
      try {
        const profileData = await ApiService.getProfile();
        // Parse username from "Hello username! This is protected profile data!"
        const message = profileData.profile || "";
        const usernameMatch = message.match(/Hello ([^!]+)!/);
        const username = usernameMatch ? usernameMatch[1] : "User";

        setUser({ username });
      } catch (error) {
        console.log('User not logged in:', error);
        setUser(null);
      } finally {
        setLoading(false);
      }
    };

    checkLoginStatus();
  }, []);

  const handleLogin = (userData) => {
    setUser({ username: userData });
  };

  const handleRegister = (userData) => {
    setUser({ username: userData.username });
  };

  const handleLogout = async () => {
    try {
      await ApiService.logout();
      setUser(null);
    } catch (error) {
      console.error('Logout failed:', error);
      // Force logout on client side even if API call fails
      setUser(null);
    }
  };

  const switchToRegister = () => {
    setShowRegister(true);
  };

  const switchToLogin = () => {
    setShowRegister(false);
  };

  // Show loading state
  if (loading) {
    return (
      <div className="loading-container">
        <img src={logo} className="loading-logo" alt="logo" />
        <p>Loading Data Center Tycoon...</p>
      </div>
    );
  }

  // If not logged in, show login or register page
  if (!user) {
    if (showRegister) {
      return <Register onRegister={handleRegister} onSwitchToLogin={switchToLogin} />;
    } else {
      return <Login onLogin={handleLogin} onSwitchToRegister={switchToRegister} />;
    }
  }

  // If logged in, show the game
  return (
    <div className="App">
      <Game username={user.username} onLogout={handleLogout} />
    </div>
  );
}

export default App;
