# Finance Frontend

SPA de finanças pessoais em React.

## Tecnologias

- **React 19** — UI
- **TypeScript 5.9** — strict mode habilitado
- **Vite 7** — build e dev server
- **Tailwind CSS v4** — estilização (via `@tailwindcss/vite`)
- **TanStack Query v5** — server state (queries e mutations)
- **React Router v7** — roteamento SPA
- **Axios** — requisições HTTP
- **Recharts** — gráficos no dashboard
- **react-hot-toast** — notificações
- **react-icons** — ícones (Heroicons v2)

## Estrutura de diretórios

```
src/
├── pages/              # Páginas da aplicação
│   ├── Dashboard.tsx   # Resumo financeiro com gráficos
│   ├── Income.tsx      # Wrapper → TransactionPage(type="income")
│   ├── Expense.tsx     # Wrapper → TransactionPage(type="expense")
│   ├── Categories.tsx  # Gestão de categorias (tree view)
│   ├── ExpenseLimits.tsx  # Tetos de gasto mensais
│   ├── Profile.tsx     # Editar perfil e senha
│   ├── Login.tsx       # Login
│   └── Register.tsx    # Cadastro
├── components/
│   ├── auth/
│   │   └── ProtectedRoute.tsx   # Guard de rota autenticada
│   ├── layout/
│   │   └── AppLayout.tsx        # Layout com sidebar
│   ├── transactions/
│   │   └── TransactionPage.tsx  # Tabela de transações (usado por Income e Expense)
│   └── ui/                      # Componentes reutilizáveis
│       ├── Modal.tsx
│       ├── ConfirmDialog.tsx
│       ├── MonthSelector.tsx
│       ├── Autocomplete.tsx
│       └── Pagination.tsx
├── services/           # Camada de API (axios)
│   ├── api.ts          # Instância axios, interceptors, baseURL
│   ├── auth.ts         # Login, registro, perfil
│   ├── categories.ts   # CRUD de categorias
│   ├── transactions.ts # CRUD de transações
│   ├── expense-limits.ts  # CRUD de tetos
│   └── dashboard.ts    # Summary, by-category, limits-progress
├── contexts/
│   └── AuthContext.tsx  # Estado de autenticação (token + user em localStorage)
├── types/
│   └── index.ts        # Interfaces TypeScript (User, Transaction, Category, etc.)
├── App.tsx             # Router + QueryClient
├── main.tsx            # Entrypoint
└── index.css           # @import "tailwindcss"
```

## Guia de estilo

### Cores semânticas

| Significado | Classes | Hex (gráficos) |
|-------------|---------|-----------------|
| Primary / balance | `blue-600`, `blue-700` | `#3b82f6` |
| Receita / positivo | `green-600`, `green-100` | `#10b981` |
| Despesa / negativo | `red-600`, `red-100` | `#ef4444` |
| Warning (teto perto) | `yellow-500` | `#f59e0b` |
| Background principal | `gray-50` | — |
| Sidebar | `gray-900` | — |

### Espaçamento e layout

- Sidebar fixa à esquerda (`w-64`, `bg-gray-900`)
- Conteúdo principal com `ml-64`, `p-8`, `bg-gray-50`, `min-h-screen`
- Cards: `bg-white rounded-xl shadow-sm p-6`
- Tabelas: `bg-white rounded-xl shadow-sm overflow-hidden`
- Espaçamento entre campos de form: `space-y-4`

### Componentes UI

- **Modal**: overlay `bg-black/50`, conteúdo centralizado `bg-white rounded-2xl`
- **ConfirmDialog**: modal com botões Cancelar (cinza) e Confirmar (`bg-red-600`)
- **Botão primário**: `bg-blue-600 text-white rounded-lg hover:bg-blue-700`
- **Botão cancelar**: `text-gray-700 hover:bg-gray-100 rounded-lg`
- **Botão ação (ícone)**: `text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg`
- **Botão deletar (ícone)**: `text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg`
- **Input**: `w-full rounded-lg border border-gray-300 px-3 py-2 focus:ring-2 focus:ring-blue-500`

