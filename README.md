# Shortleak - Docker Development Setup

Panduan lengkap untuk menjalankan aplikasi Shortleak menggunakan Docker dengan support untuk multiple environment (development, test, production).

## 📋 Prerequisites

- Docker & Docker Compose
- Task (optional, untuk task runner)
- Make (optional, untuk makefile)

## 🚀 Quick Start

### 1. Setup Environment Files

```bash
# Copy dan edit environment files
cp .env.example .env
cp shortleak-be/.env.example shortleak-be/.env
cp shortleak-fe/.env.example shortleak-fe/.env
```

### 2. Jalankan Development Environment

```bash
# Menggunakan docker-compose langsung
docker-compose up --build
```

## 🛠 Available Environments

### Development Environment
- **Database**: PostgreSQL di port 5432
- **Backend**: Go server di port 8090  
- **Frontend**: React app di port 5173
- **Hot Reload**: Enabled untuk development

```bash
# Start development
docker-compose up --build
```

### Test Environment (ke folder shortleak-be)
- **Backend**: Go dengan test configuration
- **Auto Migration**: Ya
- **Coverage Report**: Ya

```bash
# Run tests (ke folder shortleak-be)
go test ./... -coverprofile=coverage && go tool cover -html=coverage
```

## 📊 Database Management

### Migrations (ke folder shortleak-be)

```bash
# Development (Ke folder shortleak-be)
go run ./cmd/migrate/main.go refresh"
```

### Seeds (ke folder shortleak-be)

```bash
# Development (ke folder shortleak-be)
go run ./cmd/seed/main.go
```

## 📁 Project Structure

```
project-root/
├── docker-compose.yml          # Main docker configuration
├── .env                        # Root environment variables
├── init-db.sql                 # Database initialization
├── backend/
│   ├── Dockerfile              # Multi-stage Dockerfile
│   ├── .dockerignore           # Docker ignore file
│   ├── .env                    # Backend environment
│   ├── main.go
│   ├── go.mod
│   ├── go.sum
│   ├── Taskfile.yml           # Original taskfile
│   └── cmd/
│       ├── migrate/
│       └── seed/
├── frontend/
│   ├── Dockerfile              # Frontend Dockerfile
│   ├── .dockerignore           # Frontend ignore file
│   ├── .env                    # Frontend environment
│   ├── nginx.conf              # Nginx configuration
│   └── src/
```
## Login SHORTLEAK

```bash
# User
sysadmin@shortleak.com
abcDEF123!
```

## 🌐 Service URLs

| Service | Development | Test | Production |
|---------|------------|------|------------|
| Frontend | http://localhost:5173 | - | http://localhost:5173 |
| Backend | http://localhost:8090 | - | http://localhost:8090 |
| PostgreSQL | localhost:5432
| pgAdmin | http://localhost:8082 | - | - |

## 🐳 Docker Commands Reference

### Basic Operations

```bash
# Build services
docker-compose build

# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f

# Clean up
docker-compose down -v --rmi all --remove-orphans
```

### Profiles

```bash
# Test profile
docker-compose --profile test up

# Production profile
docker-compose --profile production up

# Tools profile (pgAdmin)
docker-compose --profile tools up
```

## 📝 Environment Variables

### Root .env
```env
PLATFORM=shortleak
JWT_SECRET=shortleak-jwt-secret
DB_DATABASE_DEVELOPMENT=shortleak-dev
DB_USERNAME_DEVELOPMENT=postgres
DB_PASSWORD_DEVELOPMENT=12345
# ... dst
```

### Backend .env
```env
NODE_ENV=development
PLATFORM=shortleak
JWT_SECRET=shortleak-jwt-secret
DB_DATABASE_DEVELOPMENT=shortleak-dev
# ... dst
```

## 🚀 CI/CD Pipeline

Pipeline GitHub Actions tersedia untuk:
- ✅ Unit testing
- ✅ Integration testing  
- ✅ Coverage reporting
- ✅ Docker testing

## 🔍 Troubleshooting

### Common Issues

1. **Database connection failed**
   ```bash
   # Check database status
   docker-compose ps
   make health
   ```

2. **Port already in use**
   ```bash
   # Change ports in .env file
   BACKEND_PORT=8081
   FRONTEND_PORT=3001
   ```

3. **Permission denied**
   ```bash
   # Fix file permissions
   sudo chown -R $USER:$USER .
   ```

4. **Out of disk space**
   ```bash
   # Clean up Docker
   make clean
   docker system prune -a --volumes
   ```

### Logs

```bash
# View all logs
make logs

# View specific service logs
docker-compose logs backend-dev
docker-compose logs postgres-dev
```

## 📚 Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Go Documentation](https://golang.org/doc/)
- [React Documentation](https://reactjs.org/docs/)

---

🎉 **Happy Coding!** Jika ada pertanyaan atau masalah, silakan buat issue di repository ini.