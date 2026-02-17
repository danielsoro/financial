import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { transactionService } from '../../services/transactions';
import { categoryService } from '../../services/categories';
import type { Transaction, TransactionFilter } from '../../types';
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

  const createMutation = useMutation({
    mutationFn: (data: any) => transactionService.create(data),
    onSuccess: () => {
      invalidateAll();
      closeModal();
      toast.success(type === 'income' ? 'Receita criada' : 'Despesa criada');
    },
    onError: (err: any) => toast.error(err.response?.data?.error || 'Erro ao criar'),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) => transactionService.update(id, data),
    onSuccess: () => {
      invalidateAll();
      closeModal();
      toast.success('Atualizado');
    },
    onError: (err: any) => toast.error(err.response?.data?.error || 'Erro ao atualizar'),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => transactionService.delete(id),
    onSuccess: () => {
      invalidateAll();
      setDeleting(null);
      toast.success('Excluído');
    },
    onError: (err: any) => toast.error(err.response?.data?.error || 'Erro ao excluir'),
  });

  const openCreate = () => {
    setEditing(null);
    setCategoryId(categories[0]?.id || '');
    setAmount('');
    setDescription('');
    setDate(new Date().toISOString().split('T')[0]);
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
    const data = {
      type,
      category_id: categoryId,
      amount: parseFloat(amount),
      description,
      date,
    };
    if (editing) {
      updateMutation.mutate({ id: editing.id, data });
    } else {
      createMutation.mutate(data);
    }
  };

  const colorClass = type === 'income' ? 'text-green-600' : 'text-red-600';

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">{title}</h1>
        <MonthSelector month={month} year={year} onChange={handleMonthChange} />
        <button
          onClick={openCreate}
          className="flex items-center gap-2 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
        >
          <HiPlus className="w-5 h-5" /> Adicionar
        </button>
      </div>

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
            <label className="block text-sm font-medium text-gray-700 mb-1">Valor</label>
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
