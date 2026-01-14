# Mod Troubleshooter

A web-based tool for Skyrim SE mod users to visualize, analyze, and troubleshoot mod collections from Nexus Mods.

## Features

- **Collection Browser** - View all mods in a Nexus collection
- **FOMOD Visualizer** - Explore FOMOD installer structure and options
- **Load Order Analyzer** - Understand plugin dependencies and order
- **Conflict Detector** - Identify file conflicts between mods

## Requirements

- Node.js 18+
- Go 1.22+
- Nexus Mods Premium account (for mod downloads)
- Nexus API key ([get one here](https://www.nexusmods.com/users/myaccount?tab=api))

## Quick Start

### 1. Clone and setup

```bash
git clone https://github.com/your-repo/mod-troubleshooter.git
cd mod-troubleshooter

# Copy environment template
cp .env.example .env

# Edit .env and add your Nexus API key
```

### 2. Start the backend

```bash
cd backend
go run ./cmd/server
```

### 3. Start the frontend

```bash
cd frontend
npm install
npm run dev
```

### 4. Open in browser

Visit `http://localhost:5173`

## Development with Ralph

This project uses the [Ralph protocol](https://github.com/...) for AI-assisted development.

```bash
# Make loop executable
chmod +x loop.sh

# Run Ralph (max 20 iterations)
./loop.sh 20

# Monitor progress
watch -n 5 'cat IMPLEMENTATION_PLAN.md'
```

## Project Structure

```
mod-troubleshooter/
├── backend/                 # Go API server
│   ├── cmd/server/         # Entry point
│   └── internal/           # Internal packages
├── frontend/               # React application
│   └── src/
├── specs/                  # Feature specifications
├── .cursor/rules/          # Cursor AI rules
├── IMPLEMENTATION_PLAN.md  # Task tracking
├── AGENTS.md              # Operations guide
├── PROMPT.md              # Ralph instructions
└── loop.sh                # Ralph loop script
```

## Tech Stack

- **Frontend**: React 19, TypeScript, Vite, TanStack Query
- **Backend**: Go 1.22+, SQLite
- **APIs**: Nexus Mods GraphQL & REST

## License

MIT
