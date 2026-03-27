# 🚀 START HERE - Pixelcraft Backend

## Welcome to Pixelcraft Backend! 👋

Este é o backend de autenticação e gerenciamento de usuários do Pixelcraft.

---

## ⚡ Quick Start (3 Minutes)

### Step 1: Setup Database (1 min)
```powershell
# Run this in PowerShell
psql -U pedro -d postgres -c "CREATE DATABASE pixelcraft;"
psql -U pedro -d pixelcraft -f database/schema.sql
```

### Step 2: Install Dependencies (1 min)
```powershell
go mod download
```

### Step 3: Start Server (10 seconds)
```powershell
go run cmd/api/main.go
```

✅ **Server is now running at:** http://localhost:8080

---

## 🧪 Test It Now!

### Test 1: Health Check
```bash
curl http://localhost:8080/api/v1/health
```

**Expected:**
```json
{
  "status": "healthy",
  "service": "pixelcraft-api",
  "version": "1.0.0"
}
```

### Test 2: Register a User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@test.com",
    "password": "password123"
  }'
```

**You'll get back a JWT token!** 🎉

---

## 📚 What's Next?

### I Want To...

#### 🎯 Understand What Was Built
**Read:** [`EXECUTIVE_SUMMARY.md`](EXECUTIVE_SUMMARY.md)  
**Time:** 5 minutes  
**You'll learn:** What endpoints exist, security features, tech stack

#### 🔧 Integrate with Frontend
**Read:** [`API_REFERENCE.md`](API_REFERENCE.md)  
**Time:** 10 minutes  
**You'll learn:** All endpoints, request/response formats, authentication

#### 💻 Understand the Code
**Read:** [`PROJECT_STRUCTURE.md`](PROJECT_STRUCTURE.md)  
**Time:** 15 minutes  
**You'll learn:** How code is organized, data flow, architecture patterns

#### 🚀 Deploy to Production
**Read:** [`BACKEND_IMPLEMENTATION.md`](BACKEND_IMPLEMENTATION.md) → Production Checklist  
**Time:** 20 minutes  
**You'll learn:** Security hardening, environment setup, deployment steps

#### 🐛 Fix Issues
**Read:** [`QUICK_REFERENCE.md`](QUICK_REFERENCE.md) → Troubleshooting  
**Time:** 2 minutes  
**You'll learn:** Common errors and how to fix them

---

## 📖 All Documentation

| File | Purpose | Read Time |
|------|---------|-----------|
| [`INDEX.md`](INDEX.md) | Navigation guide | 2 min |
| [`EXECUTIVE_SUMMARY.md`](EXECUTIVE_SUMMARY.md) | Project overview | 5 min |
| [`QUICK_REFERENCE.md`](QUICK_REFERENCE.md) | Cheat sheet | 3 min |
| [`API_REFERENCE.md`](API_REFERENCE.md) | API documentation | 10 min |
| [`SETUP.md`](SETUP.md) | Setup instructions | 5 min |
| [`PROJECT_STRUCTURE.md`](PROJECT_STRUCTURE.md) | Code organization | 15 min |
| [`BACKEND_IMPLEMENTATION.md`](BACKEND_IMPLEMENTATION.md) | Technical deep-dive | 20 min |

**Total Documentation:** ~2,400 lines across 7 files

---

## 🎯 API Endpoints Overview

### Public (No Auth Required)
- `POST /api/v1/auth/register` - Create account
- `POST /api/v1/auth/login` - Login

### Protected (Auth Required)
- `GET /api/v1/users/me` - Get profile
- `PUT /api/v1/users/me` - Update profile

**Full documentation:** [`API_REFERENCE.md`](API_REFERENCE.md)

---

## 🏗️ Architecture at a Glance

```
┌──────────────┐
│   Frontend   │ React + Vite
└──────┬───────┘
       │ HTTP/JSON
       ▼
┌──────────────┐
│   Handlers   │ HTTP Layer (Gin)
└──────┬───────┘
       │
┌──────▼───────┐
│   Services   │ Business Logic
└──────┬───────┘
       │
┌──────▼───────┐
│ Repositories │ Data Access
└──────┬───────┘
       │
