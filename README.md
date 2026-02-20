# Finance

App multi-tenant de gestao de financas pessoais. Permite cadastrar receitas e despesas, organizar por categorias hierarquicas, definir tetos de gastos mensais e visualizar resumos no dashboard. Dados sao isolados por tenant (subdominio).

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

Acesse `http://financial.localhost:5173`

Login padrao: `admin@admin.com` / `admin123`

## Estrutura do projeto

```
finance/
├── backend/
│   ├── cmd/api/             # Entrypoint
│   ├── internal/
│   │   ├── config/          # Variaveis de ambiente
│   │   ├── domain/
│   │   │   ├── entity/      # Entidades (User, Tenant, Category, Transaction, ExpenseLimit)
│   │   │   ├── repository/  # Interfaces dos repositorios
│   │   │   ├── usecase/     # Casos de uso
│   │   │   └── errors.go    # Erros de dominio
│   │   └── infrastructure/
│   │       ├── database/    # Implementacao PostgreSQL (pgx), SchemaManager, TenantCache
│   │       └── http/        # Handlers, middleware, router (Gin)
│   ├── migrations/          # Public migrations (tabela tenants)
│   └── tenant_migrations/   # Per-tenant migrations (users, categories, transactions, expense_limits)
├── frontend/
│   └── src/
│       ├── pages/           # Dashboard, Income, Expense, Categories, ExpenseLimits, Profile, Users
│       ├── components/      # Layout, auth, UI reutilizaveis
│       ├── services/        # Clientes da API
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
| Auth | `POST /auth/login` |
| Profile | `GET/PUT /profile`, `POST /profile/change-password` |
| Categories | `GET/POST /categories`, `PUT/DELETE /categories/:id` |
| Transactions | `GET/POST /transactions`, `GET/PUT/DELETE /transactions/:id` |
| Expense Limits | `GET/POST /expense-limits`, `POST /expense-limits/copy`, `PUT/DELETE /expense-limits/:id` |
| Dashboard | `GET /dashboard/summary`, `/by-category`, `/limits-progress` |
| Admin | `GET/POST /admin/users`, `PUT/DELETE /admin/users/:id`, `POST /admin/users/:id/reset-password` |

## Multi-Tenancy

- **Schema-per-tenant:** cada tenant tem seu proprio schema PostgreSQL (ex: `tenant_root`, `tenant_financial`)
- Tenant identificado por subdominio (`financial.localhost` -> tenant `financial`)
- Tabela `tenants` no schema `public` como registro central (id, name, domain, schema_name, is_active)
- Isolamento por schema; queries usam `SET search_path` por conexao
- 2 roles: `admin` (gerencia usuarios do tenant) e `user`
- Somente admin cria usuarios — nao ha registro publico
- Admin padrao por tenant: `admin@admin.com` / `admin123` (seed automatico)
- Novo tenant: adicionar ao env `TENANTS` (ou `var.tenants` no Terraform) → app cria schema + seed no startup

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
- **Secret Manager** — JWT secret
- **Workload Identity Federation** — autenticacao segura do GitHub Actions (sem chaves)
- **Cloudflare DNS** — registros DNS por tenant (CNAME → Cloud Run, com proxy/CDN)

### DNS (Cloudflare)

Cada tenant tem um subdominio explicito gerenciado via Terraform (`cloudflare.tf`). Para adicionar um novo tenant, inclua o nome na variavel `tenants` e rode o pipeline:

```hcl
# deploy/terraform.tfvars
tenants = ["financial", "novo-tenant"]
```

Variaveis necessarias (GitHub Secrets):

| Secret | Descricao |
|--------|-----------|
| `CLOUDFLARE_API_TOKEN` | Token da API com permissoes Zone:Read, DNS:Edit, Zone Settings:Edit |
| `DOMAIN` | Dominio raiz (ex: `dnafami.com.br`) |
| `TENANTS` | Lista de subdominos em formato HCL (ex: `["financial"]`) |

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
