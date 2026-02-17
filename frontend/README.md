# Finance Frontend

SPA multi-tenant de finanças em React.

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

## Multi-Tenancy

- Tenant é resolvido por **subdomínio** (`financial.localhost` → tenant `financial`)
- Em `localhost` sem subdomínio, fallback para `financial`
- Login envia `subdomain` automaticamente
- Sidebar mostra links de admin condicionais por role
- `AdminRoute` protege rotas de administração

## Estrutura de diretórios

```
src/
├── pages/
│   ├── Dashboard.tsx      # Resumo financeiro com gráficos
│   ├── Income.tsx         # Wrapper → TransactionPage(type="income")
│   ├── Expense.tsx        # Wrapper → TransactionPage(type="expense")
│   ├── Categories.tsx     # Gestão de categorias (tree view)
│   ├── ExpenseLimits.tsx  # Tetos de gasto mensais
│   ├── Profile.tsx        # Editar perfil e senha
│   ├── Login.tsx          # Login
│   └── Users.tsx          # Gestão de usuários (admin)
├── components/
│   ├── auth/
│   │   ├── ProtectedRoute.tsx   # Guard de rota autenticada
│   │   └── AdminRoute.tsx       # Guard de rota admin
│   ├── layout/
│   │   └── AppLayout.tsx        # Layout com sidebar (links admin condicionais)
│   ├── transactions/
│   │   └── TransactionPage.tsx  # Tabela de transações
│   └── ui/                      # Componentes reutilizáveis
│       ├── Modal.tsx
│       ├── ConfirmDialog.tsx
│       ├── MonthSelector.tsx
│       ├── Autocomplete.tsx
│       └── Pagination.tsx
├── services/
│   ├── api.ts             # Instância axios, interceptors
│   ├── auth.ts            # Login, perfil, getSubdomain()
│   ├── categories.ts      # CRUD de categorias
│   ├── transactions.ts    # CRUD de transações
│   ├── expense-limits.ts  # CRUD de tetos
│   ├── dashboard.ts       # Summary, by-category, limits-progress
│   └── admin.ts           # Gestão de usuários (admin)
├── contexts/
│   └── AuthContext.tsx     # Estado de auth (token + user com role)
├── types/
│   └── index.ts           # User (com tenant_id, role), Tenant, AdminUser, etc.
├── App.tsx                # Router + QueryClient
├── main.tsx               # Entrypoint
└── index.css              # @import "tailwindcss"
```

## Rotas

| Path | Componente | Descrição | Auth | Role |
|------|-----------|-----------|:----:|:----:|
| `/login` | Login | Tela de login | Não | — |
| `/` | Dashboard | Resumo financeiro | Sim | — |
| `/income` | Income | Tabela de receitas | Sim | — |
| `/expenses` | Expense | Tabela de despesas | Sim | — |
| `/categories` | Categories | Gestão de categorias | Sim | — |
| `/expense-limits` | ExpenseLimits | Tetos de gasto | Sim | — |
| `/profile` | Profile | Editar perfil | Sim | — |
| `/admin/users` | Users | Gestão de usuários | Sim | admin+ |

## Services (camada de API)

### `api.ts` — Instância Axios

- `baseURL: '/api/v1'`
- **Request interceptor**: injeta `Authorization: Bearer <token>` do localStorage
- **Response interceptor**: em caso de 401 (exceto login), limpa localStorage e redireciona para `/login`
- **Proxy (dev)**: Vite redireciona `/api` → `http://localhost:8080`

### Services disponíveis

| Service | Endpoints |
|---------|-----------|
| `authService` | login, getProfile, updateProfile, changePassword, getSubdomain |
| `categoryService` | list, create, update, delete |
| `transactionService` | list, getById, create, update, delete |
| `expenseLimitService` | list, create, update, delete |
| `dashboardService` | summary, byCategory, limitsProgress |
| `adminService` | listUsers, createUser, updateUser, deleteUser, resetPassword |

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

### Badges de role

| Role | Classes |
|------|---------|
| Super Admin | `bg-purple-100 text-purple-800` |
| Admin | `bg-blue-100 text-blue-800` |
| Usuário | `bg-gray-100 text-gray-600` |

## Como rodar

```bash
# Da raiz do projeto (finance/)
make frontend    # npm run dev com proxy para backend

# Ou diretamente
cd frontend
npm install
npm run dev      # Dev server em http://localhost:5173
npm run build    # Build de produção
npx tsc --noEmit # Type-check
```

Para testar subdomínios em dev, acesse `financial.localhost:5173`.
