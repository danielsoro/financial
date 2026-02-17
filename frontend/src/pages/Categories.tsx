import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { AxiosError } from 'axios';
import { categoryService } from '../services/categories';
import type { Category } from '../types';
import Modal from '../components/ui/Modal';
import ConfirmDialog from '../components/ui/ConfirmDialog';
import toast from 'react-hot-toast';
import { HiPlus, HiPencil, HiTrash, HiChevronRight, HiChevronDown } from 'react-icons/hi2';

const typeLabels: Record<string, string> = {
  income: 'Receita',
  expense: 'Despesa',
  both: 'Ambos',
};

const typeBadgeColors: Record<string, string> = {
  income: 'bg-green-100 text-green-800',
  expense: 'bg-red-100 text-red-800',
  both: 'bg-blue-100 text-blue-800',
};

interface CategoryRowProps {
  cat: Category;
  level: number;
  expanded: Record<string, boolean>;
  onToggle: (id: string) => void;
  onEdit: (cat: Category) => void;
  onDelete: (cat: Category) => void;
  onAddChild: (cat: Category) => void;
}

function CategoryRow({ cat, level, expanded, onToggle, onEdit, onDelete, onAddChild }: CategoryRowProps) {
  const hasChildren = cat.children && cat.children.length > 0;
  const isExpanded = expanded[cat.id];

  return (
    <>
      <tr className="hover:bg-gray-50">
        <td className="px-6 py-4 text-gray-900">
          <div className="flex items-center" style={{ paddingLeft: `${level * 24}px` }}>
            {hasChildren ? (
              <button
                onClick={() => onToggle(cat.id)}
                className="p-0.5 mr-1 text-gray-400 hover:text-gray-600 rounded"
              >
                {isExpanded ? (
                  <HiChevronDown className="w-4 h-4" />
                ) : (
                  <HiChevronRight className="w-4 h-4" />
                )}
              </button>
            ) : (
              <span className="w-5 mr-1" />
            )}
            {cat.name}
          </div>
        </td>
        <td className="px-6 py-4">
          <span className={`px-2 py-1 rounded-full text-xs font-medium ${typeBadgeColors[cat.type]}`}>
            {typeLabels[cat.type]}
          </span>
        </td>
        <td className="px-6 py-4 text-gray-500 text-sm">
          {cat.is_default ? 'Padrão' : 'Custom'}
        </td>
        <td className="px-6 py-4 text-right">
          <div className="flex items-center justify-end gap-2">
            <button
              onClick={() => onAddChild(cat)}
              className="p-1.5 text-gray-400 hover:text-green-600 hover:bg-green-50 rounded-lg transition-colors"
              title="Criar subcategoria"
            >
              <HiPlus className="w-4 h-4" />
            </button>
            {!cat.is_default && (
              <>
                <button
                  onClick={() => onEdit(cat)}
                  className="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                >
                  <HiPencil className="w-4 h-4" />
                </button>
                <button
                  onClick={() => onDelete(cat)}
                  className="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                >
                  <HiTrash className="w-4 h-4" />
                </button>
              </>
            )}
          </div>
        </td>
      </tr>
      {hasChildren && isExpanded &&
        cat.children!.map((child) => (
          <CategoryRow
            key={child.id}
            cat={child}
            level={level + 1}
            expanded={expanded}
            onToggle={onToggle}
            onEdit={onEdit}
            onDelete={onDelete}
            onAddChild={onAddChild}
          />
        ))}
    </>
  );
}

