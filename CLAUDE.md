# DNA Fami — Guia para Claude Code

## Visão geral

**DNA Fami** é um app de finanças **multi-tenant** com **Go backend** + **React frontend** + **PostgreSQL**. Permite cadastrar receitas/despesas, organizar por categorias hierárquicas, definir tetos de gastos mensais e visualizar resumos no dashboard. Dados são isolados por **tenant** (identificado por subdomínio).

- **Logo:** dupla hélice de DNA minimalista em azul (`#2563EB`), localizada em `frontend/public/logo.svg`

## Multi-Tenancy

- **Schema-per-tenant:** cada tenant tem seu próprio schema PostgreSQL (ex: `tenant_root`, `tenant_financial`)
- **Tenant** é resolvido por subdomínio (ex: `financial.localhost` → tenant `financial`, `localhost` → tenant `root`)
- **Tabela `tenants`** no schema `public` como registro central (id, name, domain, schema_name, is_active)
- **Isolamento:** dados isolados por schema; middleware `SchemaConn` configura `SET search_path` uma vez por request e armazena a conexão no context via `ConnFromContext`
- **Novo tenant:** adicionar ao `var.tenants` no Terraform → app cria schema + seed no startup
- **2 roles:** `admin` (gerencia usuários do tenant), `user`
- **Somente admin cria usuários** — não há registro público
- **Admin padrão:** `admin@admin.com` / `admin123` (seed automático por schema)
- **Tenant padrão:** domain=`root`, name=`Root`
- **Env `TENANTS`:** lista de tenants separados por vírgula (ex: `root,financial`)

## Comandos úteis (rodar da raiz `finance/`)

| Comando | O que faz |
|---------|-----------|
| `make db` | Sobe PostgreSQL via Docker Compose |
| `make db-down` | Para o PostgreSQL |
| `make dev` | Roda backend com hot-reload (air) |
| `make run` | Roda backend sem hot-reload |
| `make migrate` | Executa migrations pendentes |
| `make frontend` | Roda frontend dev server (Vite) |
| `cd frontend && npx tsc --noEmit` | Type-check do frontend |

## Estrutura do projeto

```
finance/
├── backend/          # Go API (Clean Architecture)
│   ├── cmd/api/      # Entrypoint
│   ├── internal/
│   │   ├── config/          # Variáveis de ambiente
│   │   ├── tenant/          # Context helpers (ContextWithSchema, SchemaFromContext)
│   │   ├── domain/
│   │   │   ├── entity/      # Entidades (User, Tenant, Category, Transaction, ExpenseLimit)
│   │   │   ├── repository/  # Interfaces dos repositórios
│   │   │   └── usecase/     # Casos de uso (auth, admin, category, transaction, expense_limit, dashboard)
│   │   └── infrastructure/
│   │       ├── database/    # Implementação PostgreSQL (pgx), SchemaManager, TenantCache, ConnFromContext
│   │       └── http/        # Handlers, middleware (auth, cors, role, schema), router (Gin)
│   ├── migrations/          # Public migrations (tabela tenants)
│   └── tenant_migrations/   # Per-tenant migrations (users, categories, transactions, expense_limits)
├── frontend/         # React SPA
│   └── src/
│       ├── pages/           # Dashboard, Income, Expense, Categories, ExpenseLimits, Profile, Users
│       ├── components/      # Layout, auth (ProtectedRoute, AdminRoute), UI
│       ├── services/        # API (auth, categories, transactions, dashboard, expense-limits, admin)
│       ├── contexts/        # AuthContext
│       └── types/           # TypeScript interfaces
├── deploy/           # Terraform (Cloud Run, Cloud SQL, IAM, Secrets, Cloudflare DNS)
├── .github/workflows/ # CI/CD (lint, build, deploy, infra)
├── Dockerfile        # Multi-stage: frontend + backend
├── Makefile
└── docker-compose.yml
```

## API Endpoints

### Públicos
- `POST /api/v1/auth/login` — login com email, password, subdomain

### Protegidos (JWT)
- Profile: `GET/PUT /profile`, `POST /profile/change-password`
- Categories: `GET/POST /categories`, `PUT/DELETE /categories/:id`
- Transactions: `GET/POST /transactions`, `GET/PUT/DELETE /transactions/:id`
- Expense Limits: `GET/POST /expense-limits`, `POST /expense-limits/copy`, `PUT/DELETE /expense-limits/:id`
- Dashboard: `GET /dashboard/summary`, `/dashboard/by-category`, `/dashboard/limits-progress`

### Admin (role: admin)
- `GET/POST /admin/users`, `PUT/DELETE /admin/users/:id`, `POST /admin/users/:id/reset-password`

## Convenções de código

### Backend

