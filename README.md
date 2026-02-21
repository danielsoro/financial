<p align="center">
  <img src="frontend/public/assets/logo.svg" alt="DNA Fami" width="64" height="64" />
</p>

<h1 align="center">DNA Fami</h1>

<p align="center">
  App multi-tenant de gestao de financas pessoais. Permite cadastrar receitas e despesas, organizar por categorias hierarquicas, definir tetos de gastos mensais, criar transacoes recorrentes e visualizar resumos no dashboard. Dados sao isolados por tenant (schema PostgreSQL). Suporta auto-registro, verificacao de email e convites.
</p>

## Tech Stack

**Backend:** Go 1.26 | Gin | pgx/v5 | JWT | golang-migrate

**Frontend:** React 19 | TypeScript 5.9 | Vite 7 | Tailwind CSS v4 | TanStack Query v5

**Infra:** PostgreSQL 16 | Docker | Google Cloud Run | Cloud SQL | Cloudflare DNS | Terraform

## Pre-requisitos

- Go 1.26+
- Node.js 22+
- Docker e Docker Compose

## Como rodar

```bash
# 1. Subir o banco
make db

# 2. Rodar migrations
make migrate

# 3. Backend (com hot-reload)
make dev

# 4. Frontend (em outro terminal)
make frontend
```

Acesse `http://localhost:5173`

Crie sua conta pela tela de registro.

## Estrutura do projeto

```
finance/
├── backend/
│   ├── cmd/api/             # Entrypoint
│   ├── internal/
│   │   ├── config/          # Variaveis de ambiente
│   │   ├── domain/
│   │   │   ├── entity/      # Entidades (User, Tenant, GlobalUser, Membership, Invite, Category, Transaction, ExpenseLimit, RecurringTransaction)
│   │   │   ├── repository/  # Interfaces dos repositorios
│   │   │   ├── usecase/     # Casos de uso
│   │   │   └── errors.go    # Erros de dominio
│   │   └── infrastructure/
│   │       ├── database/    # Implementacao PostgreSQL (pgx), SchemaManager, TenantCache
│   │       ├── email/       # Email sender (SendGrid API + LogSender para dev)
│   │       └── http/        # Handlers, middleware, router (Gin)
│   ├── migrations/          # Public migrations (tenants, global_users, memberships, invites)
│   └── tenant_migrations/   # Per-tenant migrations (users, categories, transactions, expense_limits, recurring_transactions)
├── frontend/
│   └── src/
│       ├── pages/           # Login, Register, VerifyEmail, AcceptInvite, Dashboard, Income, Expense, Categories, ExpenseLimits, RecurringTransactions, Profile, Users
│       ├── components/      # Layout, auth, UI reutilizaveis
│       ├── services/        # Clientes da API (auth, categories, transactions, dashboard, expense-limits, recurring-transactions, admin)
│       ├── contexts/        # AuthContext
│       └── types/           # Interfaces TypeScript
├── deploy/                  # Terraform (Cloud Run, Cloud SQL, IAM, Secrets, Cloudflare DNS)
├── .github/workflows/       # CI/CD (lint, build, deploy, infra)
├── Dockerfile               # Multi-stage: frontend + backend
├── docker-compose.yml       # Postgres, pgAdmin (dev local)
└── Makefile
```

## API Endpoints

Base: `/api/v1`

| Grupo | Endpoints |
|-------|-----------|
| Health | `GET /health` |
| Auth | `POST /auth/login`, `POST /auth/select-tenant`, `POST /auth/register`, `POST /auth/verify-email`, `GET /auth/invite-info`, `POST /auth/accept-invite` |
| Profile | `GET/PUT /profile`, `POST /profile/change-password` |
| Categories | `GET/POST /categories`, `PUT/DELETE /categories/:id` |
| Transactions | `GET/POST /transactions`, `GET/PUT/DELETE /transactions/:id` |
| Expense Limits | `GET/POST /expense-limits`, `POST /expense-limits/copy`, `PUT/DELETE /expense-limits/:id` |
| Recurring Transactions | `GET/POST /recurring-transactions`, `DELETE /recurring-transactions/:id`, `POST /recurring-transactions/:id/pause`, `POST /recurring-transactions/:id/resume` |
| Dashboard | `GET /dashboard/summary`, `/by-category`, `/limits-progress` |
| Admin | `GET/POST /admin/users`, `PUT/DELETE /admin/users/:id`, `POST /admin/users/:id/reset-password`, `POST /admin/invite` |

## Multi-Tenancy

- **Schema-per-tenant:** cada tenant tem seu proprio schema PostgreSQL (ex: `tenant_minha_familia`)
- **Global users:** tabela `public.global_users` para autenticacao centralizada (email/senha + verificacao de email)
- **Per-schema users:** tabela `{schema}.users` com FK para `global_user_id` — mantem FKs de transacoes intactas
- **Memberships:** tabela `public.memberships` vincula global_user → tenant (permite multi-tenant por usuario)
- Tenant identificado via JWT claims (nao por subdominio)
- **Login em 2 etapas:** login global → se multi-tenant, seleciona tenant → JWT final
- **3 roles:** `owner` (criador, unico por tenant, irremovivel), `admin`, `user`
- **Self-registration:** `POST /auth/register` cria conta global + tenant + schema automaticamente
- **Verificacao de email:** registro envia email de verificacao; login requer email verificado
- **Convites:** admin/owner convida por email → convidado aceita via link (cria conta se necessario)

## CI/CD

| Workflow | Trigger | O que faz |
|----------|---------|-----------|
| `ci-frontend` | Pull request | Lint (ESLint), type-check (tsc), build (Vite) |
| `ci-backend` | Pull request | Vet, build |
| `deploy` | Tag `v*` | Build Docker, push para Artifact Registry, deploy no Cloud Run |
| `infra` | Push em `deploy/` ou manual | Terraform plan + apply |

PRs sao bloqueados para merge caso os checks de CI falhem.

## Deploy (GCP)

A infraestrutura e gerenciada via Terraform (`deploy/`):

- **Cloud Run** — app serverless (0-2 instancias)
- **Cloud SQL** — PostgreSQL 16
- **Artifact Registry** — imagens Docker
- **Secret Manager** — JWT secret, SendGrid API key
- **Workload Identity Federation** — autenticacao segura do GitHub Actions (sem chaves)
- **Cloudflare DNS** — CNAME para Cloud Run

### DNS (Cloudflare)

DNS gerenciado via Terraform (`cloudflare.tf`). Inclui CNAME para o Cloud Run.

Variaveis necessarias (GitHub Secrets):

| Secret | Descricao |
|--------|-----------|
| `CLOUDFLARE_API_TOKEN` | Token da API com permissoes Zone:Read, DNS:Edit, Zone Settings:Edit |
| `DOMAIN` | Dominio raiz (ex: `dnafami.com.br`) |
| `SENDGRID_API_KEY` | API key do SendGrid para envio de emails |
| `EMAIL_FROM` | Endereco remetente (ex: `noreply@dnafami.com.br`) |


### Deploy

Para fazer deploy:

```bash
git tag v1.0.0
git push --tags
```

O workflow `deploy` faz build, push e deploy automaticamente.

## Comandos uteis

| Comando | O que faz |
|---------|-----------|
| `make db` | Sobe PostgreSQL via Docker Compose |
| `make db-down` | Para o PostgreSQL |
| `make dev` | Backend com hot-reload (Air) |
| `make run` | Backend sem hot-reload |
| `make migrate` | Executa migrations pendentes |
| `make frontend` | Frontend dev server (Vite) |
| `cd frontend && npx tsc --noEmit` | Type-check do frontend |
