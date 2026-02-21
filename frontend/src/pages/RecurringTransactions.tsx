import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { AxiosError } from 'axios';
import { recurringTransactionService } from '../services/recurring-transactions';
import type { RecurringTransaction, RecurringDeleteMode, RecurrenceFrequency } from '../types';
import Modal from '../components/ui/Modal';
import Pagination from '../components/ui/Pagination';
import toast from 'react-hot-toast';
import { HiTrash, HiPause, HiPlay } from 'react-icons/hi2';

const formatCurrency = (value: number) =>
  new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(value);

const formatDate = (dateStr: string) => {
  const d = new Date(dateStr + 'T00:00:00');
  return d.toLocaleDateString('pt-BR');
};

const FREQUENCY_LABELS: Record<RecurrenceFrequency, string> = {
  weekly: 'Semanal',
  biweekly: 'Quinzenal',
  monthly: 'Mensal',
  yearly: 'Anual',
};

const DELETE_MODE_LABELS: Record<RecurringDeleteMode, string> = {
  all: 'Excluir TODAS as transações',
  future_and_current: 'Excluir transações do mês atual e futuras',
  future_only: 'Excluir apenas transações futuras',
};

export default function RecurringTransactions() {
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [deleting, setDeleting] = useState<RecurringTransaction | null>(null);
  const [deleteMode, setDeleteMode] = useState<RecurringDeleteMode>('all');

  const { data: result, isLoading } = useQuery({
    queryKey: ['recurring-transactions', page],
    queryFn: () => recurringTransactionService.list({ page, per_page: 10 }).then((r) => r.data),
  });

  const invalidateAll = () => {
    queryClient.invalidateQueries({ queryKey: ['recurring-transactions'] });
    queryClient.invalidateQueries({ queryKey: ['transactions'] });
    queryClient.invalidateQueries({ queryKey: ['dashboard-summary'] });
    queryClient.invalidateQueries({ queryKey: ['dashboard-by-category'] });
    queryClient.invalidateQueries({ queryKey: ['dashboard-limits'] });
  };

  const deleteMutation = useMutation({
    mutationFn: ({ id, mode }: { id: string; mode: RecurringDeleteMode }) =>
      recurringTransactionService.delete(id, mode),
    onSuccess: () => {
      invalidateAll();
      setDeleting(null);
      toast.success('Recorrência excluída');
    },
    onError: (err: AxiosError<{ error: string }>) => toast.error(err.response?.data?.error || 'Erro ao excluir'),
  });

  const pauseMutation = useMutation({
    mutationFn: (id: string) => recurringTransactionService.pause(id),
    onSuccess: () => {
      invalidateAll();
      toast.success('Recorrência pausada');
    },
    onError: (err: AxiosError<{ error: string }>) => toast.error(err.response?.data?.error || 'Erro ao pausar'),
  });

  const resumeMutation = useMutation({
    mutationFn: (id: string) => recurringTransactionService.resume(id),
    onSuccess: () => {
      invalidateAll();
      toast.success('Recorrência retomada');
    },
    onError: (err: AxiosError<{ error: string }>) => toast.error(err.response?.data?.error || 'Erro ao retomar'),
  });

  const openDelete = (rt: RecurringTransaction) => {
    setDeleting(rt);
    setDeleteMode('all');
  };

  return (
    <div>
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Recorrências</h1>
      </div>

      {isLoading ? (
        <p className="text-gray-500">Carregando...</p>
      ) : (
        <>
          {/* Mobile cards */}
          <div className="space-y-3 md:hidden">
            {result?.data.map((rt) => (
              <div key={rt.id} className="bg-white rounded-xl shadow-sm p-4">
                <div className="flex items-start justify-between mb-2">
                  <div className="flex items-center gap-2 min-w-0">
                    <p className="font-medium text-gray-900 truncate">
                      {rt.description || rt.category_name}
                    </p>
                    <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium shrink-0 ${
                      rt.is_active ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'
                    }`}>
                      {rt.is_active ? 'Ativo' : 'Pausado'}
                    </span>
                  </div>
                  <span className={`font-medium whitespace-nowrap ${rt.type === 'income' ? 'text-green-600' : 'text-red-600'}`}>
                    {formatCurrency(rt.amount)}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <div className="text-sm text-gray-500">
                    <span>{rt.category_name}</span>
                    <span className="mx-1">&middot;</span>
                    <span>{FREQUENCY_LABELS[rt.frequency]}</span>
                    <span className="mx-1">&middot;</span>
                    <span>{formatDate(rt.start_date)}</span>
                  </div>
                  <div className="flex items-center gap-1">
                    {rt.is_active ? (
                      <button
                        onClick={() => pauseMutation.mutate(rt.id)}
                        disabled={pauseMutation.isPending}
                        className="p-2 text-gray-400 hover:text-yellow-600 hover:bg-yellow-50 rounded-lg transition-colors"
                        title="Pausar"
                      >
                        <HiPause className="w-4 h-4" />
                      </button>
                    ) : (
                      <button
                        onClick={() => resumeMutation.mutate(rt.id)}
                        disabled={resumeMutation.isPending}
                        className="p-2 text-gray-400 hover:text-green-600 hover:bg-green-50 rounded-lg transition-colors"
                        title="Retomar"
                      >
                        <HiPlay className="w-4 h-4" />
                      </button>
                    )}
                    <button
                      onClick={() => openDelete(rt)}
                      className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                    >
                      <HiTrash className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>
            ))}
            {result?.data.length === 0 && (
              <p className="text-center text-gray-400 py-8">Nenhuma recorrência cadastrada</p>
            )}
          </div>

          {/* Desktop table */}
          <div className="hidden md:block">
            <div className="bg-white rounded-xl shadow-sm overflow-hidden">
              <table className="w-full">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="text-left px-6 py-3 text-sm font-medium text-gray-500">Descrição</th>
                    <th className="text-left px-6 py-3 text-sm font-medium text-gray-500">Categoria</th>
                    <th className="text-left px-6 py-3 text-sm font-medium text-gray-500">Tipo</th>
                    <th className="text-left px-6 py-3 text-sm font-medium text-gray-500">Frequência</th>
                    <th className="text-left px-6 py-3 text-sm font-medium text-gray-500">Início</th>
                    <th className="text-right px-6 py-3 text-sm font-medium text-gray-500">Valor</th>
                    <th className="text-center px-6 py-3 text-sm font-medium text-gray-500">Status</th>
                    <th className="text-right px-6 py-3 text-sm font-medium text-gray-500">Ações</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {result?.data.map((rt) => (
                    <tr key={rt.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 text-gray-900">{rt.description || '-'}</td>
                      <td className="px-6 py-4 text-gray-500 text-sm">{rt.category_name}</td>
                      <td className="px-6 py-4 text-sm">
                        <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
                          rt.type === 'income' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                        }`}>
                          {rt.type === 'income' ? 'Receita' : 'Despesa'}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-gray-500 text-sm">{FREQUENCY_LABELS[rt.frequency]}</td>
                      <td className="px-6 py-4 text-gray-500 text-sm">{formatDate(rt.start_date)}</td>
                      <td className={`px-6 py-4 text-right font-medium ${rt.type === 'income' ? 'text-green-600' : 'text-red-600'}`}>
                        {formatCurrency(rt.amount)}
                      </td>
                      <td className="px-6 py-4 text-center">
                        <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
                          rt.is_active ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'
                        }`}>
                          {rt.is_active ? 'Ativo' : 'Pausado'}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right">
                        <div className="flex items-center justify-end gap-2">
                          {rt.is_active ? (
                            <button
                              onClick={() => pauseMutation.mutate(rt.id)}
                              disabled={pauseMutation.isPending}
                              className="p-1.5 text-gray-400 hover:text-yellow-600 hover:bg-yellow-50 rounded-lg transition-colors"
                              title="Pausar"
                            >
                              <HiPause className="w-4 h-4" />
                            </button>
                          ) : (
                            <button
                              onClick={() => resumeMutation.mutate(rt.id)}
                              disabled={resumeMutation.isPending}
                              className="p-1.5 text-gray-400 hover:text-green-600 hover:bg-green-50 rounded-lg transition-colors"
                              title="Retomar"
                            >
                              <HiPlay className="w-4 h-4" />
                            </button>
                          )}
                          <button
                            onClick={() => openDelete(rt)}
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
                      <td colSpan={8} className="px-6 py-8 text-center text-gray-400">
                        Nenhuma recorrência cadastrada
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
                onPageChange={setPage}
              />
            </div>
          )}
        </>
      )}

      {/* Delete Modal with mode selection */}
      <Modal
        isOpen={!!deleting}
        onClose={() => setDeleting(null)}
        title="Excluir Recorrência"
      >
        <div className="space-y-4">
          <p className="text-sm text-gray-600">
            Selecione o modo de exclusão para <strong>{deleting?.description || deleting?.category_name}</strong>:
          </p>
          <div className="space-y-2">
            {(['all', 'future_and_current', 'future_only'] as RecurringDeleteMode[]).map((mode) => (
              <label
                key={mode}
                className="flex items-center gap-3 p-3 rounded-lg border border-gray-200 cursor-pointer hover:bg-gray-50 transition-colors"
              >
                <input
                  type="radio"
                  name="deleteMode"
                  value={mode}
                  checked={deleteMode === mode}
                  onChange={() => setDeleteMode(mode)}
                  className="text-blue-600 focus:ring-blue-500"
                />
                <span className="text-sm text-gray-700">{DELETE_MODE_LABELS[mode]}</span>
              </label>
            ))}
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={() => setDeleting(null)}
              className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg"
            >
              Cancelar
            </button>
            <button
              type="button"
              onClick={() => deleting && deleteMutation.mutate({ id: deleting.id, mode: deleteMode })}
              disabled={deleteMutation.isPending}
              className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 disabled:opacity-50"
            >
              Excluir
            </button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
