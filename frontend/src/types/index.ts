export interface User {
  id: string;
  name: string;
  email: string;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  token: string;
  user: User;
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
  income_count: number;
  expense_count: number;
}

export interface CategoryTotal {
  category_id: string;
  category_name: string;
  total: number;
}
