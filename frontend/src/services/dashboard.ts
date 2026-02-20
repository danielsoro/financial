import api from './api';
import type { DashboardSummary, CategoryTotal, LimitProgress } from '../types';

export type DashboardScope = 'tenant' | 'user';

export const dashboardService = {
  summary: (month: number, year: number, scope: DashboardScope = 'tenant') =>
    api.get<DashboardSummary>('/dashboard/summary', { params: { month, year, scope } }),

  byCategory: (month: number, year: number, type: string = 'expense', scope: DashboardScope = 'tenant') =>
    api.get<CategoryTotal[]>('/dashboard/by-category', { params: { month, year, type, scope } }),

  limitsProgress: (month: number, year: number, scope: DashboardScope = 'tenant') =>
    api.get<LimitProgress[]>('/dashboard/limits-progress', { params: { month, year, scope } }),
};
