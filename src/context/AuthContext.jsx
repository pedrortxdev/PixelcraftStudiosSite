import React, { createContext, useState, useContext, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../services/api';
import { usersAPI } from '../services/api';

const AuthContext = createContext(null);

function getStoredAuth() {
  const token = localStorage.getItem('pixelcraft_token');
  const userJson = localStorage.getItem('pixelcraft_user');
  const expiry = parseInt(localStorage.getItem('pixelcraft_token_expiry') || '0', 10);
  if (!token || !userJson || !expiry) return { token: null, user: null };
  if (Date.now() > expiry) {
    localStorage.removeItem('pixelcraft_token');
    localStorage.removeItem('pixelcraft_user');
    localStorage.removeItem('pixelcraft_token_expiry');
    return { token: null, user: null };
  }
  try {
    const user = JSON.parse(userJson);
    return { token, user };
  } catch {
    return { token: null, user: null };
  }
}

function setStoredAuth(token, user, ttlMs = 72 * 60 * 60 * 1000) { // 72 hours to match backend JWT
  localStorage.setItem('pixelcraft_token', token);
  localStorage.setItem('pixelcraft_user', JSON.stringify(user));
  localStorage.setItem('pixelcraft_token_expiry', String(Date.now() + ttlMs));
}

export const AuthProvider = ({ children }) => {
  const navigate = useNavigate();
  const [{ user, token }, setAuth] = useState(() => getStoredAuth());
  const [loading, setLoading] = useState(true);

  // Refresh user data from server - fetches fresh roles from database
  const refreshUser = useCallback(async () => {
    const currentToken = localStorage.getItem('pixelcraft_token');
    if (!currentToken) return null;

    try {
      const freshUser = await api.users.getMe();
      setStoredAuth(currentToken, freshUser);
      setAuth({ token: currentToken, user: freshUser });
      return freshUser;
    } catch (error) {
      console.error('Failed to refresh user:', error);
      return null;
    }
  }, []);

  useEffect(() => {
    let isMounted = true;

    // Initialize from storage and refresh from server
    const initial = getStoredAuth();
    if (isMounted) setAuth(initial);

    // If we have a token, refresh user data from server for latest roles
    if (initial.token) {
      refreshUser().finally(() => {
        if (isMounted) setLoading(false);
      });
    } else {
      if (isMounted) setLoading(false);
    }

    return () => {
      isMounted = false;
    };
  }, [refreshUser]);

  const login = useCallback(async ({ email, password }) => {
    setLoading(true);
    try {
      const data = await api.auth.login({ email, password });
      // data: { token, user }
      setStoredAuth(data.token, data.user);
      setAuth({ token: data.token, user: data.user });
      navigate('/dashboard');
      return data;
    } finally {
      setLoading(false);
    }
  }, [navigate]);

  const register = useCallback(async (userData) => {
    setLoading(true);
    try {
      const data = await api.auth.register(userData);
      // Após registrar, já autentica e vai para dashboard
      setStoredAuth(data.token, data.user);
      setAuth({ token: data.token, user: data.user });
      navigate('/dashboard');
      return data;
    } finally {
      setLoading(false);
    }
  }, [navigate]);

  const logout = useCallback(() => {
    localStorage.removeItem('pixelcraft_token');
    localStorage.removeItem('pixelcraft_user');
    localStorage.removeItem('pixelcraft_token_expiry');
    setAuth({ token: null, user: null });
    navigate('/login');
  }, [navigate]);

  const updateUser = useCallback(async (updates) => {
    await usersAPI.updateMe(updates);
    const freshUser = await refreshUser();
    return freshUser;
  }, [refreshUser]);

  const uploadAvatar = useCallback(async (file) => {
    const data = await usersAPI.uploadAvatar(file);
    const newUrl = data.avatar_url;
    await refreshUser();
    return newUrl;
  }, [refreshUser]);

  const generateAIAvatar = useCallback(async (prompt, userId = null) => {
    const data = await api.ai.generateAvatar(prompt, userId);
    const newUrl = data.avatar_url;

    // Only refresh if we are updating the current user
    if (!userId || userId === user?.id) {
      await refreshUser();
    }
    return newUrl;
  }, [user, refreshUser]);

  useEffect(() => {
    if (user && user.preferences) {
      const { density, font, backgroundFilter } = user.preferences;
      document.documentElement.setAttribute('data-density', density || 'comfortable');
      document.documentElement.setAttribute('data-font', font || 'modern');
      document.documentElement.setAttribute('data-bg-filter', backgroundFilter !== false ? 'on' : 'off');
    } else {
      document.documentElement.setAttribute('data-density', 'comfortable');
      document.documentElement.setAttribute('data-font', 'modern');
      document.documentElement.setAttribute('data-bg-filter', 'on');
    }
  }, [user]);

  return (
    <AuthContext.Provider value={{ user, token, loading, login, register, logout, updateUser, uploadAvatar, generateAIAvatar, refreshUser }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => useContext(AuthContext);