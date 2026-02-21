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

- Tenant identificado via **JWT claims** (não por subdomínio)
- **Login em 2 etapas:** login global → se usuário pertence a múltiplos tenants, exibe seletor → JWT final
- **Auto-registro:** tela de registro cria conta global + tenant automaticamente
- **Convites:** tela de aceite de convite para ingressar em tenant existente
- Sidebar mostra links de admin condicionais por role (admin/owner)
- `AdminRoute` protege rotas de administração (admin ou owner)

## Estrutura de diretórios

```
src/
├── pages/
│   ├── Login.tsx          # Login global (email/senha) + seletor de tenant
│   ├── Register.tsx       # Auto-registro (conta + tenant)
│   ├── VerifyEmail.tsx    # Verificação de email (via token)
│   ├── AcceptInvite.tsx   # Aceitar convite para tenant
│   ├── Dashboard.tsx      # Resumo financeiro com gráficos
│   ├── Income.tsx         # Wrapper → TransactionPage(type="income")
│   ├── Expense.tsx        # Wrapper → TransactionPage(type="expense")
│   ├── Categories.tsx     # Gestão de categorias (tree view)
│   ├── ExpenseLimits.tsx  # Tetos de gasto mensais
│   ├── RecurringTransactions.tsx  # Transações recorrentes (pause/resume)
│   ├── Profile.tsx        # Editar perfil e senha
│   └── Users.tsx          # Gestão de usuários (admin/owner)
├── components/
│   ├── auth/
│   │   ├── ProtectedRoute.tsx   # Guard de rota autenticada
│   │   └── AdminRoute.tsx       # Guard de rota admin
│   ├── layout/
│   │   └── AppLayout.tsx        # Layout com sidebar colapsável/responsiva (desktop collapse + mobile overlay)
│   ├── transactions/
│   │   └── TransactionPage.tsx  # Tabela de transações (desktop) + cards (mobile)
│   └── ui/                      # Componentes reutilizáveis
│       ├── Modal.tsx
│       ├── ConfirmDialog.tsx
│       ├── MonthSelector.tsx
│       ├── Autocomplete.tsx
│       └── Pagination.tsx
├── services/
│   ├── api.ts             # Instância axios, interceptors
│   ├── auth.ts            # Login, registro, verificação, convites, seletor de tenant, perfil
│   ├── categories.ts      # CRUD de categorias
│   ├── transactions.ts    # CRUD de transações
│   ├── expense-limits.ts  # CRUD de tetos
│   ├── dashboard.ts       # Summary, by-category, limits-progress
│   └── admin.ts           # Gestão de usuários + convites (admin/owner)
├── contexts/
│   └── AuthContext.tsx     # Estado de auth (token + user com role)
├── types/
│   └── index.ts           # User, LoginResponse, SelectTenantResponse, TenantInfo, InviteInfo, AdminUser, etc.
├── App.tsx                # Router + QueryClient
├── main.tsx               # Entrypoint
└── index.css              # @import "tailwindcss" + overflow-x fix
```

## Rotas

| Path | Componente | Descrição | Auth | Role |
|------|-----------|-----------|:----:|:----:|
| `/login` | Login | Login + seletor de tenant | Não | — |
| `/register` | Register | Criar conta + tenant | Não | — |
| `/verify-email` | VerifyEmail | Verificar email | Não | — |
| `/accept-invite` | AcceptInvite | Aceitar convite | Não | — |
| `/` | Dashboard | Resumo financeiro | Sim | — |
| `/income` | Income | Tabela de receitas | Sim | — |
| `/expenses` | Expense | Tabela de despesas | Sim | — |
| `/categories` | Categories | Gestão de categorias | Sim | — |
| `/expense-limits` | ExpenseLimits | Tetos de gasto | Sim | — |
| `/recurring` | RecurringTransactions | Transações recorrentes | Sim | — |
| `/profile` | Profile | Editar perfil | Sim | — |
| `/admin/users` | Users | Gestão de usuários | Sim | admin/owner |

## Services (camada de API)

### `api.ts` — Instância Axios

- `baseURL: '/api/v1'`
- **Request interceptor**: injeta `Authorization: Bearer <token>` do localStorage
- **Response interceptor**: em caso de 401 (exceto login), limpa localStorage e redireciona para `/login`
- **Proxy (dev)**: Vite redireciona `/api` → `http://localhost:8080`

### Services disponíveis

| Service | Endpoints |
|---------|-----------|
| `authService` | login, selectTenant, register, verifyEmail, getInviteInfo, acceptInvite, getProfile, updateProfile, changePassword |
| `categoryService` | list, create, update, delete |
| `transactionService` | list, getById, create, update, delete |
| `expenseLimitService` | list, create, update, delete, copy |
| `recurringTransactionService` | list, create, delete, pause, resume |
| `dashboardService` | summary, byCategory, limitsProgress |
| `adminService` | listUsers, createUser, updateUser, deleteUser, resetPassword, inviteUser |

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
| Owner | `bg-purple-100 text-purple-800` |
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

## Responsividade

- **Breakpoint principal:** `md` (768px) — abaixo é mobile, acima é desktop
- **Sidebar:** colapsável no desktop (expandida `w-64` / colapsada `w-16` só ícones), overlay no mobile com hamburger menu
- **Tabelas → Cards:** páginas com tabela usam `hidden md:block` (tabela desktop) + `md:hidden` (cards mobile)
- **FAB:** botão "Adicionar" vira circular fixo (`fixed bottom-6 right-6`) no mobile
- **MonthSelector:** tamanho e espaçamento adaptáveis (`min-w-[140px]` mobile, `sm:min-w-[180px]` desktop)
- **Overflow:** `html, body { overflow-x: hidden }` em `index.css` para evitar scroll horizontal no mobile
