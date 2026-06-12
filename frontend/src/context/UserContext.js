'use client';

import React, { createContext, useContext, useState, useEffect } from 'react';

export const API_BASE = 'http://localhost:8080/api';
const UserContext = createContext();

export function UserProvider({ children }) {
  const [user, setUser]     = useState(null);
  const [loading, setLoading] = useState(true);

  const fetchUser = async (username) => {
    try {
      const res = await fetch(`${API_BASE}/users/${username}`);
      if (res.ok) {
        const data = await res.json();
        setUser(data);
        localStorage.setItem('gg_user', data.username);
        return data;
      }
    } catch (_) {}
    localStorage.removeItem('gg_user');
    setUser(null);
    return null;
  };

  useEffect(() => {
    const stored = localStorage.getItem('gg_user');
    if (stored) fetchUser(stored).finally(() => setLoading(false));
    else setLoading(false);
  }, []);

  const login = async (username) => {
    setLoading(true);
    const u = await fetchUser(username);
    setLoading(false);
    if (!u) throw new Error('User not found. Please register first.');
    return u;
  };

  const logout = () => { localStorage.removeItem('gg_user'); setUser(null); };

  const register = async (profile) => {
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/users`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(profile),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Registration failed');
      setUser(data);
      localStorage.setItem('gg_user', data.username);
      return data;
    } finally {
      setLoading(false);
    }
  };

  const refreshUser = async () => { if (user) await fetchUser(user.username); };

  return (
    <UserContext.Provider value={{ user, loading, login, logout, register, refreshUser }}>
      {children}
    </UserContext.Provider>
  );
}

export function useUser() {
  const ctx = useContext(UserContext);
  if (!ctx) throw new Error('useUser must be used inside UserProvider');
  return ctx;
}
