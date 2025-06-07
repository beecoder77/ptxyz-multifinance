# PT XYZ Multifinance Backend Service

Backend service untuk PT XYZ Multifinance yang menangani pembiayaan White Goods, Motor, dan Mobil.

## Teknologi yang Digunakan

- Go 1.21+
- PostgreSQL 15+
- Redis (untuk caching dan rate limiting)
- Docker & Docker Compose
- JWT untuk autentikasi
- OpenTelemetry untuk observability
- Prometheus & Grafana untuk monitoring

## Arsitektur

Proyek ini menggunakan Clean Architecture dengan struktur sebagai berikut:

```
.
├── cmd/                    # Entry points aplikasi
│   └── api/               # API server
│       └── main.go        # Main application entry point
├── configs/               # File konfigurasi
│   ├── prometheus/       # Konfigurasi Prometheus
│   └── app.yaml         # Konfigurasi aplikasi
├── deployments/          # Deployment configurations
│   ├── docker/          # Docker related files
│   └── k8s/             # Kubernetes manifests
├── internal/             # Private application code
│   ├── domain/          # Enterprise business rules
│   │   ├── entity/     # Domain entities
│   │   ├── repository/ # Repository interfaces
│   │   └── service/    # Domain services
│   ├── usecase/        # Application business rules
│   │   ├── customer/   # Customer related use cases
│   │   ├── transaction/# Transaction related use cases
│   │   └── dto/       # Data Transfer Objects
│   ├── repository/     # Repository implementations
│   │   ├── postgres/  # PostgreSQL implementations
│   │   └── redis/     # Redis implementations
│   ├── delivery/       # Interface adapters
│   │   ├── http/      # HTTP handlers
│   │   └── grpc/      # gRPC handlers (if needed)
│   ├── middleware/     # HTTP middleware
│   │   ├── auth/      # Authentication middleware
│   │   ├── logging/   # Logging middleware
│   │   └── tracing/   # OpenTelemetry tracing
│   └── pkg/           # Internal shared packages
│       ├── logger/    # Logging utilities
│       ├── validator/ # Input validation
│       └── errors/    # Error handling
├── migrations/          # Database migrations
│   ├── 000001_init_database.up.sql    # Database initialization
│   ├── 000002_init_schema.up.sql      # Schema creation
│   ├── 000003_insert_dummy_data.up.sql # Test data
│   └── README.md                      # Migration documentation
├── pkg/                # Public shared packages
│   ├── config/        # Configuration utilities
│   ├── database/      # Database utilities
│   └── httpserver/    # HTTP server utilities
├── scripts/           # Development and deployment scripts
│   ├── setup.sh      # Development environment setup
│   └── deploy.sh     # Deployment scripts
├── tests/            # Integration & E2E tests
│   ├── integration/  # Integration tests
│   └── e2e/         # End-to-end tests
├── .env.example      # Example environment variables
├── docker-compose.yml # Docker compose configuration
├── Dockerfile        # Main application Dockerfile
├── Dockerfile.migrate # Database migration Dockerfile
├── go.mod           # Go modules file
├── go.sum          # Go modules checksum
└── Makefile        # Build automation
```

### Layer Descriptions

1. **Domain Layer** (`internal/domain/`)
   - Berisi aturan bisnis dan entitas
   - Tidak bergantung pada framework atau teknologi
   - Mendefinisikan interfaces untuk repository

2. **Use Case Layer** (`internal/usecase/`)
   - Implementasi business logic
   - Orchestration antara domain entities
   - Transformasi data antara domain dan delivery layer

3. **Repository Layer** (`internal/repository/`)
   - Implementasi interfaces dari domain layer
   - Handling database operations
   - Caching implementation

4. **Delivery Layer** (`internal/delivery/`)
   - HTTP/gRPC handlers
   - Request/Response handling
   - Input validation
   - Error handling

### Design Patterns

- **Repository Pattern**: Abstraksi akses data
- **Dependency Injection**: Loose coupling antar komponen
- **Factory Pattern**: Object creation
- **Middleware Pattern**: Request/Response processing
- **Builder Pattern**: Complex object construction

### Database Design

- **Tables**:
  - `customers`: Data nasabah
  - `credit_limits`: Limit kredit nasabah
  - `transactions`: Transaksi pembiayaan
  - `installments`: Cicilan pembayaran

- **Features**:
  - Soft delete
  - Audit trails (created_at, updated_at)
  - Foreign key constraints
  - Indexing untuk performance
  - Data validation constraints

## Fitur Utama

- Manajemen konsumen dan limit pembiayaan
- Pencatatan dan pemrosesan transaksi
- Integrasi dengan e-commerce dan dealer
- Sistem keamanan berbasis OWASP
- Monitoring dan observability
- High availability dan scalability

## Setup Development

1. Clone repository
```bash
git clone https://github.com/xyz-multifinance/backend.git
```

