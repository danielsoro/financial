import { HiChevronLeft, HiChevronRight } from 'react-icons/hi2';

const MONTH_NAMES = [
  'Janeiro',
  'Fevereiro',
  'Março',
  'Abril',
  'Maio',
  'Junho',
  'Julho',
  'Agosto',
  'Setembro',
  'Outubro',
  'Novembro',
  'Dezembro',
];

interface MonthSelectorProps {
  month: number;
  year: number;
  onChange: (month: number, year: number) => void;
}

export default function MonthSelector({ month, year, onChange }: MonthSelectorProps) {
  const handlePrevious = () => {
    if (month === 1) {
      onChange(12, year - 1);
    } else {
      onChange(month - 1, year);
    }
  };

  const handleNext = () => {
    if (month === 12) {
      onChange(1, year + 1);
    } else {
      onChange(month + 1, year);
    }
  };

  return (
    <div className="inline-flex items-center gap-2 sm:gap-4">
      <button
        onClick={handlePrevious}
        className="rounded-lg bg-white shadow p-2 hover:bg-gray-100 transition-colors"
        aria-label="Mês anterior"
      >
        <HiChevronLeft className="h-5 w-5 text-gray-600" />
      </button>

      <span className="text-base sm:text-lg font-semibold text-gray-800 min-w-[140px] sm:min-w-[180px] text-center">
        {MONTH_NAMES[month - 1]} {year}
      </span>

      <button
        onClick={handleNext}
        className="rounded-lg bg-white shadow p-2 hover:bg-gray-100 transition-colors"
        aria-label="Próximo mês"
      >
        <HiChevronRight className="h-5 w-5 text-gray-600" />
      </button>
    </div>
  );
}