- **Linguagem:** Go 1.26
- **Framework HTTP:** Gin
- **Banco:** PostgreSQL via pgx/v5 (sem ORM)
- **Auth:** JWT (golang-jwt/v5) com claims: sub, tenant_id, role
- **Arquitetura:** Clean Architecture — entity → repository interface → usecase → handler
- **Erros de domínio** definidos em `internal/domain/errors.go` e mapeados para HTTP status nos handlers
- **Acesso ao banco:** repositórios obtêm a conexão via `database.ConnFromContext(ctx)` — a conexão já vem com `search_path` configurado pelo middleware `SchemaConn`
- **`AcquireWithSchema`:** usado apenas pelo middleware `SchemaConn` e pelo login handler (que precisa resolver o tenant antes do middleware)
- **Fluxo do login:** handler resolve tenant via `ResolveTenant`, configura conexão com `AcquireWithSchema` + `ContextWithConn`, depois chama `Authenticate`
- **CORS:** controlado por `ALLOWED_ORIGIN` — aceita o domínio e todos os subdomínios (tenants). Se vazio ou `*`, aceita qualquer origin

### Frontend

- **React 19** + **TypeScript 5.9** (strict mode) + **Vite 7**
- **Tailwind CSS v4** via `@tailwindcss/vite` plugin
- **TanStack Query v5** para server state (queries e mutations)
- **React Router v7** para rotas
- **Context** apenas para auth (`AuthContext`)
- **Formulários:** controlados com `useState`, sem lib de forms
- **Notificações:** `react-hot-toast`
- **Ícones:** `react-icons/hi2`
- **Gráficos:** `recharts`

#### Layout Responsivo

- **Sidebar colapsável** — estados `collapsed` e `mobileOpen` no `AppLayout`
- **Desktop (md+):** sidebar alterna entre expandida (`w-64`) e colapsada (`w-16`, só ícones)
- **Mobile (<md):** sidebar oculta com overlay, hamburger menu no header fixo (`h-14`)
- **Botão collapse:** chevron no fundo da sidebar (desktop only)
- **Navegação mobile:** fecha sidebar ao mudar de rota (`useLocation`)
- **Main content:** `transition-[margin]`, padding responsivo (`px-4` mobile / `md:px-8` desktop)

#### Ordem de imports

1. React (`import { useState } from 'react'`)
2. Libs externas (react-router, tanstack, toast, icons, recharts)
3. Services (`../services/...`)
4. Components (`../components/...`)
5. Contexts (`../contexts/...`)
6. Types (`import type { ... } from '../types'`)

## Padrões de estilo do frontend

### Cores semânticas

| Uso | Classes Tailwind |
|-----|-----------------|
| Primary / balance | `blue-600`, `blue-700`, `blue-500` |
| Income / positivo | `green-600`, `green-100`, `green-800` |
| Expense / negativo | `red-600`, `red-100`, `red-800` |
| Warning | `yellow-500` |
| Background | `gray-50` (main), `gray-900` (sidebar) |
| Texto | `gray-900` (primary), `gray-700`, `gray-500` |

### Componentes UI reutilizáveis (`src/components/ui/`)

- **Modal** — overlay com backdrop, botão fechar
- **ConfirmDialog** — modal de confirmação (Cancelar / Confirmar em vermelho)
- **MonthSelector** — navegação mês/ano com botões anterior/próximo
- **Autocomplete** — dropdown com busca e navegação por teclado
- **Pagination** — botões Anterior/Próximo com indicador de página

### Responsividade

- **Breakpoint principal:** `md` (768px) — mobile vs desktop
- **Tabelas → Cards no mobile:** páginas com tabela renderizam `hidden md:block` (tabela desktop) + `md:hidden` (cards mobile)
- **FAB (Floating Action Button):** botão "Adicionar" vira `fixed bottom-6 right-6` circular no mobile, inline no desktop (`hidden md:flex`)
- **MonthSelector:** `min-w-[140px]` mobile, `sm:min-w-[180px]` desktop; `gap-2` mobile, `sm:gap-4` desktop
- **Overflow:** `html/body { overflow-x: hidden }` no `index.css`

### Padrão de mutation (React Query)

```tsx
const mutation = useMutation({
  mutationFn: (data) => service.method(data),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['key'] });
    closeModal();
    toast.success('Mensagem');
  },
  onError: (err: any) => toast.error(err.response?.data?.error || 'Erro'),
});
```

## Idioma

- Toda UI em **português brasileiro**
- Moeda: **BRL** (`pt-BR`, `Intl.NumberFormat`)
- Datas: formato brasileiro (`toLocaleDateString('pt-BR')`)

## Regras

- Não adicionar dependências sem necessidade explícita
- Rodar `npx tsc --noEmit` (de `frontend/`) antes de finalizar mudanças no frontend
- Manter Clean Architecture no backend: nunca importar infra no domínio
- Respeitar as cores semânticas (verde = receita, vermelho = despesa, azul = primary)
