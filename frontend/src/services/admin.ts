import api from './api';
import type { AdminUser } from '../types';

export const adminService = {
  listUsers: () => api.get<AdminUser[]>('/admin/users'),

  createUser: (data: { name: string; email: string; password: string; role: string }) =>
    api.post<AdminUser>('/admin/users', data),

  updateUser: (id: string, data: { name: string; email: string; role: string }) =>
    api.put<AdminUser>(`/admin/users/${id}`, data),

  deleteUser: (id: string) => api.delete(`/admin/users/${id}`),

  resetPassword: (id: string, data: { new_password: string }) =>
    api.post(`/admin/users/${id}/reset-password`, data),
};
