import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { dashboardService } from '../services/dashboard';
import MonthSelector from '../components/ui/MonthSelector';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, Legend } from 'recharts';
import { HiArrowTrendingUp, HiArrowTrendingDown, HiBanknotes } from 'react-icons/hi2';

const formatCurrency = (value: number) =>
  new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(value);

const COLORS = ['#3b82f6', '#ef4444', '#10b981', '#f59e0b', '#8b5cf6', '#ec4899', '#06b6d4', '#84cc16'];

export default function Dashboard() {
  const now = new Date();
  const [month, setMonth] = useState(now.getMonth() + 1);
  const [year, setYear] = useState(now.getFullYear());

  const { data: summary } = useQuery({
    queryKey: ['dashboard-summary', month, year],
    queryFn: () => dashboardService.summary(month, year).then((r) => r.data),
  });

  const { data: byCategory = [] } = useQuery({
    queryKey: ['dashboard-by-category', month, year],
    queryFn: () => dashboardService.byCategory(month, year, 'expense').then((r) => r.data),
  });

  const { data: limitsProgress = [] } = useQuery({
    queryKey: ['dashboard-limits', month, year],
    queryFn: () => dashboardService.limitsProgress(month, year).then((r) => r.data),
  });

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <MonthSelector month={month} year={year} onChange={(m, y) => { setMonth(m); setYear(y); }} />
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="bg-white rounded-xl shadow-sm p-6">
          <div className="flex items-center gap-3 mb-2">
            <div className="p-2 bg-green-100 rounded-lg">
              <HiArrowTrendingUp className="w-5 h-5 text-green-600" />
            </div>
            <span className="text-sm text-gray-500">Receitas</span>
          </div>
          <p className="text-2xl font-bold text-green-600">{formatCurrency(summary?.total_income ?? 0)}</p>
          <p className="text-xs text-gray-400 mt-1">{summary?.income_count ?? 0} transações</p>
        </div>

        <div className="bg-white rounded-xl shadow-sm p-6">
          <div className="flex items-center gap-3 mb-2">
            <div className="p-2 bg-red-100 rounded-lg">
              <HiArrowTrendingDown className="w-5 h-5 text-red-600" />
            </div>
            <span className="text-sm text-gray-500">Despesas</span>
          </div>
          <p className="text-2xl font-bold text-red-600">{formatCurrency(summary?.total_expenses ?? 0)}</p>
          <p className="text-xs text-gray-400 mt-1">{summary?.expense_count ?? 0} transações</p>
        </div>

        <div className="bg-white rounded-xl shadow-sm p-6">
          <div className="flex items-center gap-3 mb-2">
            <div className="p-2 bg-blue-100 rounded-lg">
              <HiBanknotes className="w-5 h-5 text-blue-600" />
            </div>
            <span className="text-sm text-gray-500">Saldo</span>
          </div>
          <p className={`text-2xl font-bold ${(summary?.balance ?? 0) >= 0 ? 'text-blue-600' : 'text-red-600'}`}>
            {formatCurrency(summary?.balance ?? 0)}
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Category Chart */}
        <div className="bg-white rounded-xl shadow-sm p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Despesas por Categoria</h2>
          {byCategory.length === 0 ? (
            <p className="text-gray-400 text-center py-8">Sem dados para este mês</p>
          ) : (
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={byCategory}
                  dataKey="total"
                  nameKey="category_name"
                  cx="50%"
                  cy="50%"
                  outerRadius={100}
                  // eslint-disable-next-line @typescript-eslint/no-explicit-any
                  label={((props: Record<string, any>) => `${props.category_name}: ${formatCurrency(props.total)}`) as any}
                >
                  {byCategory.map((_, index) => (
                    <Cell key={index} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip formatter={(value) => formatCurrency(Number(value))} />
                <Legend />
              </PieChart>
            </ResponsiveContainer>
          )}
        </div>

        {/* Limits Progress */}
        <div className="bg-white rounded-xl shadow-sm p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Tetos de Gastos</h2>
          {limitsProgress.length === 0 ? (
            <p className="text-gray-400 text-center py-8">Nenhum teto definido</p>
          ) : (
            <div className="space-y-4">
              {limitsProgress.map((lp) => {
                const isOver = lp.percentage > 100;
                const barColor = isOver ? 'bg-red-500' : lp.percentage > 80 ? 'bg-yellow-500' : 'bg-green-500';
                return (
                  <div key={lp.limit.id}>
                    <div className="flex justify-between text-sm mb-1">
                      <span className="font-medium text-gray-700">
                        {lp.limit.category_id ? lp.limit.category_name : 'Global'}
                      </span>
                      <span className="text-gray-500">
                        {formatCurrency(lp.spent)} / {formatCurrency(lp.limit.amount)}
                      </span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2.5">
                      <div
                        className={`h-2.5 rounded-full transition-all ${barColor}`}
                        style={{ width: `${Math.min(lp.percentage, 100)}%` }}
                      />
                    </div>
                    <div className="flex justify-between text-xs mt-1">
                      <span className={isOver ? 'text-red-500 font-medium' : 'text-gray-400'}>
                        {lp.percentage.toFixed(1)}%
                      </span>
                      <span className="text-gray-400">
                        {isOver ? `Excedido em ${formatCurrency(lp.spent - lp.limit.amount)}` : `Restante: ${formatCurrency(lp.remaining)}`}
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
