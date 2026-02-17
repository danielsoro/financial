import api from './api';
import type { AuthResponse, User } from '../types';

function getSubdomain(): string {
  const hostname = window.location.hostname;

  // Dev: financial.localhost → financial; localhost → financial (fallback)
  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    return 'financial';
  }

  // Dev: *.localhost → extract subdomain
  if (hostname.endsWith('.localhost')) {
    return hostname.replace('.localhost', '');
  }

  // Cloud Run: xxx.run.app → fallback
  if (hostname.endsWith('.run.app')) {
    return 'financial';
  }

  // Prod: sub.domain.com → extract first part
  const parts = hostname.split('.');
  if (parts.length >= 3) {
    return parts[0];
  }

  // Bare domain → empty (super_admin login)
  return '';
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
