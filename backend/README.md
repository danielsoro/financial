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
│   ├── entity/          → Entidades de domínio (User, Tenant, GlobalUser, Membership, Invite, Category, Transaction, ExpenseLimit, RecurringTransaction)
│   ├── repository/      → Interfaces dos repositórios
│   ├── usecase/         → Casos de uso (auth, registration, invite, admin, category, transaction, expense_limit, recurring_transaction, dashboard)
│   └── errors.go        → Erros de domínio
└── infrastructure/      → Implementações concretas
    ├── database/        → Repositórios PostgreSQL, SchemaManager, TenantCache, AcquireWithSchema
    ├── email/           → Email sender (SendGrid API + LogSender para dev) + templates HTML
    └── http/
        ├── handler/     → HTTP handlers (auth, registration, invite, admin, category, transaction, expense_limit, recurring_transaction, dashboard)
        ├── middleware/   → Auth JWT, CORS, Role (RequireAdmin), SchemaConn (SET search_path)
        └── router/      → Configuração de rotas
```

**Fluxo de uma requisição:**
HTTP Request → Router → Middleware (CORS → Auth [JWT + TenantCache → schema context] → Role) → Handler → UseCase → Repository [AcquireWithSchema → SET search_path] → Database

## Multi-Tenancy

- **Schema-per-tenant:** cada tenant tem seu próprio schema PostgreSQL (ex: `tenant_minha_familia`)
- **Global users:** tabela `public.global_users` para autenticação centralizada (email/senha + verificação de email)
- **Per-schema users:** tabela `{schema}.users` com FK para `global_user_id` — mantém FKs de transações intactas
- **Memberships:** tabela `public.memberships` vincula global_user → tenant (permite multi-tenant por usuário)
- **Tabela `tenants`** no schema `public` como registro central (com `owner_id` referenciando global_user)
- **Isolamento:** middleware `SchemaConn` configura `SET search_path` por request via `ConnFromContext`
- **JWT claims:** `sub` (per-schema user_id), `tenant_id`, `global_user_id`, `role`
- **Startup:** `RunMigrations` → `SchemaManager.InitAllTenants` → `TenantCache.Load`
- **Novo tenant:** criado via self-registration (`POST /auth/register`) — app cria schema + migrations dinamicamente
- **3 roles:** `owner` (criador, único por tenant, irremovível), `admin`, `user`
- **Self-registration:** cria conta global + tenant + schema automaticamente
- **Convites:** admin/owner convida por email → convidado aceita via link (cria conta se necessário)

## Entidades

### Tenant
Organização/família. Campos: id, name, domain (unique), schema_name (unique), owner_id (FK global_user), is_active, timestamps. Armazenado no schema `public`.

### GlobalUser
Usuário global para autenticação centralizada. Campos: id, name, email (unique), password_hash, email_verified, verification_token, timestamps. Armazenado no schema `public`.

### Membership
Vínculo entre global_user e tenant. Campos: id, global_user_id, tenant_id, role, timestamps. Armazenado no schema `public`.

### TenantMembership (DTO)
Projeção para seleção de tenant no login: tenant_id, tenant_name, role.

### User
Usuário do tenant com role (owner/admin/user), global_user_id (FK). Armazenado no schema do tenant.

### Invite
Convite por email para ingressar em um tenant. Campos: id, tenant_id, email, role, token, invited_by, used, expires_at, timestamps. Armazenado no schema `public`.

### InviteInfo (DTO)
Projeção pública do convite: tenant_name, email, role, inviter_name.

### Transaction
Transação financeira (receita ou despesa) com user_id, valor, descrição, data e categoria. Armazenada no schema do tenant.

### Category
Categoria de transação. Suporta hierarquia (subcategorias via `parent_id`). Tipos: `income`, `expense`, `both`. Armazenada no schema do tenant.

### ExpenseLimit
Teto de gasto mensal — pode ser global (sem `category_id`) ou por categoria. Armazenado no schema do tenant.

### RecurringTransaction
Transação recorrente com frequência (monthly/weekly/daily), modo (indefinido/data final/parcelas), pause/resume. Armazenada no schema do tenant.

### DashboardSummary / CategoryTotal
Agregações para o dashboard: totais de receita/despesa/saldo e totais por categoria.

## Endpoints da API

Base: `/api/v1`

### Auth (público)

| Método | Rota | Descrição |
|--------|------|-----------|
| POST | `/auth/login` | Login global (email, password) → JWT ou selector_token + lista de tenants |
| POST | `/auth/select-tenant` | Seleciona tenant (selector_token, tenant_id) → JWT |
| POST | `/auth/register` | Cria conta global + tenant (name, email, password, tenant_name) |
| POST | `/auth/verify-email` | Verifica email (token) |
| GET | `/auth/invite-info` | Info do convite (?token=xxx) |
| POST | `/auth/accept-invite` | Aceita convite (token, name?, password?) |

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

### Recorrências (autenticado)

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/recurring-transactions` | Listar recorrências do tenant |
| POST | `/recurring-transactions` | Criar recorrência |
| DELETE | `/recurring-transactions/:id` | Excluir recorrência |
| POST | `/recurring-transactions/:id/pause` | Pausar recorrência |
| POST | `/recurring-transactions/:id/resume` | Retomar recorrência |

