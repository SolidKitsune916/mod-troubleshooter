# Quick Start Guide

## Step 1: Install Prerequisites

**You need to install these tools first:**

1. **Node.js** (includes npm)
   - Download: https://nodejs.org/ (LTS version)
   - Install and restart your terminal

2. **Go**
   - Download: https://go.dev/dl/
   - Install and restart your terminal

## Step 2: Verify Installation

Open PowerShell and run:
```powershell
node --version
npm --version
go version
```

If all three commands show version numbers, you're ready!

## Step 3: Start the Servers

### Terminal 1 - Backend Server
```powershell
cd "C:\SolidKitsune Project\mod-troubleshooter\backend"
go run .\cmd\server\main.go
```
Wait for: `Server starting on http://localhost:8080`

### Terminal 2 - Frontend Server
```powershell
cd "C:\SolidKitsune Project\mod-troubleshooter\frontend"
npm install
npm run dev
```
Wait for: `Local: http://localhost:5173`

## Step 4: Open in Browser

Navigate to: **http://localhost:5173**

## Troubleshooting

- **"node is not recognized"** → Install Node.js and restart terminal
- **"go is not recognized"** → Install Go and restart terminal  
- **Port already in use** → Close other applications using ports 8080 or 5173
- **npm install fails** → Check internet connection and try again

For detailed help, see `SETUP_GUIDE.md` and `START_SERVERS.md`
