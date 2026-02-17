import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { expenseLimitService } from '../services/expense-limits';
import { categoryService } from '../services/categories';
import type { ExpenseLimit } from '../types';
import MonthSelector from '../components/ui/MonthSelector';
import Modal from '../components/ui/Modal';
import ConfirmDialog from '../components/ui/ConfirmDialog';
import Autocomplete from '../components/ui/Autocomplete';
import toast from 'react-hot-toast';
import { HiPlus, HiPencil, HiTrash } from 'react-icons/hi2';

const formatCurrency = (value: number) =>
  new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(value);

export default function ExpenseLimits() {
  const queryClient = useQueryClient();
  const now = new Date();
  const [month, setMonth] = useState(now.getMonth() + 1);
  const [year, setYear] = useState(now.getFullYear());
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<ExpenseLimit | null>(null);
  const [deleting, setDeleting] = useState<ExpenseLimit | null>(null);

  const [categoryId, setCategoryId] = useState<string>('');
  const [amount, setAmount] = useState('');

  const { data: limits = [], isLoading } = useQuery({
    queryKey: ['expense-limits', month, year],
    queryFn: () => expenseLimitService.list(month, year).then((r) => r.data),
  });

  const { data: categories = [] } = useQuery({
    queryKey: ['categories', 'expense'],
    queryFn: () => categoryService.list('expense').then((r) => r.data),
  });

  const createMutation = useMutation({
    mutationFn: (data: any) => expenseLimitService.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense-limits'] });
      closeModal();
      toast.success('Teto criado');
    },
    onError: (err: any) => toast.error(err.response?.data?.error || 'Erro ao criar'),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, amount }: { id: string; amount: number }) =>
      expenseLimitService.update(id, amount),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense-limits'] });
      closeModal();
      toast.success('Teto atualizado');
    },
    onError: (err: any) => toast.error(err.response?.data?.error || 'Erro ao atualizar'),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => expenseLimitService.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense-limits'] });
      setDeleting(null);
      toast.success('Teto excluído');
    },
    onError: (err: any) => toast.error(err.response?.data?.error || 'Erro ao excluir'),
  });

  const openCreate = () => {
    setEditing(null);
    setCategoryId('');
    setAmount('');
    setModalOpen(true);
  };

  const openEdit = (limit: ExpenseLimit) => {
    setEditing(limit);
    setAmount(String(limit.amount));
    setModalOpen(true);
  };

  const closeModal = () => {
    setModalOpen(false);
    setEditing(null);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (editing) {
      updateMutation.mutate({ id: editing.id, amount: parseFloat(amount) });
    } else {
      createMutation.mutate({
        category_id: categoryId || undefined,
        month,
        year,
        amount: parseFloat(amount),
      });
    }
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Tetos de Gastos</h1>
        <button
          onClick={openCreate}
          className="flex items-center gap-2 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
        >
          <HiPlus className="w-5 h-5" /> Novo Teto
        </button>
      </div>

      <div className="mb-6">
        <MonthSelector month={month} year={year} onChange={(m, y) => { setMonth(m); setYear(y); }} />
      </div>

      {isLoading ? (
        <p className="text-gray-500">Carregando...</p>
      ) : (
        <div className="space-y-4">
          {limits.length === 0 && (
            <p className="text-gray-400 text-center py-8">Nenhum teto definido para este mês</p>
          )}
          {limits.map((limit) => (
            <div key={limit.id} className="bg-white rounded-xl shadow-sm p-6">
              <div className="flex items-center justify-between mb-2">
                <h3 className="font-medium text-gray-900">
                  {limit.category_id
                    ? (categories.find((c) => c.id === limit.category_id)?.full_path || limit.category_name)
                    : 'Global (todas as despesas)'}
                </h3>
                <div className="flex items-center gap-2">
                  <span className="text-lg font-semibold text-gray-900">
                    {formatCurrency(limit.amount)}
                  </span>
                  <button
                    onClick={() => openEdit(limit)}
                    className="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                  >
                    <HiPencil className="w-4 h-4" />
                  </button>
                  <button
                    onClick={() => setDeleting(limit)}
                    className="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                  >
                    <HiTrash className="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      <Modal isOpen={modalOpen} onClose={closeModal} title={editing ? 'Editar Teto' : 'Novo Teto'}>
        <form onSubmit={handleSubmit} className="space-y-4">
          {!editing && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Categoria</label>
              <Autocomplete
                value={categoryId}
                onChange={setCategoryId}
                options={categories.map((c) => ({ value: c.id, label: c.full_path || c.name }))}
                emptyOptionLabel="Global (todas as despesas)"
                placeholder="Selecione uma categoria"
              />
            </div>
          )}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Valor Limite</label>
            <input
              type="number"
              step="0.01"
              min="0.01"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              required
              className="w-full rounded-lg border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
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
        title="Excluir Teto"
        message="Tem certeza que deseja excluir este teto?"
      />
    </div>
  );
}