### Admin (role: admin ou owner)

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/admin/users` | Listar usuários do tenant |
| POST | `/admin/users` | Criar usuário (name, email, password, role) |
| PUT | `/admin/users/:id` | Atualizar usuário (name, email, role) |
| DELETE | `/admin/users/:id` | Excluir usuário |
| POST | `/admin/users/:id/reset-password` | Redefinir senha |
| POST | `/admin/invite` | Enviar convite por email (email, role) |

## Configuração

Variáveis de ambiente (arquivo `.env` na raiz do backend):

| Variável | Obrigatória | Descrição |
|----------|:-----------:|-----------|
| `DATABASE_URL` | Sim | String de conexão PostgreSQL |
| `JWT_SECRET` | Sim | Chave para assinar tokens JWT |
| `PORT` | Não | Porta do servidor (padrão: `8080`) |
| `APP_URL` | Não | URL base da aplicação (ex: `https://dnafami.com.br`). Usada em links de emails (verificação, convites) |
| `ALLOWED_ORIGIN` | Não | Origin para CORS (exact match + localhost). Se vazio ou `*`, aceita qualquer origin |
| `SENDGRID_API_KEY` | Não | API key do SendGrid. Se vazio, usa `LogSender` (logs no stdout) |
| `EMAIL_FROM` | Não | Endereço remetente dos emails (ex: `noreply@dnafami.com.br`) |

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
| `002_global_users` | Cria tabelas `global_users`, `memberships`, `invites` no schema `public` |
| `003_tenants_add_owner` | Adiciona coluna `owner_id` na tabela `tenants` (FK para global_users) |

### Per-tenant (`tenant_migrations/`)

Executadas automaticamente pelo `SchemaManager` para cada tenant ativo no startup.

| Migration | Descrição |
|-----------|-----------|
| `001_schema` | Cria tabelas: users, categories (com hierarquia), transactions, expense_limits |
| `002_seed` | Insere categorias padrão (Alimentação, Transporte, Moradia, Saúde, Educação, Lazer, Salário, Freelance, Investimentos, Outros) |
| `003_recurring_transactions` | Cria tabela recurring_transactions |
| `004_recurring_redesign` | Redesign da tabela recurring_transactions (adiciona pause/resume, modos de recorrência) |
| `005_add_global_user_id` | Adiciona coluna `global_user_id` na tabela `users` (FK para global_users) |

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
| `ErrDuplicateTenant` | 409 |
| `ErrCategoryInUse` | 409 |
| `ErrAlreadyMember` | 409 |
| `ErrCyclicCategory` | 400 |
| `ErrInvalidPassword` | 400 |
| `ErrInvalidRole` | 400 |
| `ErrInvalidFrequency` | 400 |
| `ErrSameMonth` | 400 |
| `ErrAlreadyPaused` | 400 |
| `ErrAlreadyActive` | 400 |
| `ErrEmailNotVerified` | 403 |
| `ErrMaxTenantsReached` | 400 |
| `ErrInviteExpired` | 400 |
| `ErrInviteAlreadyUsed` | 400 |
| `ErrNoMemberships` | 400 |