2. Copy dan sesuaikan environment variables
```bash
cp .env.example .env
```

3. Jalankan dengan Docker Compose
```bash
docker-compose up -d
```

4. Jalankan migrasi database
```bash
make migrate-up
```

## API Documentation

API documentation tersedia di:
- Swagger UI: http://localhost:8080/swagger/index.html
- Postman Collection: docs/postman/xyz-multifinance.json

## Testing

Untuk menjalankan unit test:
```bash
make test
```

Untuk menjalankan integration test:
```bash
make test-integration
```

## Git Workflow (GitFlow)

Proyek ini menggunakan GitFlow untuk version control workflow. Berikut adalah branch strategy yang digunakan:

### Main Branches

- `main` - Production branch
  * Branch yang selalu production-ready
  * Setiap commit di main adalah versi release baru
  * Tag version menggunakan semantic versioning (v1.2.3)
  * Protected branch: require code review dan CI/CD pass

- `develop` - Development branch
  * Branch utama untuk development
  * Berisi fitur-fitur yang sudah selesai untuk release berikutnya
  * Protected branch: require code review
  * Auto-deploy ke environment staging

### Supporting Branches

- `feature/*` - Feature branches
  * Dibuat dari: `develop`
  * Merge ke: `develop`
  * Naming: `feature/add-payment-gateway`
  * Workflow:
    ```bash
    # Membuat feature branch
    git checkout develop
    git pull origin develop
    git checkout -b feature/add-payment-gateway
    
    # Setelah selesai
    git checkout develop
    git pull origin develop
    git merge --no-ff feature/add-payment-gateway
    git push origin develop
    ```

- `hotfix/*` - Hotfix branches
  * Dibuat dari: `main`
  * Merge ke: `main` dan `develop`
  * Naming: `hotfix/fix-payment-calculation`
  * Untuk bug fixes di production
  * Workflow:
    ```bash
    # Membuat hotfix branch
    git checkout main
    git pull origin main
    git checkout -b hotfix/fix-payment-calculation
    
    # Setelah selesai
    git checkout main
    git pull origin main
    git merge --no-ff hotfix/fix-payment-calculation
    git tag -a v1.2.1 -m "Fix payment calculation"
    git push origin main --tags
    
    git checkout develop
    git pull origin develop
    git merge --no-ff hotfix/fix-payment-calculation
    git push origin develop
    ```

- `release/*` - Release branches
  * Dibuat dari: `develop`
  * Merge ke: `main` dan `develop`
  * Naming: `release/v1.2.0`
  * Untuk persiapan release
  * Workflow:
    ```bash
    # Membuat release branch
    git checkout develop
    git pull origin develop
    git checkout -b release/v1.2.0
    
    # Setelah testing selesai
    git checkout main
    git pull origin main
    git merge --no-ff release/v1.2.0
    git tag -a v1.2.0 -m "Version 1.2.0"
    git push origin main --tags
    
    git checkout develop
    git pull origin develop
    git merge --no-ff release/v1.2.0
    git push origin develop
    ```

### Branch Protection Rules

1. `main` branch:
   - Require pull request reviews
   - Require status checks to pass
   - Require linear history
   - Include administrators
   - Allow force pushes: NO
   - Allow deletions: NO

2. `develop` branch:
   - Require pull request reviews
   - Require status checks to pass
   - Allow force pushes: NO
   - Allow deletions: NO

### Commit Convention

Menggunakan Conventional Commits dengan format:
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Types:
- `feat`: Fitur baru
- `fix`: Bug fix
- `docs`: Perubahan dokumentasi
- `style`: Formatting, missing semicolons, etc
- `refactor`: Refactoring code
- `test`: Menambah/mengubah tests
- `chore`: Updating build tasks, package manager configs, etc

Contoh:
```
feat(payment): add new payment gateway integration

- Add Midtrans integration
- Add payment notification webhook
- Add payment status tracking

Closes #123
```

### Pull Request Template

Setiap pull request harus mengikuti template yang sudah disediakan di `.github/pull_request_template.md`:
- Description of changes
- Type of change (bugfix, feature, etc)
- How to test
- Checklist
- Screenshots (if applicable)
- Related issues

### Release Process

1. Create release branch
2. Update version in relevant files
3. Update CHANGELOG.md
4. Create pull request to main
5. After merge, tag version in main
6. Update develop with changes

## Monitoring & Observability

- Metrics: Prometheus (http://localhost:9090)
- Dashboards: Grafana (http://localhost:3000)
- Tracing: Jaeger (http://localhost:16686)
- Logging: ELK Stack

## Security

Implementasi keamanan mengikuti standar OWASP Top 10:
- Authentication & Authorization
- Rate Limiting
- Input Validation
- SQL Injection Prevention
- XSS Protection
- CORS Policy
- Security Headers 