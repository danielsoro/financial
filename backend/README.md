# Finance Backend

API REST multi-tenant para gestão de finanças.

## Tecnologias

- **Go 1.26** — linguagem principal
- **Gin** — framework HTTP
- **pgx/v5** — driver PostgreSQL (sem ORM)
- **golang-jwt/v5** — autenticação JWT
- **golang-migrate/v4** — migrations de banco de dados
- **godotenv** — carregamento de variáveis de ambiente
- **Air** — hot-reload em desenvolvimento

## Arquitetura

Clean Architecture com 3 camadas:

```
cmd/api/main.go          → Bootstrap e injeção de dependências
internal/
├── config/              → Configuração (env vars)
├── tenant/              → Context helpers (ContextWithSchema, SchemaFromContext)
├── domain/              → Regras de negócio (sem dependências externas)
│   ├── entity/          → Entidades de domínio (User, Tenant, Category, Transaction, ExpenseLimit)
│   ├── repository/      → Interfaces dos repositórios
│   ├── usecase/         → Casos de uso (auth, admin, category, transaction, expense_limit, dashboard)
│   └── errors.go        → Erros de domínio
└── infrastructure/      → Implementações concretas
    ├── database/        → Repositórios PostgreSQL, SchemaManager, TenantCache, AcquireWithSchema
    └── http/
        ├── handler/     → HTTP handlers (auth, admin, category, transaction, expense_limit, dashboard)
        ├── middleware/   → Auth JWT (com schema injection), CORS, Role (RequireAdmin)
        └── router/      → Configuração de rotas
```

**Fluxo de uma requisição:**
HTTP Request → Router → Middleware (CORS → Auth [JWT + TenantCache → schema context] → Role) → Handler → UseCase → Repository [AcquireWithSchema → SET search_path] → Database

## Multi-Tenancy

- **Schema-per-tenant:** cada tenant tem seu próprio schema PostgreSQL (ex: `tenant_root`, `tenant_financial`)
- **Tenant** é identificado por subdomínio (enviado no login). `localhost` → tenant `root`
- **Tabela `tenants`** no schema `public` como registro central
- **Isolamento:** cada request autenticada resolve o tenant via JWT → TenantCache → `SET search_path`
- **JWT claims:** `sub` (user_id), `tenant_id`, `role`
- **Startup:** `RunMigrations` → `EnsureTenantsFromEnv` → `SchemaManager.InitAllTenants` → `TenantCache.Load`
- **Novo tenant:** adicionar ao env `TENANTS` → app cria schema + migrations + seed admin no startup
- **2 roles:** `admin` (gerencia usuários do tenant), `user`
- **Admin padrão:** `admin@admin.com` / `admin123` (seed automático por schema)

## Entidades

### Tenant
Organização/empresa. Campos: id, name, domain (unique), schema_name (unique), is_active, timestamps. Armazenado no schema `public`.

### User
Usuário do sistema com role (admin/user), autenticação por email/senha. Armazenado no schema do tenant.

### Transaction
Transação financeira (receita ou despesa) com user_id, valor, descrição, data e categoria. Armazenada no schema do tenant.

### Category
Categoria de transação. Suporta hierarquia (subcategorias via `parent_id`). Tipos: `income`, `expense`, `both`. Armazenada no schema do tenant.

### ExpenseLimit
Teto de gasto mensal — pode ser global (sem `category_id`) ou por categoria. Armazenado no schema do tenant.

### DashboardSummary / CategoryTotal
Agregações para o dashboard: totais de receita/despesa/saldo e totais por categoria.

## Endpoints da API

Base: `/api/v1`

### Auth (público)

| Método | Rota | Descrição |
|--------|------|-----------|
| POST | `/auth/login` | Login com email, password, subdomain (retorna JWT) |

