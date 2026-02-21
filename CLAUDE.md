# DNA Fami — Guia para Claude Code

## Visão geral

**DNA Fami** é um app de finanças **multi-tenant** com **Go backend** + **React frontend** + **PostgreSQL**. Permite cadastrar receitas/despesas, organizar por categorias hierárquicas, definir tetos de gastos mensais, criar transações recorrentes com parcelamento automático e visualizar resumos no dashboard. Dados são isolados por **tenant** (schema PostgreSQL).

- **Logo:** dupla hélice de DNA minimalista em azul (`#2563EB`), localizada em `frontend/public/logo.svg`

## Multi-Tenancy

- **Schema-per-tenant:** cada tenant tem seu próprio schema PostgreSQL (ex: `tenant_minha_familia`)
- **Global users:** tabela `public.global_users` para autenticação centralizada (email/senha + verificação)
- **Per-schema users:** tabela `{schema}.users` com FK para `global_user_id` — mantém FKs de transações intactas
- **Memberships:** tabela `public.memberships` vincula global_user → tenant (permite multi-tenant por usuário)
- **JWT `sub`** continua sendo o per-schema user_id; claims incluem `global_user_id`, `tenant_id`, `role`
- **Isolamento:** middleware `SchemaConn` configura `SET search_path` por request via `ConnFromContext`
- **Novo tenant:** criado via self-registration (POST /auth/register) — app cria schema + migrations dinamicamente
- **3 roles:** `owner` (criador, único por tenant, irremovível), `admin`, `user`
- **Self-registration:** cria conta global + tenant + schema automaticamente
- **Convites:** admin/owner convida por email → convidado aceita via link (cria conta se necessário)

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
│   │   ├── config/          # Variáveis de ambiente (DB, JWT, SendGrid, APP_URL)
│   │   ├── tenant/          # Context helpers (ContextWithSchema, SchemaFromContext)
│   │   ├── domain/
│   │   │   ├── entity/      # Entidades (User, Tenant, GlobalUser, Membership, Invite, Category, Transaction, ExpenseLimit, RecurringTransaction)
│   │   │   ├── repository/  # Interfaces dos repositórios
│   │   │   └── usecase/     # Casos de uso (auth, registration, invite, admin, category, transaction, expense_limit, recurring_transaction, dashboard)
│   │   └── infrastructure/
│   │       ├── database/    # Implementação PostgreSQL (pgx), SchemaManager, TenantCache, ConnFromContext
│   │       ├── email/       # Email sender (SendGrid API + LogSender para dev) + templates HTML
│   │       └── http/        # Handlers, middleware (auth, cors, role, schema), router (Gin)
│   ├── migrations/          # Public migrations (tenants, global_users, memberships, invites)
│   └── tenant_migrations/   # Per-tenant migrations (users, categories, transactions, expense_limits, recurring_transactions)
├── frontend/         # React SPA
│   └── src/
│       ├── pages/           # Login, Register, VerifyEmail, AcceptInvite, Dashboard, Income, Expense, Categories, ExpenseLimits, RecurringTransactions, Profile, Users
│       ├── components/      # Layout, auth (ProtectedRoute, AdminRoute), UI
│       ├── services/        # API (auth, categories, transactions, dashboard, expense-limits, recurring-transactions, admin)
│       ├── contexts/        # AuthContext
│       └── types/           # TypeScript interfaces
├── .github/workflows/ # CI/CD (lint, build, deploy, infra)
├── Dockerfile        # Multi-stage: frontend + backend
├── Makefile
└── docker-compose.yml
```

## API Endpoints

### Públicos
- `POST /api/v1/auth/login` — login global (email, password) → token ou selector
- `POST /api/v1/auth/select-tenant` — seleciona tenant (selector_token, tenant_id) → JWT
- `POST /api/v1/auth/register` — cria conta + tenant (name, email, password, tenant_name)
- `POST /api/v1/auth/verify-email` — verifica email (token)
- `GET /api/v1/auth/invite-info` — info do convite (?token=xxx)
- `POST /api/v1/auth/accept-invite` — aceita convite (token, name?, password?)

### Protegidos (JWT)
- Profile: `GET/PUT /profile`, `POST /profile/change-password`
- Categories: `GET/POST /categories`, `PUT/DELETE /categories/:id`
- Transactions: `GET/POST /transactions`, `GET/PUT/DELETE /transactions/:id`
- Expense Limits: `GET/POST /expense-limits`, `POST /expense-limits/copy`, `PUT/DELETE /expense-limits/:id`
- Recurring Transactions: `GET/POST /recurring-transactions`, `DELETE /recurring-transactions/:id`, `POST /recurring-transactions/:id/pause`, `POST /recurring-transactions/:id/resume`
- Dashboard: `GET /dashboard/summary`, `/dashboard/by-category`, `/dashboard/limits-progress`

### Admin (role: admin ou owner)
- `GET/POST /admin/users`, `PUT/DELETE /admin/users/:id`, `POST /admin/users/:id/reset-password`
- `POST /admin/invite` — envia convite por email (email, role)

## Fluxo de autenticação

### Login em duas etapas
1. `POST /auth/login {email, password}` → valida contra `global_users`
2. Se 1 tenant → auto-select: retorna `{token, user}`
3. Se N tenants → retorna `{selector_token, tenants: [{tenant_id, tenant_name, role}]}`
4. `POST /auth/select-tenant {selector_token, tenant_id}` → retorna `{token, user}`

### Registro
1. `POST /auth/register {name, email, password, tenant_name}`
2. Cria global_user (email_verified=false), tenant, schema, per-schema user, membership
3. Envia email de verificação → `POST /auth/verify-email {token}`

### JWT Claims
```json
{
  "sub": "per-schema-user-id",
  "tenant_id": "...",
  "global_user_id": "...",
  "role": "owner|admin|user",
  "exp": "...", "iat": "..."
}
```

## Recorrências e Parcelamento

- **Modos de recorrência:** Indefinido (sem fim), Data final (até uma data), Número de parcelas (`MaxOccurrences`)
- **Parcelamento (Número de parcelas):** `Amount` na entidade = valor total; na geração, o usecase divide em parcelas iguais (`Amount / MaxOccurrences`)
- **Arredondamento:** última parcela absorve a diferença de centavos para que a soma seja exata
- **Descrição das parcelas:** `"Descrição - Parcela X/N"` (ex: `"Aluguel - Parcela 3/12"`)
- **Listagem:** exibe o valor total com detalhe das parcelas: `"R$ 1.200,00 (12x R$ 100,00)"`
- **Formulário:** quando modo = "Número de parcelas", label muda para "Valor total" com preview do valor por parcela abaixo do input
- **Modos "Indefinido" e "Data final"** não são afetados — `Amount` continua sendo o valor de cada ocorrência

## Convenções de código

### Backend

- **Linguagem:** Go 1.26
- **Framework HTTP:** Gin
- **Banco:** PostgreSQL via pgx/v5 (sem ORM)
- **Auth:** JWT (golang-jwt/v5) com claims: sub, tenant_id, global_user_id, role
- **Arquitetura:** Clean Architecture — entity → repository interface → usecase → handler
- **Erros de domínio** definidos em `internal/domain/errors.go` e mapeados para HTTP status nos handlers
- **Acesso ao banco (per-schema):** repositórios obtêm a conexão via `database.ConnFromContext(ctx)` — a conexão já vem com `search_path` configurado pelo middleware `SchemaConn`
- **Acesso ao banco (global):** `GlobalUserRepo`, `MembershipRepo`, `InviteRepo` usam o pool direto (schema public)
- **`AcquireWithSchema`:** usado pelo middleware `SchemaConn`, login handler, e registration flow
- **Email:** interface `Sender` com `SendGridSender` (via `sendgrid-go` SDK) + `LogSender` (fallback se `SENDGRID_API_KEY` vazio — logs no stdout)
- **Config de email:** `SENDGRID_API_KEY` (API key do SendGrid) + `EMAIL_FROM` (endereço remetente, ex: `noreply@dnafami.com.br`)
- **CORS:** controlado por `ALLOWED_ORIGIN` — exact match + localhost. Se vazio ou `*`, aceita qualquer origin

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
