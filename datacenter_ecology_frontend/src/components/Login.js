// src/components/Login.js
import React, { useState } from 'react';
import './Login.css';
import logo from '../logo.svg';
import ApiService from '../services/api';

function Login({ onLogin, onSwitchToRegister }) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    // Basic validation
    if (!username.trim() || !password.trim()) {
      setError('Please enter both username and password');
      return;
    }
    
    try {
      setIsLoading(true);
      setError('');
      
      // Call login API
      const response = await ApiService.login(username, password);
      
      // Handle successful login
      console.log('Login successful:', response);
      onLogin(username);
    } catch (err) {
      setError(err.message || 'Login failed. Please check your credentials.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="login-container">
      {/* Animated data center elements */}
      <div className="datacenter-background">
        <div className="server-rack rack1"></div>
        <div className="server-rack rack2"></div>
        <div className="server-rack rack3"></div>
        <div className="server-rack rack4"></div>
        <div className="server-rack rack5"></div>
        
        <div className="data-node node1"></div>
        <div className="data-node node2"></div>
        <div className="data-node node3"></div>
        <div className="data-node node4"></div>
        <div className="data-node node5"></div>
        <div className="data-node node6"></div>
        <div className="data-node node7"></div>
        <div className="data-node node8"></div>
        
        <div className="data-line line1"></div>
        <div className="data-line line2"></div>
        <div className="data-line line3"></div>
        <div className="data-line line4"></div>
        <div className="data-line line5"></div>
      </div>
      
      <div className="login-card">
        <div className="login-header">
          <img src={logo} className="login-logo" alt="logo" />
          <h1>Data Center Tycoon</h1>
          <p className="login-subtitle">Green Revolution Hackathon Project</p>
        </div>
        
        {error && <div className="login-error">{error}</div>}
        
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="username">Username</label>
            <input
              type="text"
              id="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Enter your username"
              disabled={isLoading}
            />
          </div>
          
          <div className="form-group">
            <label htmlFor="password">Password</label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter your password"
              disabled={isLoading}
            />
          </div>
          
          <button type="submit" className="login-button" disabled={isLoading}>
            {isLoading ? 'Logging in...' : 'Login to Play'}
          </button>
        </form>
        
        <div className="login-footer">
          <p>Don't have an account? <button className="text-button" onClick={onSwitchToRegister} disabled={isLoading}>Register now</button></p>
          <p>GreenTech Impact Award & Moonshot Venture Award Entry</p>
        </div>
      </div>
    </div>
  );
}

export default Login;