export default function Categories() {
  const queryClient = useQueryClient();
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Category | null>(null);
  const [deleting, setDeleting] = useState<Category | null>(null);
  const [parentCategory, setParentCategory] = useState<Category | null>(null);
  const [name, setName] = useState('');
  const [type, setType] = useState('expense');
  const [expanded, setExpanded] = useState<Record<string, boolean>>({});

  const { data: categories = [], isLoading } = useQuery({
    queryKey: ['categories', 'tree'],
    queryFn: () => categoryService.list(undefined, 'tree').then((r) => r.data),
  });

  const createMutation = useMutation({
    mutationFn: (data: { name: string; type: string; parent_id?: string }) =>
      categoryService.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories'] });
      closeModal();
      toast.success('Categoria criada');
    },
    onError: (err: AxiosError<{ error: string }>) => toast.error(err.response?.data?.error || 'Erro ao criar categoria'),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: { name: string; type: string } }) =>
      categoryService.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories'] });
      closeModal();
      toast.success('Categoria atualizada');
    },
    onError: (err: AxiosError<{ error: string }>) => toast.error(err.response?.data?.error || 'Erro ao atualizar'),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => categoryService.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories'] });
      setDeleting(null);
      toast.success('Categoria excluída');
    },
    onError: (err: AxiosError<{ error: string }>) => toast.error(err.response?.data?.error || 'Erro ao excluir'),
  });

  const toggleExpand = (id: string) => {
    setExpanded((prev) => ({ ...prev, [id]: !prev[id] }));
  };

  const openCreate = () => {
    setEditing(null);
    setParentCategory(null);
    setName('');
    setType('expense');
    setModalOpen(true);
  };

  const openCreateChild = (parent: Category) => {
    setEditing(null);
    setParentCategory(parent);
    setName('');
    setType(parent.type);
    setModalOpen(true);
    // Auto-expand parent so user sees the new child
    setExpanded((prev) => ({ ...prev, [parent.id]: true }));
  };

  const openEdit = (cat: Category) => {
    setEditing(cat);
    setParentCategory(null);
    setName(cat.name);
    setType(cat.type);
    setModalOpen(true);
  };

  const closeModal = () => {
    setModalOpen(false);
    setEditing(null);
    setParentCategory(null);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (editing) {
      updateMutation.mutate({ id: editing.id, data: { name, type } });
    } else {
      const data: { name: string; type: string; parent_id?: string } = { name, type };
      if (parentCategory) {
        data.parent_id = parentCategory.id;
      }
      createMutation.mutate(data);
    }
  };

  const isSubcategoryMode = !editing && parentCategory !== null;
  const modalTitle = editing
    ? 'Editar Categoria'
    : parentCategory
      ? 'Nova Subcategoria'
      : 'Nova Categoria';

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Categorias</h1>
        <button
          onClick={openCreate}
          className="flex items-center gap-2 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
        >
          <HiPlus className="w-5 h-5" /> Nova Categoria
        </button>
      </div>

      {isLoading ? (
        <p className="text-gray-500">Carregando...</p>
      ) : (
        <div className="bg-white rounded-xl shadow-sm overflow-hidden">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="text-left px-6 py-3 text-sm font-medium text-gray-500">Nome</th>
                <th className="text-left px-6 py-3 text-sm font-medium text-gray-500">Tipo</th>
                <th className="text-left px-6 py-3 text-sm font-medium text-gray-500">Origem</th>
                <th className="text-right px-6 py-3 text-sm font-medium text-gray-500">Ações</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {categories.map((cat) => (
                <CategoryRow
                  key={cat.id}
                  cat={cat}
                  level={0}
                  expanded={expanded}
                  onToggle={toggleExpand}
                  onEdit={openEdit}
                  onDelete={setDeleting}
                  onAddChild={openCreateChild}
                />
              ))}
            </tbody>
          </table>
        </div>
      )}

      <Modal isOpen={modalOpen} onClose={closeModal} title={modalTitle}>
        <form onSubmit={handleSubmit} className="space-y-4">
          {isSubcategoryMode && (
            <div className="bg-gray-50 rounded-lg px-3 py-2 text-sm text-gray-700">
              Subcategoria de: <span className="font-semibold">{parentCategory!.name}</span>
            </div>
          )}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Nome</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              className="w-full rounded-lg border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          {!isSubcategoryMode && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Tipo</label>
              <select
                value={type}
                onChange={(e) => setType(e.target.value)}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="expense">Despesa</option>
                <option value="income">Receita</option>
                <option value="both">Ambos</option>
              </select>
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
        title="Excluir Categoria"
        message={`Tem certeza que deseja excluir a categoria "${deleting?.name}"? ${
          deleting?.children && deleting.children.length > 0
            ? 'Todas as subcategorias também serão excluídas.'
            : ''
        }`}
      />
    </div>
  );
}
