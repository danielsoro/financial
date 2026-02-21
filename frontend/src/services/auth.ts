import api from './api';
import type { LoginResponse, SelectTenantResponse, User, InviteInfo } from '../types';

export const authService = {
  login: (email: string, password: string) =>
    api.post<LoginResponse>('/auth/login', { email, password }),

  selectTenant: (selectorToken: string, tenantId: string) =>
    api.post<SelectTenantResponse>('/auth/select-tenant', {
      selector_token: selectorToken,
      tenant_id: tenantId,
    }),

  register: (data: { name: string; email: string; password: string; tenant_name: string }) =>
    api.post<{ message: string }>('/auth/register', data),

  verifyEmail: (token: string) =>
    api.post<{ message: string }>('/auth/verify-email', { token }),

  getInviteInfo: (token: string) =>
    api.get<InviteInfo>('/auth/invite-info', { params: { token } }),

  acceptInvite: (data: { token: string; name?: string; password?: string }) =>
    api.post<{ message: string }>('/auth/accept-invite', data),

  getProfile: () => api.get<User>('/profile'),

  updateProfile: (data: { name: string; email: string }) =>
    api.put<User>('/profile', data),

  changePassword: (data: { old_password: string; new_password: string }) =>
    api.post('/profile/change-password', data),
};
