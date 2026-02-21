import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { AxiosError } from 'axios';
import { transactionService } from '../../services/transactions';
import { categoryService } from '../../services/categories';
import { recurringTransactionService } from '../../services/recurring-transactions';
import type { Transaction, TransactionFilter, RecurrenceFrequency } from '../../types';
import Modal from '../ui/Modal';
import ConfirmDialog from '../ui/ConfirmDialog';
import Pagination from '../ui/Pagination';
import MonthSelector from '../ui/MonthSelector';
import Autocomplete from '../ui/Autocomplete';
import toast from 'react-hot-toast';
import { HiPlus, HiPencil, HiTrash } from 'react-icons/hi2';

const formatCurrency = (value: number) =>
  new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(value);

const formatDate = (dateStr: string) => {
  const d = new Date(dateStr + 'T00:00:00');
  return d.toLocaleDateString('pt-BR');
};

interface Props {
  type: 'income' | 'expense';
  title: string;
}

function getMonthRange(month: number, year: number) {
  const start = new Date(year, month - 1, 1);
  const end = new Date(year, month, 0);
  return {
    start_date: start.toISOString().split('T')[0],
    end_date: end.toISOString().split('T')[0],
  };
}

export default function TransactionPage({ type, title }: Props) {
  const queryClient = useQueryClient();
  const now = new Date();
  const [month, setMonth] = useState(now.getMonth() + 1);
  const [year, setYear] = useState(now.getFullYear());
  const { start_date, end_date } = getMonthRange(month, year);
  const [filter, setFilter] = useState<TransactionFilter>({ type, page: 1, per_page: 10, start_date, end_date });

  const handleMonthChange = (m: number, y: number) => {
    setMonth(m);
    setYear(y);
    const range = getMonthRange(m, y);
    setFilter((prev) => ({ ...prev, ...range, page: 1 }));
  };
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Transaction | null>(null);
  const [deleting, setDeleting] = useState<Transaction | null>(null);

  const [categoryId, setCategoryId] = useState('');
  const [amount, setAmount] = useState('');
  const [description, setDescription] = useState('');
  const [date, setDate] = useState(new Date().toISOString().split('T')[0]);

  // Recurring fields
  const [isRecurring, setIsRecurring] = useState(false);
  const [frequency, setFrequency] = useState<RecurrenceFrequency>('monthly');
  const [endCondition, setEndCondition] = useState<'indefinite' | 'end_date' | 'max_occurrences'>('indefinite');
  const [endDate, setEndDate] = useState('');
  const [maxOccurrences, setMaxOccurrences] = useState('');

  const { data: result, isLoading } = useQuery({
    queryKey: ['transactions', filter],
    queryFn: () => transactionService.list(filter).then((r) => r.data),
  });

  const { data: categories = [] } = useQuery({
    queryKey: ['categories', type],
    queryFn: () => categoryService.list(type).then((r) => r.data),
  });

  const invalidateAll = () => {
    queryClient.invalidateQueries({ queryKey: ['transactions'] });
    queryClient.invalidateQueries({ queryKey: ['dashboard-summary'] });
    queryClient.invalidateQueries({ queryKey: ['dashboard-by-category'] });
    queryClient.invalidateQueries({ queryKey: ['dashboard-limits'] });
  };

  type TransactionData = Omit<Transaction, 'id' | 'user_id' | 'category_name' | 'created_at' | 'updated_at'>;

  const createMutation = useMutation({
    mutationFn: (data: TransactionData) => transactionService.create(data),
    onSuccess: () => {
      invalidateAll();
      closeModal();
      toast.success(type === 'income' ? 'Receita criada' : 'Despesa criada');
    },
    onError: (err: AxiosError<{ error: string }>) => toast.error(err.response?.data?.error || 'Erro ao criar'),
  });

  const createRecurringMutation = useMutation({
    mutationFn: (data: {
      type: 'income' | 'expense';
      category_id: string;
      amount: number;
      description: string;
      frequency: RecurrenceFrequency;
      start_date: string;
      end_date: string | null;
      max_occurrences: number | null;
      day_of_month: number | null;
    }) => recurringTransactionService.create(data),
    onSuccess: () => {
      invalidateAll();
      queryClient.invalidateQueries({ queryKey: ['recurring-transactions'] });
      closeModal();
      toast.success('Recorrência criada');
    },
    onError: (err: AxiosError<{ error: string }>) => toast.error(err.response?.data?.error || 'Erro ao criar recorrência'),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: TransactionData }) => transactionService.update(id, data),
    onSuccess: () => {
      invalidateAll();
      closeModal();
      toast.success('Atualizado');
    },
    onError: (err: AxiosError<{ error: string }>) => toast.error(err.response?.data?.error || 'Erro ao atualizar'),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => transactionService.delete(id),
    onSuccess: () => {
      invalidateAll();
      setDeleting(null);
      toast.success('Excluído');
    },
    onError: (err: AxiosError<{ error: string }>) => toast.error(err.response?.data?.error || 'Erro ao excluir'),
  });

  const openCreate = () => {
    setEditing(null);
    setCategoryId(categories[0]?.id || '');
    setAmount('');
    setDescription('');
    setDate(new Date().toISOString().split('T')[0]);
    setIsRecurring(false);
    setFrequency('monthly');
    setEndCondition('indefinite');
    setEndDate('');
    setMaxOccurrences('');
    setModalOpen(true);
  };

  const openEdit = (tx: Transaction) => {
    setEditing(tx);
    setCategoryId(tx.category_id);
    setAmount(String(tx.amount));
    setDescription(tx.description);
    setDate(tx.date);
    setModalOpen(true);
  };

  const closeModal = () => {
    setModalOpen(false);
    setEditing(null);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (editing) {
      updateMutation.mutate({
        id: editing.id,
        data: { type, category_id: categoryId, amount: parseFloat(amount), description, date },
      });
    } else if (isRecurring) {
      const startDate = date;
      const parsedDay = new Date(startDate + 'T00:00:00').getDate();
      createRecurringMutation.mutate({
        type,
        category_id: categoryId,
        amount: parseFloat(amount),
        description,
        frequency,
        start_date: startDate,
        end_date: endCondition === 'end_date' ? endDate : null,
        max_occurrences: endCondition === 'max_occurrences' ? parseInt(maxOccurrences) : null,
        day_of_month: frequency === 'monthly' || frequency === 'yearly' ? parsedDay : null,
      });
    } else {
      createMutation.mutate({ type, category_id: categoryId, amount: parseFloat(amount), description, date });
    }
  };

  const colorClass = type === 'income' ? 'text-green-600' : 'text-red-600';

  return (
    <div>
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
        <h1 className="text-2xl font-bold text-gray-900">{title}</h1>
        <div className="flex items-center gap-3">
          <MonthSelector month={month} year={year} onChange={handleMonthChange} />
          <button
            onClick={openCreate}
            className="hidden md:flex items-center gap-2 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors whitespace-nowrap"
          >
            <HiPlus className="w-5 h-5" /> Adicionar
          </button>
        </div>
      </div>

      {/* FAB mobile */}
      <button
        onClick={openCreate}
        className="md:hidden fixed bottom-6 right-6 z-20 bg-blue-600 text-white p-4 rounded-full shadow-lg hover:bg-blue-700 transition-colors"
        aria-label="Adicionar"
      >
        <HiPlus className="w-6 h-6" />
      </button>

      <div className="flex gap-4 mb-4">
        <Autocomplete
          value={filter.category_id || ''}
          onChange={(val) => setFilter({ ...filter, category_id: val || undefined, page: 1 })}
          options={categories.map((c) => ({ value: c.id, label: c.full_path || c.name }))}
          emptyOptionLabel="Todas categorias"
          placeholder="Filtrar por categoria"
        />
      </div>

      {isLoading ? (
        <p className="text-gray-500">Carregando...</p>
      ) : (
        <>
          {/* Mobile cards */}
          <div className="space-y-3 md:hidden">
            {result?.data.map((tx) => (
              <div key={tx.id} className="bg-white rounded-xl shadow-sm p-4">
                <div className="flex items-start justify-between mb-2">
                  <p className="font-medium text-gray-900 truncate mr-2">
                    {tx.description || '-'}
                  </p>
                  <span className={`font-medium whitespace-nowrap ${colorClass}`}>
                    {formatCurrency(tx.amount)}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <div className="text-sm text-gray-500">
                    <span>{tx.category_name}</span>
                    <span className="mx-1">&middot;</span>
                    <span>{formatDate(tx.date)}</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => openEdit(tx)}
                      className="p-2 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                    >
                      <HiPencil className="w-4 h-4" />
                    </button>
                    <button
                      onClick={() => setDeleting(tx)}
                      className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                    >
                      <HiTrash className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>
            ))}
            {result?.data.length === 0 && (
              <p className="text-center text-gray-400 py-8">Nenhuma transação encontrada</p>
            )}
          </div>

          {/* Desktop table */}
          <div className="hidden md:block">
          <div className="bg-white rounded-xl shadow-sm overflow-hidden">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="text-left px-6 py-3 text-sm font-medium text-gray-500">Data</th>
                  <th className="text-left px-6 py-3 text-sm font-medium text-gray-500">Descrição</th>
                  <th className="text-left px-6 py-3 text-sm font-medium text-gray-500">Categoria</th>
                  <th className="text-right px-6 py-3 text-sm font-medium text-gray-500">Valor</th>
                  <th className="text-right px-6 py-3 text-sm font-medium text-gray-500">Ações</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {result?.data.map((tx) => (
                  <tr key={tx.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 text-gray-600 text-sm">{formatDate(tx.date)}</td>
                    <td className="px-6 py-4 text-gray-900">{tx.description || '-'}</td>
                    <td className="px-6 py-4 text-gray-500 text-sm">{tx.category_name}</td>
                    <td className={`px-6 py-4 text-right font-medium ${colorClass}`}>
                      {formatCurrency(tx.amount)}
                    </td>
                    <td className="px-6 py-4 text-right">
                      <div className="flex items-center justify-end gap-2">
                        <button
                          onClick={() => openEdit(tx)}
                          className="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                        >
                          <HiPencil className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => setDeleting(tx)}
                          className="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                        >
                          <HiTrash className="w-4 h-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
                {result?.data.length === 0 && (
                  <tr>
                    <td colSpan={5} className="px-6 py-8 text-center text-gray-400">
                      Nenhuma transação encontrada
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
          </div>

          {result && result.total_pages > 1 && (
            <div className="mt-4">
              <Pagination
                page={result.page}
                totalPages={result.total_pages}
                onPageChange={(p) => setFilter({ ...filter, page: p })}
              />
            </div>
          )}
        </>
      )}

      <Modal isOpen={modalOpen} onClose={closeModal} title={editing ? 'Editar' : 'Adicionar'}>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Categoria</label>
            <Autocomplete
              value={categoryId}
              onChange={setCategoryId}
              options={categories.map((c) => ({ value: c.id, label: c.full_path || c.name }))}
              placeholder="Selecione uma categoria"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              {isRecurring && endCondition === 'max_occurrences' ? 'Valor total' : 'Valor'}
            </label>
            <input
              type="number"
              step="0.01"
              min="0.01"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              required
              className="w-full rounded-lg border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            {isRecurring && endCondition === 'max_occurrences' && (() => {
              const n = parseInt(maxOccurrences);
              const a = parseFloat(amount);
              if (n > 0 && a > 0 && !isNaN(n) && !isNaN(a)) {
                const installment = Math.round((a / n) * 100) / 100;
                return (
                  <p className="text-xs text-gray-500 mt-1">
                    ({n}x de {formatCurrency(installment)})
                  </p>
                );
              }
              return null;
            })()}
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Descrição</label>
            <input
              type="text"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Data</label>
            <input
              type="date"
              value={date}
              onChange={(e) => setDate(e.target.value)}
              required
              className="w-full rounded-lg border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          {!editing && (
            <div className="space-y-3">
              <label className="flex items-center gap-2 text-sm font-medium text-gray-700 cursor-pointer">
                <input
                  type="checkbox"
                  checked={isRecurring}
                  onChange={(e) => setIsRecurring(e.target.checked)}
                  className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                />
                Recorrente
              </label>
              {isRecurring && (
                <div className="space-y-3 pl-6 border-l-2 border-blue-200">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Frequência</label>
                    <select
                      value={frequency}
                      onChange={(e) => setFrequency(e.target.value as RecurrenceFrequency)}
                      className="w-full rounded-lg border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="weekly">Semanal</option>
                      <option value="biweekly">Quinzenal</option>
                      <option value="monthly">Mensal</option>
                      <option value="yearly">Anual</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Condição de término</label>
                    <select
                      value={endCondition}
                      onChange={(e) => setEndCondition(e.target.value as 'indefinite' | 'end_date' | 'max_occurrences')}
                      className="w-full rounded-lg border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="indefinite">Indefinido</option>
                      <option value="end_date">Data final</option>
                      <option value="max_occurrences">Número de parcelas</option>
                    </select>
                  </div>
                  {endCondition === 'end_date' && (
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Data final</label>
                      <input
                        type="date"
                        value={endDate}
                        onChange={(e) => setEndDate(e.target.value)}
                        required
                        className="w-full rounded-lg border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                    </div>
                  )}
                  {endCondition === 'max_occurrences' && (
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Número de parcelas</label>
                      <input
                        type="number"
                        min="1"
                        value={maxOccurrences}
                        onChange={(e) => setMaxOccurrences(e.target.value)}
                        required
                        className="w-full rounded-lg border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                    </div>
                  )}
                </div>
              )}
            </div>
          )}
          <div className="flex justify-end gap-3">
            <button type="button" onClick={closeModal} className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg">
              Cancelar
            </button>
            <button type="submit" className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
              {editing ? 'Salvar' : 'Criar'}
            </button>
          </div>
        </form>
      </Modal>

      <ConfirmDialog
        isOpen={!!deleting}
        onClose={() => setDeleting(null)}
        onConfirm={() => deleting && deleteMutation.mutate(deleting.id)}
        title="Excluir"
        message="Tem certeza que deseja excluir esta transação?"
      />
    </div>
  );
}
