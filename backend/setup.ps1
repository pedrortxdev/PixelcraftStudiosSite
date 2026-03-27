# Pixelcraft Backend Setup Script
# Run with: .\setup.ps1

Write-Host "🎮 Pixelcraft Backend Setup" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

# Check if Go is installed
Write-Host "📦 Checking Go installation..." -ForegroundColor Yellow
try {
    $goVersion = go version
    Write-Host "✅ $goVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ Go is not installed. Please install Go 1.21+ from https://go.dev/dl/" -ForegroundColor Red
    exit 1
}

# Check if PostgreSQL is installed
Write-Host "📦 Checking PostgreSQL..." -ForegroundColor Yellow
try {
    $psqlVersion = psql --version
    Write-Host "✅ $psqlVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ PostgreSQL is not installed or not in PATH" -ForegroundColor Red
    Write-Host "   Please install PostgreSQL from https://www.postgresql.org/download/" -ForegroundColor Red
    exit 1
}

# Install Go dependencies
Write-Host ""
Write-Host "📥 Installing Go dependencies..." -ForegroundColor Yellow
go mod download
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Dependencies installed" -ForegroundColor Green
} else {
    Write-Host "❌ Failed to install dependencies" -ForegroundColor Red
    exit 1
}

go mod tidy

# Setup database
Write-Host ""
Write-Host "🗄️  Setting up database..." -ForegroundColor Yellow
Write-Host "   This will create the 'pixelcraft' database and tables" -ForegroundColor Gray

$setupDb = Read-Host "Do you want to setup the database now? (y/n)"
if ($setupDb -eq "y" -or $setupDb -eq "Y") {
    Write-Host "   Connecting to PostgreSQL..." -ForegroundColor Gray
    
    # Create database
    $createDbSql = "CREATE DATABASE pixelcraft;"
    psql -U pedro -d postgres -c $createDbSql 2>$null
    
    # Run schema
    psql -U pedro -d pixelcraft -f database/schema.sql
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Database setup complete" -ForegroundColor Green
    } else {
        Write-Host "⚠️  Database may already exist or there was an error" -ForegroundColor Yellow
    }
} else {
    Write-Host "⏭️  Skipping database setup" -ForegroundColor Gray
}

# Build the application
Write-Host ""
Write-Host "🔨 Building application..." -ForegroundColor Yellow
go build -o pixelcraft-api.exe cmd/api/main.go
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Build successful: pixelcraft-api.exe" -ForegroundColor Green
} else {
    Write-Host "❌ Build failed" -ForegroundColor Red
    exit 1
}

# Final instructions
Write-Host ""
Write-Host "================================" -ForegroundColor Cyan
Write-Host "🎉 Setup Complete!" -ForegroundColor Green
Write-Host ""
Write-Host "To start the API server, run:" -ForegroundColor White
Write-Host "  go run cmd/api/main.go" -ForegroundColor Cyan
Write-Host "  or" -ForegroundColor Gray
Write-Host "  .\pixelcraft-api.exe" -ForegroundColor Cyan
Write-Host ""
Write-Host "API will be available at: http://localhost:8080" -ForegroundColor White
Write-Host "Health check: http://localhost:8080/api/v1/health" -ForegroundColor White
Write-Host ""
Write-Host "📚 See README.md for API documentation" -ForegroundColor Gray
Write-Host "================================" -ForegroundColor Cyan
