import api from './api';
import type { AuthResponse, User } from '../types';

export const authService = {
  register: (name: string, email: string, password: string) =>
    api.post<AuthResponse>('/auth/register', { name, email, password }),

  login: (email: string, password: string) =>
    api.post<AuthResponse>('/auth/login', { email, password }),

  getProfile: () => api.get<User>('/profile'),

  updateProfile: (data: { name: string; email: string }) =>
    api.put<User>('/profile', data),

  changePassword: (data: { old_password: string; new_password: string }) =>
    api.post('/profile/change-password', data),
};
