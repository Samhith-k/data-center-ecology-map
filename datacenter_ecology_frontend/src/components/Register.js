// src/components/Register.js
import React, { useState } from 'react';
import './Login.css'; // Reusing the same CSS
import logo from '../logo.svg';
import ApiService from '../services/api';

function Register({ onRegister, onSwitchToLogin }) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    // Basic validation
    if (!username.trim() || !password.trim() || !confirmPassword.trim()) {
      setError('Please fill in all fields');
      return;
    }
    
    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }
    
    try {
      setIsLoading(true);
      setError('');
      
      // Call register API
      const response = await ApiService.register(username, password);
      
      // Handle successful registration
      console.log('Registration successful:', response);
      onRegister({ username });
    } catch (err) {
      setError(err.message || 'Registration failed. Please try again.');
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
          <p className="login-subtitle">Create New Account</p>
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
              placeholder="Choose a username"
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
              placeholder="Create a password"
              disabled={isLoading}
            />
          </div>
          
          <div className="form-group">
            <label htmlFor="confirmPassword">Confirm Password</label>
            <input
              type="password"
              id="confirmPassword"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder="Confirm your password"
              disabled={isLoading}
            />
          </div>
          
          <button type="submit" className="login-button" disabled={isLoading}>
            {isLoading ? 'Creating Account...' : 'Create Account'}
          </button>
        </form>
        
        <div className="login-footer">
          <p>Already have an account? <button className="text-button" onClick={onSwitchToLogin} disabled={isLoading}>Login instead</button></p>
          <p>GreenTech Impact Award & Moonshot Venture Award Entry</p>
        </div>
      </div>
    </div>
  );
}

export default Register;