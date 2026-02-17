import { HiChevronLeft, HiChevronRight } from 'react-icons/hi2';

interface PaginationProps {
  page: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

export default function Pagination({ page, totalPages, onPageChange }: PaginationProps) {
  if (totalPages <= 1) return null;

  return (
    <div className="flex items-center justify-center gap-4 mt-6">
      <button
        onClick={() => onPageChange(page - 1)}
        disabled={page <= 1}
        className="inline-flex items-center gap-1 px-3 py-2 text-sm font-medium rounded-lg bg-white shadow hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        <HiChevronLeft className="h-4 w-4" />
        Anterior
      </button>

      <span className="text-sm text-gray-600">
        Página <span className="font-semibold text-gray-900">{page}</span> de{' '}
        <span className="font-semibold text-gray-900">{totalPages}</span>
      </span>

      <button
        onClick={() => onPageChange(page + 1)}
        disabled={page >= totalPages}
        className="inline-flex items-center gap-1 px-3 py-2 text-sm font-medium rounded-lg bg-white shadow hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        Próximo
        <HiChevronRight className="h-4 w-4" />
      </button>
    </div>
  );
}
