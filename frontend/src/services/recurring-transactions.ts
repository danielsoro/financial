import api from './api';
import type { RecurringTransaction, RecurringTransactionFilter, RecurringDeleteMode, PaginatedResponse } from '../types';

export const recurringTransactionService = {
  list: (filter: RecurringTransactionFilter) =>
    api.get<PaginatedResponse<RecurringTransaction>>('/recurring-transactions', { params: filter }),

  create: (data: Omit<RecurringTransaction, 'id' | 'user_id' | 'category_name' | 'is_active' | 'paused_at' | 'created_at' | 'updated_at'>) =>
    api.post<RecurringTransaction>('/recurring-transactions', data),

  delete: (id: string, mode: RecurringDeleteMode) =>
    api.delete(`/recurring-transactions/${id}`, { data: { mode } }),

  pause: (id: string) =>
    api.post(`/recurring-transactions/${id}/pause`),

  resume: (id: string) =>
    api.post(`/recurring-transactions/${id}/resume`),
};
