# Night Owls Go - Community Watch Shift Scheduler

A comprehensive full-stack application for managing volunteer shifts in community watch programs. This system provides both a **Go backend API** and a **SvelteKit frontend** for complete shift scheduling, user management, incident reporting, and administrative oversight.

## 🚀 Key Features

### Core Functionality
- **📱 Passwordless Authentication** - OTP-based login via SMS (Twilio integration)
- **⏰ Dynamic Shift Scheduling** - Cron-based recurring shifts with seasonal validity
- **👥 Buddy System** - Volunteer pairing for enhanced safety
- **📊 Incident Reporting** - GPS-enabled reports with severity levels and archiving
- **📢 Broadcast System** - Targeted messaging to user groups
- **🏥 Emergency Contacts** - Centralized emergency contact management
- **📈 Admin Dashboard** - Comprehensive analytics and management tools

### Technical Features
- **💾 Ephemeral Shifts** - Shift instances created only upon booking
- **📤 Transactional Outbox** - Reliable message delivery with retry logic
- **🔄 Background Jobs** - Automated report archiving, broadcast processing
- **🌐 PWA Support** - Progressive Web App with push notifications
- **🔧 Development Mode** - Enhanced debugging and testing features

## 🛠 Tech Stack

### Backend (Go 1.24.2+)
- **[Fuego](https://github.com/go-fuego/fuego)** - Modern HTTP framework with OpenAPI integration
- **SQLite** - Embedded database with full SQL capabilities
- **[sqlc](https://sqlc.dev/)** - Type-safe SQL query generation
- **[golang-migrate](https://github.com/golang-migrate/migrate)** - Database schema migrations
- **[slog](https://pkg.go.dev/log/slog)** - Structured logging (Go 1.21+)
- **[robfig/cron](https://github.com/robfig/cron)** - Background job scheduling
- **[golang-jwt](https://github.com/golang-jwt/jwt)** - JWT authentication
- **[Swaggo](https://github.com/swaggo/swag)** - Auto-generated OpenAPI documentation

### Frontend (SvelteKit)
- **[SvelteKit](https://kit.svelte.dev/)** - Full-stack web framework
- **[TailwindCSS](https://tailwindcss.com/)** - Utility-first CSS framework
- **[TanStack Query](https://tanstack.com/query)** - Data fetching and caching
- **[Zod](https://zod.dev/)** - Runtime type validation
- **[Workbox](https://developers.google.com/web/tools/workbox)** - Service worker and PWA features

### Prerequisites

1. **Go 1.24.2+** - Download from [golang.org/dl/](https://golang.org/dl/)
2. **Node.js 20+** with **pnpm** - For frontend development
3. **Development Tools:**
   ```bash
   # Database migrations
   go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   
   # SQL code generation
   go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
   
   # OpenAPI documentation
   go install github.com/swaggo/swag/cmd/swag@latest
   ```

### Project Setup

1. **Clone and install dependencies:**
   ```bash
   git clone <repository-url>
   cd night-owls-go
   go mod tidy
   cd app && pnpm install
   ```

2. **Environment Configuration:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

   **Essential Environment Variables:**
   ```env
   # Server Configuration
   SERVER_PORT=5888
   DATABASE_PATH=./community_watch.db
   STATIC_DIR=./app/build
   
   # Security
   JWT_SECRET=your-super-secure-jwt-secret-change-this
   JWT_EXPIRATION_HOURS=24
   
   # OTP Configuration
   OTP_VALIDITY_MINUTES=5
   OTP_LOG_PATH=./sms_outbox.log
   
   # Development
   DEV_MODE=true
   LOG_LEVEL=debug
   LOG_FORMAT=json
   
   # Twilio (Optional - for SMS OTP)
   TWILIO_ACCOUNT_SID=your-twilio-sid
   TWILIO_AUTH_TOKEN=your-twilio-token
   TWILIO_VERIFY_SID=your-verify-sid
   
   # PWA/Push Notifications (Optional)
   VAPID_PUBLIC=your-vapid-public-key
   VAPID_PRIVATE=your-vapid-private-key
   VAPID_SUBJECT=mailto:admin@example.com
   ```

3. **Database Setup:**
   ```bash
   # Migrations run automatically on startup, or manually:
   migrate -database "sqlite3://./community_watch.db" -path internal/db/migrations up
   ```

## 🏃‍♂️ Running the Application

### Development Mode

1. **Start the backend:**
   ```bash
   go run ./cmd/server/main.go
   ```

2. **Start the frontend (in another terminal):**
   ```bash
   cd app
   pnpm run dev
   ```

3. **Access the application:**
   - **Frontend:** http://localhost:5173
   - **Backend API:** http://localhost:5888
   - **API Documentation:** http://localhost:5888/swagger
   - **Health Check:** http://localhost:5888/health

### Production Mode

```bash
# Build frontend
cd app && pnpm run build && cd ..

# Start backend (serves both API and built frontend)
go run ./cmd/server/main.go
```

Access at: http://localhost:5888

## 📋 API Documentation

The API is fully documented with **OpenAPI/Swagger** and available at `/swagger` when running.

### Core Endpoints

#### Authentication
- `POST /api/auth/register` - Register/login with phone number
- `POST /api/auth/verify` - Verify OTP and get JWT token
- `POST /api/auth/dev-login` - Development-only direct login

#### Public Endpoints
- `GET /schedules` - List available shift schedules
- `GET /shifts/available` - List available shift slots
- `GET /api/emergency-contacts` - Get emergency contacts
- `GET /health` - Health check endpoint

#### Protected User Endpoints (Require JWT)
- `POST /bookings` - Book a shift
- `GET /bookings/my` - Get user's bookings
- `POST /bookings/{id}/checkin` - Check in to a shift
- `DELETE /bookings/{id}` - Cancel a booking
- `POST /bookings/{id}/report` - Submit incident report
- `POST /reports/off-shift` - Submit off-shift report
- `GET /api/broadcasts` - Get user broadcasts
- `POST /push/subscribe` - Subscribe to push notifications

#### Admin Endpoints (Require JWT + Admin Role)
- **Users:** `GET|POST|PUT|DELETE /api/admin/users/**`
- **Schedules:** `GET|POST|PUT|DELETE /api/admin/schedules/**`
- **Reports:** `GET|PUT /api/admin/reports/**` (list, archive, unarchive)
- **Broadcasts:** `GET|POST /api/admin/broadcasts/**`
- **Dashboard:** `GET /api/admin/dashboard`
- **Emergency Contacts:** `GET|POST|PUT|DELETE /api/admin/emergency-contacts/**`

### Authentication Flow

1. **Register/Login:** `POST /api/auth/register` with phone number
2. **Receive OTP:** Via SMS (or check logs in development)
3. **Verify:** `POST /api/auth/verify` with OTP code
4. **Use JWT:** Include `Authorization: Bearer <token>` in subsequent requests

## 🏗 Architecture

### Clean Architecture Layers

```
┌─────────────────────┐
│   Frontend (SvelteKit)   │
├─────────────────────┤
│   API Layer         │  ← HTTP handlers, middleware, routing
│   (internal/api/)   │
├─────────────────────┤
│   Service Layer     │  ← Business logic, orchestration
│   (internal/service/)│
├─────────────────────┤
│   Data Layer        │  ← Database queries, models
│   (internal/db/)    │
└─────────────────────┘
```

### Key Services

- **UserService** - Authentication, user management
- **ScheduleService** - Shift scheduling, cron parsing
- **BookingService** - Shift bookings, check-ins
- **ReportService** - Incident reporting, GPS tracking
- **BroadcastService** - Message broadcasting to user groups
- **AdminDashboardService** - Analytics and reporting
- **EmergencyContactService** - Emergency contact management
- **ReportArchivingService** - Automated report lifecycle

### Background Jobs (Cron)

- **Every 1 minute:** Process outbox messages (SMS/push notifications)
- **Every 30 seconds:** Process pending broadcasts
- **Daily at 2 AM:** Auto-archive old reports (configurable retention)

## 🧪 Testing

```bash
# Run Go tests
go test ./...

# Run Go tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# Run frontend tests
cd app
pnpm run test:unit

# Run end-to-end tests
pnpm run test:e2e
```

## 🔧 Development Tools

### Code Generation

```bash
# Regenerate SQL queries (after schema changes)
sqlc generate

# Update OpenAPI documentation
swag init -g cmd/server/main.go -o ./docs/swagger

# Frontend type checking
cd app && pnpm run check
```

### Database Management

```bash
# Create new migration
migrate create -ext sql -dir internal/db/migrations -seq your_migration_name

# Apply migrations
migrate -database "sqlite3://./community_watch.db" -path internal/db/migrations up

# Rollback last migration
migrate -database "sqlite3://./community_watch.db" -path internal/db/migrations down 1
```

## 📊 Features in Detail

### Shift Management
- **Cron-based scheduling** with timezone support
- **Seasonal validity** (start/end dates)
- **Ephemeral shift instances** (created on booking)
- **Buddy system** for volunteer safety
- **Attendance tracking** with GPS check-in

### Reporting System
- **Three severity levels:** Normal, Suspicion, Incident
- **GPS coordinates** and accuracy tracking
- **Automatic archiving** based on age and severity
- **Off-shift reporting** capability
- **Admin report management** (archive/unarchive)

### User Management
- **Phone-based authentication** (international support)
- **Role-based access** (admin/owl)
- **OTP verification** via Twilio or mock
- **User profiles** with contact information

### Broadcasting
- **Targeted messaging** by user role/status
- **Scheduled broadcasts** with automatic processing
- **Delivery tracking** and retry logic
- **Audience selection** (all/admins/owls/active users)

## 🚨 Troubleshooting

### Common Issues

1. **Migration Errors:** Ensure database path is writable and migrations folder exists
2. **Build Failures:** Check Go version (1.24.2+) and run `go mod tidy`
3. **Frontend Issues:** Ensure Node.js 20+ and run `pnpm install`
4. **JWT Errors:** Verify `JWT_SECRET` is set and consistent
5. **OTP Issues:** Check Twilio credentials or enable `DEV_MODE`

### Development Tips

- Use `DEV_MODE=true` for enhanced logging and mock OTP
- Check `/health` endpoint for system status
- Monitor `./sms_outbox.log` for OTP codes in development
- Use `/swagger` for API testing and documentation

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes and test thoroughly
4. Commit your changes: `git commit -m 'Add amazing feature'`
5. Push to the branch: `git push origin feature/amazing-feature`
6. Open a Pull Request

### Code Standards

- Follow Go best practices and formatting (`go fmt`)
- Write tests for new functionality
- Update documentation for API changes
- Use conventional commit messages

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Built with ❤️ for community safety and volunteer coordination** 