### Perfil (autenticado)

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/profile` | Dados do usuário logado |
| PUT | `/profile` | Atualizar nome/email |
| POST | `/profile/change-password` | Alterar senha |

### Categorias (autenticado)

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/categories` | Listar categorias do tenant (`?type=`, `?view=flat\|tree`) |
| POST | `/categories` | Criar categoria |
| PUT | `/categories/:id` | Atualizar categoria |
| DELETE | `/categories/:id` | Excluir categoria |

### Transações (autenticado)

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/transactions` | Listar do tenant (`?type=`, `?category_id=`, `?start_date=`, `?end_date=`, `?page=`, `?per_page=`) |
| GET | `/transactions/:id` | Buscar por ID |
| POST | `/transactions` | Criar transação |
| PUT | `/transactions/:id` | Atualizar transação |
| DELETE | `/transactions/:id` | Excluir transação |

### Tetos de gastos (autenticado)

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/expense-limits` | Listar do tenant (`?month=`, `?year=`) |
| POST | `/expense-limits` | Criar teto |
| POST | `/expense-limits/copy` | Copiar tetos de um mês para outro |
| PUT | `/expense-limits/:id` | Atualizar teto |
| DELETE | `/expense-limits/:id` | Excluir teto |

### Dashboard (autenticado)

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/dashboard/summary` | Resumo do mês para o tenant |
| GET | `/dashboard/by-category` | Totais por categoria |
| GET | `/dashboard/limits-progress` | Progresso dos tetos |

### Admin (role: admin)

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/admin/users` | Listar usuários do tenant |
| POST | `/admin/users` | Criar usuário (name, email, password, role) |
| PUT | `/admin/users/:id` | Atualizar usuário (name, email, role) |
| DELETE | `/admin/users/:id` | Excluir usuário |
| POST | `/admin/users/:id/reset-password` | Redefinir senha |

## Configuração

Variáveis de ambiente (arquivo `.env` na raiz do backend):

| Variável | Obrigatória | Descrição |
|----------|:-----------:|-----------|
| `DATABASE_URL` | Sim | String de conexão PostgreSQL |
| `JWT_SECRET` | Sim | Chave para assinar tokens JWT |
| `PORT` | Não | Porta do servidor (padrão: `8080`) |
| `TENANTS` | Sim | Lista de tenants separados por vírgula (ex: `root,financial`). O app cria schemas no startup |
| `ALLOWED_ORIGIN` | Não | Domínio para CORS (ex: `dnafami.com.br`). Aceita o domínio e subdomínios (`*.dnafami.com.br`). Se vazio ou `*`, aceita qualquer origin. `localhost` sempre é permitido |

## Como rodar

```bash
# Da raiz do projeto (finance/)
make db          # Sobe PostgreSQL via Docker
make migrate     # Executa migrations
make dev         # Roda com hot-reload (air)
make run         # Roda sem hot-reload
```

## Migrations

### Public (`migrations/`)

| Migration | Descrição |
|-----------|-----------|
| `001_tenants` | Cria tabela `tenants` no schema `public` (registro central de tenants) |

### Per-tenant (`tenant_migrations/`)

Executadas automaticamente pelo `SchemaManager` para cada tenant ativo no startup.

| Migration | Descrição |
|-----------|-----------|
| `001_schema` | Cria tabelas: users, categories (com hierarquia), transactions, expense_limits |
| `002_seed` | Insere categorias padrão (Alimentação, Transporte, Moradia, Saúde, Educação, Lazer, Salário, Freelance, Investimentos, Outros) |

## Erros de domínio

| Erro | HTTP Status |
|------|:-----------:|
| `ErrNotFound` | 404 |
| `ErrTenantNotFound` | 404 |
| `ErrInvalidCredentials` | 401 |
| `ErrForbidden` | 403 |
| `ErrDuplicateEmail` | 409 |
| `ErrDuplicateCategory` | 409 |
| `ErrDuplicateLimit` | 409 |
| `ErrDuplicateDomain` | 409 |
| `ErrCategoryInUse` | 409 |
| `ErrCyclicCategory` | 400 |
| `ErrInvalidPassword` | 400 |
| `ErrInvalidRole` | 400 |
| `ErrSameMonth` | 400 |
