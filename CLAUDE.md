# Finance — Guia para Claude Code

## Visão geral

App de finanças **multi-tenant** com **Go backend** + **React frontend** + **PostgreSQL**. Permite cadastrar receitas/despesas, organizar por categorias hierárquicas, definir tetos de gastos mensais e visualizar resumos no dashboard. Dados são isolados por **tenant** (identificado por subdomínio).

## Multi-Tenancy

- **Tenant** é resolvido por subdomínio (ex: `financial.localhost` → tenant `financial`)
- **3 roles:** `super_admin`, `admin` (gerencia usuários do tenant), `user`
- **Somente admin cria usuários** — não há registro público
- **Filtro de dados:** `tenant_id` é o filtro primário em queries; `user_id` permanece para atribuição/auditoria
- **Super admin default:** `super@admin.com` / `admin123`
- **Tenant padrão:** domain=`financial`, name=`Financial Demo`

## Comandos úteis (rodar da raiz `finance/`)

| Comando | O que faz |
|---------|-----------|
| `make db` | Sobe PostgreSQL via Docker Compose |
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
│   │   ├── domain/
│   │   │   ├── entity/      # Entidades (User, Tenant, Category, Transaction, ExpenseLimit)
│   │   │   ├── repository/  # Interfaces dos repositórios
│   │   │   └── usecase/     # Casos de uso (auth, admin, category, transaction, expense_limit, dashboard)
│   │   └── infrastructure/
│   │       ├── database/    # Implementação PostgreSQL (pgx)
│   │       └── http/        # Handlers, middleware (auth, cors, role), router (Gin)
│   └── migrations/          # SQL migrations (golang-migrate)
├── frontend/         # React SPA
│   └── src/
│       ├── pages/           # Dashboard, Income, Expense, Categories, ExpenseLimits, Profile, Users
│       ├── components/      # Layout, auth (ProtectedRoute, AdminRoute), UI
│       ├── services/        # API (auth, categories, transactions, dashboard, expense-limits, admin)
│       ├── contexts/        # AuthContext
│       └── types/           # TypeScript interfaces
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
- Expense Limits: `GET/POST /expense-limits`, `PUT/DELETE /expense-limits/:id`
- Dashboard: `GET /dashboard/summary`, `/dashboard/by-category`, `/dashboard/limits-progress`

### Admin (role: admin ou super_admin)
- `GET/POST /admin/users`, `PUT/DELETE /admin/users/:id`, `POST /admin/users/:id/reset-password`

## Convenções de código

### Backend

- **Linguagem:** Go 1.25
- **Framework HTTP:** Gin
- **Banco:** PostgreSQL via pgx/v5 (sem ORM)
- **Auth:** JWT (golang-jwt/v5) com claims: sub, tenant_id, role
- **Arquitetura:** Clean Architecture — entity → repository interface → usecase → handler
- **Erros de domínio** definidos em `internal/domain/errors.go` e mapeados para HTTP status nos handlers

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
