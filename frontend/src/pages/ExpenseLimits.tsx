import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { AxiosError } from "axios";
import { expenseLimitService } from "../services/expense-limits";
import { categoryService } from "../services/categories";
import type { ExpenseLimit } from "../types";
import MonthSelector from "../components/ui/MonthSelector";
import Modal from "../components/ui/Modal";
import ConfirmDialog from "../components/ui/ConfirmDialog";
import Autocomplete from "../components/ui/Autocomplete";
import toast from "react-hot-toast";
import { HiPlus, HiPencil, HiTrash, HiDocumentDuplicate } from "react-icons/hi2";

const formatCurrency = (value: number) =>
  new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(
    value,
  );

export default function ExpenseLimits() {
  const queryClient = useQueryClient();
  const now = new Date();
  const [month, setMonth] = useState(now.getMonth() + 1);
  const [year, setYear] = useState(now.getFullYear());
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<ExpenseLimit | null>(null);
  const [deleting, setDeleting] = useState<ExpenseLimit | null>(null);

  const [copyModalOpen, setCopyModalOpen] = useState(false);
  const [toMonth, setToMonth] = useState(month === 12 ? 1 : month + 1);
  const [toYear, setToYear] = useState(month === 12 ? year + 1 : year);

  const [categoryId, setCategoryId] = useState<string>("");
  const [amount, setAmount] = useState("");

  const { data: limits = [], isLoading } = useQuery({
    queryKey: ["expense-limits", month, year],
    queryFn: () => expenseLimitService.list(month, year).then((r) => r.data),
  });

  const { data: categories = [] } = useQuery({
    queryKey: ["categories", "expense"],
    queryFn: () => categoryService.list("expense").then((r) => r.data),
  });

  const invalidateAll = () => {
    queryClient.invalidateQueries({ queryKey: ["expense-limits"] });
    queryClient.invalidateQueries({ queryKey: ["dashboard-limits"] });
  };

  const createMutation = useMutation({
    mutationFn: (data: { category_id?: string; month: number; year: number; amount: number }) =>
      expenseLimitService.create(data),
    onSuccess: () => {
      invalidateAll();
      closeModal();
      toast.success("Teto criado");
    },
    onError: (err: AxiosError<{ error: string }>) =>
      toast.error(err.response?.data?.error || "Erro ao criar"),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, amount }: { id: string; amount: number }) =>
      expenseLimitService.update(id, amount),
    onSuccess: () => {
      invalidateAll();
      closeModal();
      toast.success("Teto atualizado");
    },
    onError: (err: AxiosError<{ error: string }>) =>
      toast.error(err.response?.data?.error || "Erro ao atualizar"),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => expenseLimitService.delete(id),
    onSuccess: () => {
      invalidateAll();
      setDeleting(null);
      toast.success("Teto excluído");
    },
    onError: (err: AxiosError<{ error: string }>) =>
      toast.error(err.response?.data?.error || "Erro ao excluir"),
  });

  const copyMutation = useMutation({
    mutationFn: () =>
      expenseLimitService.copy({
        from_month: month,
        from_year: year,
        to_month: toMonth,
        to_year: toYear,
      }),
    onSuccess: (res) => {
      invalidateAll();
      setCopyModalOpen(false);
      toast.success(`${res.data.copied} teto(s) copiado(s)`);
    },
    onError: (err: AxiosError<{ error: string }>) =>
      toast.error(err.response?.data?.error || "Erro ao copiar"),
  });

  const openCopy = () => {
    const nextMonth = month === 12 ? 1 : month + 1;
    const nextYear = month === 12 ? year + 1 : year;
    setToMonth(nextMonth);
    setToYear(nextYear);
    setCopyModalOpen(true);
  };

  const openCreate = () => {
    setEditing(null);
    setCategoryId("");
    setAmount("");
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
        <div className="hidden md:flex items-center gap-2">
          {limits.length > 0 && (
            <button
              onClick={openCopy}
              className="flex items-center gap-2 border border-blue-600 text-blue-600 px-4 py-2 rounded-lg hover:bg-blue-50 transition-colors whitespace-nowrap"
            >
              <HiDocumentDuplicate className="w-5 h-5" /> Copiar para...
            </button>
          )}
          <button
            onClick={openCreate}
            className="flex items-center gap-2 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors whitespace-nowrap"
          >
            <HiPlus className="w-5 h-5" /> Novo Teto
          </button>
        </div>
      </div>

      {/* FAB mobile */}
      <button
        onClick={openCreate}
        className="md:hidden fixed bottom-6 right-6 z-20 bg-blue-600 text-white p-4 rounded-full shadow-lg hover:bg-blue-700 transition-colors"
        aria-label="Novo Teto"
      >
        <HiPlus className="w-6 h-6" />
      </button>

      <div className="mb-6">
        <MonthSelector
          month={month}
          year={year}
          onChange={(m, y) => {
            setMonth(m);
            setYear(y);
          }}
        />
      </div>

      {!isLoading && limits.length > 0 && (
        <div className="md:hidden mb-4">
          <button
            onClick={openCopy}
            className="flex items-center gap-2 border border-blue-600 text-blue-600 px-4 py-2 rounded-lg hover:bg-blue-50 transition-colors"
          >
            <HiDocumentDuplicate className="w-5 h-5" /> Copiar para...
          </button>
        </div>
      )}

      {isLoading ? (
        <p className="text-gray-500">Carregando...</p>
      ) : (
        <div className="space-y-4">
          {limits.length === 0 && (
            <p className="text-gray-400 text-center py-8">
              Nenhum teto definido para este mês
            </p>
          )}
          {limits.map((limit) => (
            <div key={limit.id} className="bg-white rounded-xl shadow-sm p-6">
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 mb-2">
                <h3 className="font-medium text-gray-900">
                  {limit.category_id
                    ? categories.find((c) => c.id === limit.category_id)
                        ?.full_path || limit.category_name
                    : "Global (todas as despesas)"}
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

      <Modal
        isOpen={modalOpen}
        onClose={closeModal}
        title={editing ? "Editar Teto" : "Novo Teto"}
      >
        <form onSubmit={handleSubmit} className="space-y-4">
          {!editing && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Categoria
              </label>
              <Autocomplete
                value={categoryId}
                onChange={setCategoryId}
                options={categories.map((c) => ({
                  value: c.id,
                  label: c.full_path || c.name,
                }))}
                emptyOptionLabel="Global (todas as despesas)"
                placeholder="Selecione uma categoria"
              />
            </div>
          )}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Valor Limite
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
          </div>
          <div className="flex justify-end gap-3">
            <button
              type="button"
              onClick={closeModal}
              className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg"
            >
              Cancelar
            </button>
            <button
              type="submit"
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              {editing ? "Salvar" : "Criar"}
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

      <Modal
        isOpen={copyModalOpen}
        onClose={() => setCopyModalOpen(false)}
        title="Copiar Tetos para..."
      >
        <div className="space-y-4">
          <MonthSelector
            month={toMonth}
            year={toYear}
            onChange={(m, y) => {
              setToMonth(m);
              setToYear(y);
            }}
          />
          <div className="flex justify-end gap-3">
            <button
              type="button"
              onClick={() => setCopyModalOpen(false)}
              className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg"
            >
              Cancelar
            </button>
            <button
              type="button"
              onClick={() => copyMutation.mutate()}
              disabled={copyMutation.isPending}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
            >
              Copiar
            </button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
