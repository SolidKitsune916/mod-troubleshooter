# Ralph Instructions - Mod Troubleshooter

## Phase 0: Orient

1. Study `AGENTS.md` for project-specific build/test commands
2. Study `IMPLEMENTATION_PLAN.md` for current tasks and priorities
3. Study `specs/` directory for detailed feature requirements

## Phase 1: Load Relevant Rules

Before implementing, load the appropriate rules based on what you'll modify:

| If touching... | Study this rule file |
|----------------|---------------------|
| Any `backend/**/*.go` | `.cursor/rules/2000-golang-backend.mdc` |
| Any `frontend/src/**/*.{ts,tsx}` | `.cursor/rules/1000-react-general.mdc` |
| `frontend/src/components/**/*.tsx` | `.cursor/rules/1001-react-components.mdc` |
| `frontend/src/hooks/**/*.ts` | `.cursor/rules/1002-react-hooks.mdc` |
| `*Form*.tsx` | `.cursor/rules/1003-react-forms.mdc` |
| Any UI code (`*.tsx`) | `.cursor/rules/1004-accessibility-wcag.mdc` |
| Any UI code (`*.tsx`) | `.cursor/rules/1005-qol-ux.mdc` |
| `*.test.{ts,tsx}` | `.cursor/rules/1006-testing.mdc` |
| `*.tsx` or `*.css` | `.cursor/rules/1007-tailwindcss.mdc` |
| `frontend/src/services/**/*.ts` | `.cursor/rules/1009-services.mdc` |
| `frontend/src/store/**/*.ts` | `.cursor/rules/1010-state-management.mdc` |

**Load multiple rules when applicable.**

For detailed feature requirements, consult:
- `specs/00-project-overview.md` - Overall architecture
- `specs/01-go-backend.md` - Backend API details
- `specs/02-fomod-visualizer.md` - FOMOD feature
- `specs/03-load-order-analyzer.md` - Load order feature
- `specs/04-conflict-detector.md` - Conflict detection feature

## Phase 2: Execute

1. Pick the highest-priority incomplete task from `IMPLEMENTATION_PLAN.md`
2. Search the codebase first—**do not assume functionality is missing**
3. Implement that ONE task completely following the loaded rules
4. No placeholders, no stubs, no "TODO" comments—implement fully
5. Run validation:
   - Backend: `cd backend && go test -v ./... && go vet ./...`
   - Frontend: `cd frontend && npm run typecheck && npm test`
6. If tests pass: `git add -A && git commit -m "feat: [US-XXX] [task summary]"`
7. Update `IMPLEMENTATION_PLAN.md`: mark complete, note discoveries
8. If you learned operational patterns, update `AGENTS.md` briefly

## Critical Rules

- **One task per iteration, then exit**
- **Tests must pass before committing**
- **Follow the loaded rule files exactly**—they define project standards
- **Follow the spec files**—they define feature requirements
- If you find bugs unrelated to your task, document in `IMPLEMENTATION_PLAN.md`
- Keep `IMPLEMENTATION_PLAN.md` current—future iterations depend on it
- Keep `AGENTS.md` operational only—no progress notes, no changelogs

## Tech Stack Reminders

### Backend (Go)
- Go 1.22+ with standard library HTTP server
- Use `r.PathValue("param")` for path parameters (Go 1.22+)
- Always handle errors explicitly
- Use context for cancellation
- Configure CORS for `http://localhost:5173`

### Frontend (React)
- React 19 + TypeScript strict mode
- TanStack Query for data fetching
- Gaming dark theme (see existing viewer-app styles)
- Accessibility: WCAG 2.2 AA compliance
- 44px minimum touch targets

### API Communication
- Frontend calls backend at `http://localhost:8080/api/...`
- All responses use `{ data, error, message }` envelope
- Content-Type: application/json
