import api from './api';
import type { AuthResponse, User } from '../types';

function getSubdomain(): string {
  const hostname = window.location.hostname;

  // Dev: localhost → root tenant
  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    return 'root';
  }

  // Dev: *.localhost → extract subdomain
  if (hostname.endsWith('.localhost')) {
    return hostname.replace('.localhost', '');
  }

  // Cloud Run: xxx.run.app → root tenant
  if (hostname.endsWith('.run.app')) {
    return 'root';
  }

  // Prod: sub.domain.com (3+ parts) → extract first part
  const parts = hostname.split('.');
  if (parts.length >= 3) {
    return parts[0];
  }

  // Prod: domain.com (2 parts) → root tenant
  return 'root';
}

export const authService = {
  login: (email: string, password: string) =>
    api.post<AuthResponse>('/auth/login', { email, password, subdomain: getSubdomain() }),

  getProfile: () => api.get<User>('/profile'),

  updateProfile: (data: { name: string; email: string }) =>
    api.put<User>('/profile', data),

  changePassword: (data: { old_password: string; new_password: string }) =>
    api.post('/profile/change-password', data),

  getSubdomain,
};
