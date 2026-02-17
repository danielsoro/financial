import api from './api';
import type { ExpenseLimit } from '../types';

export const expenseLimitService = {
  list: (month: number, year: number) =>
    api.get<ExpenseLimit[]>('/expense-limits', { params: { month, year } }),

  create: (data: { category_id?: string; month: number; year: number; amount: number }) =>
    api.post<ExpenseLimit>('/expense-limits', data),

  update: (id: string, amount: number) =>
    api.put(`/expense-limits/${id}`, { amount }),

  delete: (id: string) =>
    api.delete(`/expense-limits/${id}`),
};
