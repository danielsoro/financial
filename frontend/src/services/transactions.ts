import api from './api';
import type { Transaction, TransactionFilter, PaginatedResponse } from '../types';

export const transactionService = {
  list: (filter: TransactionFilter) =>
    api.get<PaginatedResponse<Transaction>>('/transactions', { params: filter }),

  getById: (id: string) =>
    api.get<Transaction>(`/transactions/${id}`),

  create: (data: Omit<Transaction, 'id' | 'user_id' | 'category_name' | 'created_at' | 'updated_at'>) =>
    api.post<Transaction>('/transactions', data),

  update: (id: string, data: Omit<Transaction, 'id' | 'user_id' | 'category_name' | 'created_at' | 'updated_at'>) =>
    api.put<Transaction>(`/transactions/${id}`, data),

  delete: (id: string) =>
    api.delete(`/transactions/${id}`),
};
