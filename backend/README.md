# Finance Backend

API REST para gestão de finanças pessoais.

## Tecnologias

- **Go 1.25** — linguagem principal
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
├── domain/              → Regras de negócio (sem dependências externas)
│   ├── entity/          → Entidades de domínio
│   ├── repository/      → Interfaces dos repositórios
│   ├── usecase/         → Casos de uso
│   └── errors.go        → Erros de domínio
└── infrastructure/      → Implementações concretas
    ├── database/        → Repositórios PostgreSQL
    └── http/
        ├── handler/     → HTTP handlers
        ├── middleware/   → Auth JWT e CORS
        └── router/      → Configuração de rotas
```

**Fluxo de uma requisição:**
HTTP Request → Router → Middleware → Handler → UseCase → Repository Interface → Database Implementation

## Entidades

### User
Usuário do sistema com autenticação por email/senha.

### Transaction
Transação financeira (receita ou despesa) com valor, descrição, data e categoria. Suporta paginação e filtros por tipo, categoria e período.

### Category
Categoria de transação. Suporta hierarquia (subcategorias via `parent_id`). Tipos: `income`, `expense`, `both`. Categorias padrão (sem `user_id`) são compartilhadas entre todos os usuários.

### ExpenseLimit
Teto de gasto mensal — pode ser global (sem `category_id`) ou por categoria. Inclui cálculo de progresso (gasto, restante, percentual).

### DashboardSummary / CategoryTotal
Agregações para o dashboard: totais de receita/despesa/saldo e totais por categoria.

## Endpoints da API

Base: `/api/v1`

### Auth (público)

| Método | Rota | Descrição |
|--------|------|-----------|
| POST | `/auth/register` | Registrar novo usuário |
| POST | `/auth/login` | Login (retorna JWT) |

### Perfil (autenticado)

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/profile` | Dados do usuário logado |
| PUT | `/profile` | Atualizar nome/email |
| POST | `/profile/change-password` | Alterar senha |

### Categorias (autenticado)

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/categories` | Listar categorias (`?type=`, `?view=flat\|tree`) |
| POST | `/categories` | Criar categoria |
| PUT | `/categories/:id` | Atualizar categoria |
| DELETE | `/categories/:id` | Excluir categoria |

### Transações (autenticado)

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/transactions` | Listar (`?type=`, `?category_id=`, `?start_date=`, `?end_date=`, `?page=`, `?per_page=`) |
| GET | `/transactions/:id` | Buscar por ID |
| POST | `/transactions` | Criar transação |
| PUT | `/transactions/:id` | Atualizar transação |
| DELETE | `/transactions/:id` | Excluir transação |

### Tetos de gastos (autenticado)

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/expense-limits` | Listar (`?month=`, `?year=`) |
| POST | `/expense-limits` | Criar teto |
| PUT | `/expense-limits/:id` | Atualizar teto |
| DELETE | `/expense-limits/:id` | Excluir teto |

### Dashboard (autenticado)

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/dashboard/summary` | Resumo do mês (`?month=`, `?year=`) |
| GET | `/dashboard/by-category` | Totais por categoria (`?month=`, `?year=`, `?type=`) |
| GET | `/dashboard/limits-progress` | Progresso dos tetos (`?month=`, `?year=`) |

## Configuração

Variáveis de ambiente (arquivo `.env` na raiz do backend):

| Variável | Obrigatória | Descrição |
|----------|:-----------:|-----------|
| `DATABASE_URL` | Sim | String de conexão PostgreSQL |
| `JWT_SECRET` | Sim | Chave para assinar tokens JWT |
| `PORT` | Não | Porta do servidor (padrão: `8080`) |

## Como rodar

```bash
# Da raiz do projeto (finance/)
make db          # Sobe PostgreSQL via Docker
make migrate     # Executa migrations
make dev         # Roda com hot-reload (air)
make run         # Roda sem hot-reload
```

## Migrations

| Migration | Descrição |
|-----------|-----------|
| `001_schema` | Cria tabelas: users, categories, transactions, expense_limits |
| `002_seed` | Insere categorias padrão (Alimentação, Transporte, Salário, etc.) |
| `003_category_hierarchy` | Adiciona `parent_id` às categorias para suporte a subcategorias |

## Erros de domínio

| Erro | HTTP Status |
|------|:-----------:|
| `ErrNotFound` | 404 |
| `ErrInvalidCredentials` | 401 |
| `ErrForbidden` | 403 |
| `ErrDuplicateEmail` | 409 |
| `ErrDuplicateCategory` | 409 |
| `ErrDuplicateLimit` | 409 |
| `ErrCategoryInUse` | 409 |
| `ErrCyclicCategory` | 400 |
| `ErrInvalidPassword` | 400 |
