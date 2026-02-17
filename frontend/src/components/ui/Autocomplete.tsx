import { useState, useRef, useEffect, useCallback, useMemo } from 'react';
import { HiChevronUpDown, HiCheck } from 'react-icons/hi2';

export interface AutocompleteOption {
  value: string;
  label: string;
}

interface AutocompleteProps {
  value: string;
  onChange: (value: string) => void;
  options: AutocompleteOption[];
  placeholder?: string;
  emptyOptionLabel?: string;
  className?: string;
}

export default function Autocomplete({
  value,
  onChange,
  options,
  placeholder = '',
  emptyOptionLabel,
  className = '',
}: AutocompleteProps) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState('');
  const [activeIndex, setActiveIndex] = useState(-1);
  const containerRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLUListElement>(null);

  const allOptions = useMemo(() => {
    if (emptyOptionLabel) {
      return [{ value: '', label: emptyOptionLabel }, ...options];
    }
    return options;
  }, [options, emptyOptionLabel]);

  const filtered = useMemo(() => {
    if (!search) return allOptions;
    const term = search.toLowerCase();
    return allOptions.filter((o) => o.label.toLowerCase().includes(term));
  }, [allOptions, search]);

  const selectedLabel = allOptions.find((o) => o.value === value)?.label || '';

  const close = useCallback(() => {
    setOpen(false);
    setSearch('');
    setActiveIndex(-1);
  }, []);

  const select = useCallback(
    (val: string) => {
      onChange(val);
      close();
    },
    [onChange, close],
  );

  // Click-outside
  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        close();
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [open, close]);

  // Scroll active option into view
  useEffect(() => {
    if (activeIndex < 0 || !listRef.current) return;
    const el = listRef.current.children[activeIndex] as HTMLElement | undefined;
    el?.scrollIntoView({ block: 'nearest' });
  }, [activeIndex]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!open) {
      if (e.key === 'ArrowDown' || e.key === 'Enter') {
        e.preventDefault();
        setOpen(true);
      }
      return;
    }

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setActiveIndex((i) => (i < filtered.length - 1 ? i + 1 : 0));
        break;
      case 'ArrowUp':
        e.preventDefault();
        setActiveIndex((i) => (i > 0 ? i - 1 : filtered.length - 1));
        break;
      case 'Enter':
        e.preventDefault();
        if (activeIndex >= 0 && activeIndex < filtered.length) {
          select(filtered[activeIndex].value);
        }
        break;
      case 'Escape':
        e.preventDefault();
        close();
        break;
    }
  };

  const handleInputClick = () => {
    if (!open) {
      setOpen(true);
      setSearch('');
      setActiveIndex(-1);
    }
  };

  const activeDescendant = activeIndex >= 0 ? `autocomplete-option-${activeIndex}` : undefined;

  return (
    <div ref={containerRef} className={`relative ${className}`}>
      <div className="relative">
        <input
          ref={inputRef}
          role="combobox"
          aria-expanded={open}
          aria-haspopup="listbox"
          aria-activedescendant={activeDescendant}
          type="text"
          className="w-full rounded-lg border border-gray-300 px-3 py-2 pr-8 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder={placeholder}
          value={open ? search : selectedLabel}
          onChange={(e) => {
            setSearch(e.target.value);
            setActiveIndex(-1);
          }}
          onClick={handleInputClick}
          onKeyDown={handleKeyDown}
          readOnly={!open}
        />
        <HiChevronUpDown className="absolute right-2 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
      </div>

      {open && (
        <ul
          ref={listRef}
          role="listbox"
          className="absolute z-50 mt-1 w-full max-h-60 overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg"
        >
          {filtered.length === 0 ? (
            <li className="px-3 py-2 text-sm text-gray-400">Nenhum resultado</li>
          ) : (
            filtered.map((opt, i) => {
              const isSelected = opt.value === value;
              const isActive = i === activeIndex;
              return (
                <li
                  key={opt.value || '__empty__'}
                  id={`autocomplete-option-${i}`}
                  role="option"
                  aria-selected={isSelected}
                  className={`flex items-center justify-between px-3 py-2 text-sm cursor-pointer ${
                    isActive ? 'bg-blue-50 text-blue-700' : 'text-gray-900 hover:bg-gray-50'
                  }`}
                  onMouseEnter={() => setActiveIndex(i)}
                  onMouseDown={(e) => {
                    e.preventDefault(); // prevent input blur
                    select(opt.value);
                  }}
                >
                  <span>{opt.label}</span>
                  {isSelected && <HiCheck className="w-4 h-4 text-blue-600 flex-shrink-0" />}
                </li>
              );
            })
          )}
        </ul>
      )}
    </div>
  );
}