### Padrão de formulário

Formulários controlados com `useState`. Estrutura:

```tsx
<form onSubmit={handleSubmit} className="space-y-4">
  <div>
    <label className="block text-sm font-medium text-gray-700 mb-1">Label</label>
    <input className="w-full rounded-lg border border-gray-300 px-3 py-2 ..." />
  </div>
  <div className="flex justify-end gap-3">
    <button type="button" onClick={closeModal}>Cancelar</button>
    <button type="submit">{editing ? 'Salvar' : 'Criar'}</button>
  </div>
</form>
```

### Padrão de mutation (React Query)

```tsx
const createMutation = useMutation({
  mutationFn: (data) => service.create(data),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['resource'] });
    closeModal();
    toast.success('Mensagem de sucesso');
  },
  onError: (err: any) => toast.error(err.response?.data?.error || 'Erro'),
});
```

Mutations de transações invalidam múltiplas queries: `transactions`, `dashboard-summary`, `dashboard-by-category`, `dashboard-limits`.

### Padrão de estado de modais

```tsx
const [modalOpen, setModalOpen] = useState(false);
const [editing, setEditing] = useState<Type | null>(null);
const [deleting, setDeleting] = useState<Type | null>(null);
```

## Rotas

| Path | Componente | Descrição | Auth |
|------|-----------|-----------|:----:|
| `/login` | Login | Tela de login | Não |
| `/register` | Register | Tela de cadastro | Não |
| `/` | Dashboard | Resumo financeiro | Sim |
| `/income` | Income | Tabela de receitas | Sim |
| `/expenses` | Expense | Tabela de despesas | Sim |
| `/categories` | Categories | Gestão de categorias | Sim |
| `/expense-limits` | ExpenseLimits | Tetos de gasto | Sim |
| `/profile` | Profile | Editar perfil | Sim |

## Services (camada de API)

### `api.ts` — Instância Axios

- `baseURL: '/api/v1'`
- **Request interceptor**: injeta `Authorization: Bearer <token>` do localStorage
- **Response interceptor**: em caso de 401 (exceto login), limpa localStorage e redireciona para `/login`
- **Proxy (dev)**: Vite redireciona `/api` → `http://localhost:8080`

### Services disponíveis

| Service | Endpoints |
|---------|-----------|
| `authService` | register, login, getProfile, updateProfile, changePassword |
| `categoryService` | list, create, update, delete |
| `transactionService` | list, getById, create, update, delete |
| `expenseLimitService` | list, create, update, delete |
| `dashboardService` | summary, byCategory, limitsProgress |

## Convenções

### Ordem de imports

1. React
2. Libs externas (react-router, tanstack, toast, icons)
3. Services
4. Components
5. Contexts
6. Types (`import type { ... }`)

### Naming

- Páginas: PascalCase (`Dashboard.tsx`, `ExpenseLimits.tsx`)
- Componentes: PascalCase (`MonthSelector.tsx`, `ConfirmDialog.tsx`)
- Services: camelCase (`transactionService`, `dashboardService`)
- Types: PascalCase (`Transaction`, `DashboardSummary`)

### Formatação

- Moeda: `new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' })`
- Data: `new Date(dateStr + 'T00:00:00').toLocaleDateString('pt-BR')`
- Meses em português: `['Janeiro', 'Fevereiro', ..., 'Dezembro']`

### QueryClient

```tsx
const queryClient = new QueryClient({
  defaultOptions: {
    queries: { staleTime: 60_000, retry: 1 },
  },
});
```

## Como rodar

```bash
# Da raiz do projeto (finance/)
make frontend    # npm run dev com proxy para backend

# Ou diretamente
cd frontend
npm install
npm run dev      # Dev server em http://localhost:5173
npm run build    # Build de produção
npm run lint     # ESLint
npx tsc --noEmit # Type-check
```
