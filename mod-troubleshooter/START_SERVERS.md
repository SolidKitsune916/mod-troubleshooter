# Starting the Application Locally

## Prerequisites

1. **Node.js and npm** - Install from https://nodejs.org/
2. **Go** - Install from https://go.dev/dl/

## Starting the Backend Server

Open a terminal and run:

```bash
cd mod-troubleshooter/backend
go run ./cmd/server
```

The backend will start on **http://localhost:8080**

**Note:** The backend can run without a Nexus API key for basic testing, but collection features will require an API key configured in Settings.

## Starting the Frontend Server

Open a **separate** terminal and run:

```bash
cd mod-troubleshooter/frontend
npm install  # First time only
npm run dev
```

The frontend will start on **http://localhost:5173**

## Accessing the Application

Once both servers are running:
- Open your browser to **http://localhost:5173**
- The frontend will proxy API requests to the backend automatically

## Quick Start Scripts

### Windows (PowerShell)

**Terminal 1 - Backend:**
```powershell
cd mod-troubleshooter\backend
go run .\cmd\server\main.go
```

**Terminal 2 - Frontend:**
```powershell
cd mod-troubleshooter\frontend
npm install
npm run dev
```

### Linux/Mac

**Terminal 1 - Backend:**
```bash
cd mod-troubleshooter/backend
go run ./cmd/server
```

**Terminal 2 - Frontend:**
```bash
cd mod-troubleshooter/frontend
npm install
npm run dev
```

## Troubleshooting

### Backend won't start
- Ensure Go is installed: `go version`
- Check if port 8080 is already in use
- Verify dependencies: `go mod download` in the backend directory

### Frontend won't start
- Ensure Node.js is installed: `node --version` and `npm --version`
- Install dependencies: `npm install` in the frontend directory
- Check if port 5173 is already in use

### API requests failing
- Ensure backend is running on port 8080
- Check browser console for CORS errors
- Verify backend logs for errors

## Environment Variables (Optional)

Create a `.env` file in the `backend` directory:

```env
PORT=8080
NEXUS_API_KEY=your_api_key_here
DATA_DIR=./data
CACHE_TTL_HOURS=168
ENVIRONMENT=development
CORS_ORIGINS=http://localhost:5173,http://localhost:3000
```
