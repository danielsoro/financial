export interface User {
  id: string;
  name: string;
  email: string;
  role: 'owner' | 'admin' | 'user';
  global_user_id?: string;
  created_at: string;
  updated_at: string;
}

export interface TenantInfo {
  tenant_id: string;
  tenant_name: string;
  role: string;
}

export interface LoginResponse {
  // Single tenant (auto-select)
  token?: string;
  user?: User;
  tenant_id?: string;

  // Multi-tenant (selector)
  selector_token?: string;
  tenants?: TenantInfo[];
}

export interface SelectTenantResponse {
  token: string;
  user: User;
}

export interface Tenant {
  id: string;
  name: string;
  domain?: string;
  is_active: boolean;
  owner_id?: string;
  created_at: string;
  updated_at: string;
}

export interface AdminUser {
  id: string;
  name: string;
  email: string;
  role: 'owner' | 'admin' | 'user';
  created_at: string;
  updated_at: string;
}

export interface InviteInfo {
  tenant_name: string;
  email: string;
  role: string;
  user_exists: boolean;
}

export interface Category {
  id: string;
  user_id?: string;
  parent_id?: string | null;
  name: string;
  type: 'income' | 'expense' | 'both';
  is_default: boolean;
  full_path?: string;
  children?: Category[];
  created_at: string;
  updated_at: string;
}

export interface Transaction {
  id: string;
  user_id: string;
  category_id: string;
  category_name?: string;
  type: 'income' | 'expense';
  amount: number;
  description: string;
  date: string;
  recurring_id?: string | null;
  created_at: string;
  updated_at: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface TransactionFilter {
  type?: string;
  category_id?: string;
  start_date?: string;
  end_date?: string;
  page?: number;
  per_page?: number;
}

export interface ExpenseLimit {
  id: string;
  user_id: string;
  category_id: string | null;
  category_name?: string;
  month: number;
  year: number;
  amount: number;
  created_at: string;
  updated_at: string;
}

export interface LimitProgress {
  limit: ExpenseLimit;
  spent: number;
  remaining: number;
  percentage: number;
}

export interface DashboardSummary {
  total_income: number;
  total_expenses: number;
  balance: number;
  previous_balance: number;
  income_count: number;
  expense_count: number;
}

export interface CategoryTotal {
  category_id: string;
  category_name: string;
  total: number;
}

export type RecurrenceFrequency = 'weekly' | 'biweekly' | 'monthly' | 'yearly';

export interface RecurringTransaction {
  id: string;
  user_id: string;
  category_id: string;
  category_name?: string;
  type: 'income' | 'expense';
  amount: number;
  description: string;
  frequency: RecurrenceFrequency;
  start_date: string;
  end_date: string | null;
  max_occurrences: number | null;
  day_of_month: number | null;
  is_active: boolean;
  paused_at: string | null;
  created_at: string;
  updated_at: string;
}

export type RecurringDeleteMode = 'all' | 'future_and_current' | 'future_only';

export type ResumeConflictStrategy = 'create' | 'update';

export interface ResumeConflictResponse {
  conflict: true;
  existing_transactions: Transaction[];
}

export interface RecurringTransactionFilter {
  type?: string;
  is_active?: boolean;
  page?: number;
  per_page?: number;
}
