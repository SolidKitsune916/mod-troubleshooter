# Setup Guide - Installing Prerequisites

## Required Tools

You need to install two tools before running the application:

### 1. Node.js (for Frontend)

**Download and Install:**
- Visit: https://nodejs.org/
- Download the **LTS version** (recommended)
- Run the installer and follow the prompts
- **Important**: Check "Add to PATH" during installation

**Verify Installation:**
```powershell
node --version
npm --version
```

### 2. Go (for Backend)

**Download and Install:**
- Visit: https://go.dev/dl/
- Download the Windows installer (`.msi` file)
- Run the installer and follow the prompts
- **Important**: The installer should add Go to PATH automatically

**Verify Installation:**
```powershell
go version
```

## After Installation

1. **Close and reopen your terminal** (or restart PowerShell) so PATH changes take effect

2. **Verify both tools are available:**
```powershell
node --version
npm --version
go version
```

3. **Then follow the instructions in `START_SERVERS.md`**

## Quick Install Commands (Alternative)

### Using Chocolatey (if you have it installed):

```powershell
# Install Node.js
choco install nodejs-lts -y

# Install Go
choco install golang -y
```

### Using Winget (Windows Package Manager):

```powershell
# Install Node.js
winget install OpenJS.NodeJS.LTS

# Install Go
winget install GoLang.Go
```

## Troubleshooting

### "npm is not recognized"
- Node.js may not be installed
- Node.js may not be in PATH
- **Solution**: Reinstall Node.js and ensure "Add to PATH" is checked

### "go is not recognized"
- Go may not be installed
- Go may not be in PATH
- **Solution**: Reinstall Go or manually add `C:\Program Files\Go\bin` to your PATH

### Adding to PATH Manually:
1. Open System Properties → Environment Variables
2. Edit "Path" under User variables
3. Add the installation directories:
   - `C:\Program Files\nodejs` (for Node.js)
   - `C:\Program Files\Go\bin` (for Go)
4. Restart your terminal
