import api from './api';
import type { Category } from '../types';

export const categoryService = {
  list: (type?: string, view?: 'flat' | 'tree') =>
    api.get<Category[]>('/categories', { params: { type, view } }),

  create: (data: { name: string; type: string; parent_id?: string }) =>
    api.post<Category>('/categories', data),

  update: (id: string, data: { name: string; type: string; parent_id?: string | null }) =>
    api.put<Category>(`/categories/${id}`, data),

  delete: (id: string) =>
    api.delete(`/categories/${id}`),
};