┌──────▼───────┐
│  PostgreSQL  │ Database
└──────────────┘
```

**Learn more:** [`PROJECT_STRUCTURE.md`](PROJECT_STRUCTURE.md)

---

## 🔐 Security Highlights

✅ **Passwords:** bcrypt hashing (never stored plain text)  
✅ **Authentication:** JWT tokens (72h expiration)  
✅ **CPF:** AES-256 encryption (never exposed)  
✅ **API:** CORS protection, input validation

**Full security details:** [`BACKEND_IMPLEMENTATION.md`](BACKEND_IMPLEMENTATION.md)

---

## 🛠️ Tech Stack

- **Language:** Go 1.21+
- **Framework:** Gin (HTTP router)
- **Database:** PostgreSQL 14+
- **Authentication:** JWT (golang-jwt/jwt)
- **Password Hashing:** bcrypt
- **Database Driver:** sqlx + lib/pq

---

## 💡 Common Commands

```bash
# Run server
go run cmd/api/main.go

# Build binary
go build -o pixelcraft-api cmd/api/main.go

# Run tests
go test ./...

# Format code
go fmt ./...
```

**More commands:** [`QUICK_REFERENCE.md`](QUICK_REFERENCE.md)

---

## 🆘 Need Help?

### Documentation Not Clear?
- Check [`INDEX.md`](INDEX.md) for navigation guide
- All docs are cross-referenced

### Something Not Working?
- [`QUICK_REFERENCE.md`](QUICK_REFERENCE.md) → Troubleshooting section
- Check server logs in terminal

### Want to Extend the Code?
- [`PROJECT_STRUCTURE.md`](PROJECT_STRUCTURE.md) → Layer responsibilities
- Code is organized by Clean Architecture

---

## ✅ Checklist

Before you start coding, make sure:

- [ ] PostgreSQL is running
- [ ] Database `pixelcraft` exists
- [ ] Schema is applied (`database/schema.sql`)
- [ ] Go dependencies installed (`go mod download`)
- [ ] `.env` file configured
- [ ] Server starts successfully
- [ ] Health endpoint responds

**Detailed setup:** [`SETUP.md`](SETUP.md)

---

## 🎓 Learning Path

### Beginner (Total: 15 min)
1. Run Quick Start above (3 min)
2. Read [`EXECUTIVE_SUMMARY.md`](EXECUTIVE_SUMMARY.md) (5 min)
3. Read [`QUICK_REFERENCE.md`](QUICK_REFERENCE.md) (3 min)
4. Test endpoints with cURL (4 min)

### Developer (Total: 45 min)
1. Run Quick Start (3 min)
2. Read [`PROJECT_STRUCTURE.md`](PROJECT_STRUCTURE.md) (15 min)
3. Read [`BACKEND_IMPLEMENTATION.md`](BACKEND_IMPLEMENTATION.md) (20 min)
4. Explore code files (7 min)

### Frontend Integration (Total: 20 min)
1. Run Quick Start (3 min)
2. Read [`API_REFERENCE.md`](API_REFERENCE.md) (10 min)
3. Test all endpoints (7 min)

---

## 🚀 What You Can Do Now

✅ **Register users** with username, email, password  
✅ **Login users** with username OR email  
✅ **Get user profile** with JWT authentication  
✅ **Update profile** with optional fields (name, discord, whatsapp)  
✅ **Secure passwords** with bcrypt  
✅ **Encrypt CPF** for billing (when needed)  
✅ **CORS** configured for frontend

**Full feature list:** [`EXECUTIVE_SUMMARY.md`](EXECUTIVE_SUMMARY.md)

---

## 📊 Project Stats

- **Lines of Code:** ~1,200
- **Documentation:** ~2,400 lines
- **Files:** 10 Go source files
- **Endpoints:** 4 (2 public, 2 protected)
- **Time to Setup:** 3 minutes
- **Time to Understand:** 15-45 minutes

---

## 🎯 Next Steps

1. ✅ **You are here:** Getting started
2. 📖 **Read docs:** Choose from list above
3. 🧪 **Test API:** Use cURL examples
4. 🔌 **Integrate:** Connect frontend
5. 🚀 **Deploy:** Follow production checklist

---

**Questions?** Check [`INDEX.md`](INDEX.md) for the right documentation file.

**Version:** 1.0.0  
**Last Updated:** 2025-01-12

---

## 🎉 You're Ready!

The backend is **100% functional** and ready to use. Pick a documentation file from above and start exploring!

**Recommended first read:** [`EXECUTIVE_SUMMARY.md`](EXECUTIVE_SUMMARY.md)
