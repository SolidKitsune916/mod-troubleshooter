# Project Instructions

## Quick Reference

```bash
# Frontend (React/Vite)
npm run dev          # Development server
npm run build        # Production build
npm run typecheck    # Type checking
npm test             # Run tests
npm run lint         # Lint code

# Backend (Go)
make dev             # Development with hot reload
make build           # Build binary
go test -v ./...     # Run tests
go vet ./...         # Vet code
golangci-lint run    # Lint
```

## Project Structure

```
src/                     # React frontend
cmd/                     # Go backend entry points
internal/                # Go private packages
.claude/                 # Claude rules (use these)
.cursor/rules/           # Cursor rules (ignore)
```

## Before Implementing

**Load relevant rules from `.claude/` based on what you're modifying:**

| Task | Load These Rules |
|------|------------------|
| React components | `1000-react-general.md`, `1001-react-components.md` |
| Custom hooks | `1000-react-general.md`, `1002-react-hooks.md` |
| Forms | `1001-react-components.md`, `1003-react-forms.md` |
| Any UI code | `1004-accessibility-wcag.md`, `1005-qol-ux.md` |
| Tests | `1006-testing.md` |
| Styling | `1007-tailwindcss.md` |
| TypeScript | `1008-typescript.md` |
| API services | `1009-services.md` |
| State/Context | `1010-state-management.md` |
| Go backend | `2000-golang-backend.md` |

**Load multiple rules when tasks overlap.** Example: Creating a form component → load `1001` + `1003` + `1004` + `1007`.

## Core Standards

- **TypeScript**: Strict mode, no `any`, explicit prop interfaces
- **React**: Functional components, hooks only, composition over prop drilling
- **Go**: Explicit error handling, context for cancellation, proper CORS
- **Accessibility**: WCAG 2.2 AA compliance, semantic HTML, keyboard navigation
- **Testing**: Test behavior not implementation, use Testing Library

## Workflow

1. Load relevant `.claude/*.md` rules
2. Search codebase first—don't assume functionality is missing
3. Implement fully—no placeholders, no TODOs
4. Run validation commands before committing
5. Commit with descriptive message: `feat:`, `fix:`, `refactor:`
