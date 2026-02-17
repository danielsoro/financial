import api from './api';
import type { DashboardSummary, CategoryTotal, LimitProgress } from '../types';

export const dashboardService = {
  summary: (month: number, year: number) =>
    api.get<DashboardSummary>('/dashboard/summary', { params: { month, year } }),

  byCategory: (month: number, year: number, type: string = 'expense') =>
    api.get<CategoryTotal[]>('/dashboard/by-category', { params: { month, year, type } }),

  limitsProgress: (month: number, year: number) =>
    api.get<LimitProgress[]>('/dashboard/limits-progress', { params: { month, year } }),
};
