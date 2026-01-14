# Ralph Instructions

## Phase 0: Orient

1. Study `AGENTS.md` for project-specific build/test commands
2. Study `IMPLEMENTATION_PLAN.md` for current tasks and priorities
3. Study `specs/*` if present for detailed requirements

## Phase 1: Load Relevant Rules

Before implementing, load the appropriate rules based on what you'll modify:

| If touching... | Study this rule file |
|----------------|---------------------|
| Any `src/**/*.{ts,tsx}` | `.cursor/rules/1000-react-general.mdc` |
| `src/components/**/*.tsx` | `.cursor/rules/1001-react-components.mdc` |
| `src/hooks/**/*.ts` or `*use*.ts` | `.cursor/rules/1002-react-hooks.mdc` |
| `*Form*.tsx` or `src/components/forms/*` | `.cursor/rules/1003-react-forms.mdc` |
| Any UI code (`*.tsx`) | `.cursor/rules/1004-accessibility-wcag.mdc` |
| Any UI code (`*.tsx`) | `.cursor/rules/1005-qol-ux.mdc` |
| `*.test.{ts,tsx}` | `.cursor/rules/1006-testing.mdc` |
| `*.tsx` or `*.css` | `.cursor/rules/1007-tailwindcss.mdc` |
| Any `*.ts` or `*.tsx` | `.cursor/rules/1008-typescript.mdc` |
| `src/services/**/*.ts` or `src/api/**/*.ts` | `.cursor/rules/1009-services.mdc` |
| `src/store/**/*.ts` or `*Context*.tsx` | `.cursor/rules/1010-state-management.mdc` |
| Any `**/*.go`, `go.mod`, `go.sum` | `.cursor/rules/2000-golang-backend.mdc` |

**Load multiple rules when applicable.** For example, creating a form component requires: 1001 (components) + 1003 (forms) + 1004 (accessibility) + 1007 (tailwind).

For deep reference on complex implementations, consult:
- `docs/React-TypeScript-Best-Practices-Updated.md`
- `docs/WCAG-2_2-Guide-Updated.md`
- `docs/QoL-UX-Best-Practices-Updated.md`

## Phase 2: Execute

1. Pick the highest-priority incomplete task from `IMPLEMENTATION_PLAN.md`
2. Search the codebase first—**do not assume functionality is missing**
3. Implement that ONE task completely following the loaded rules
4. No placeholders, no stubs, no "TODO" comments—implement fully
5. Run validation (see `AGENTS.md` for commands)
6. If tests pass: `git add -A && git commit -m "feat: [task summary]"`
7. Update `IMPLEMENTATION_PLAN.md`: mark complete, note discoveries
8. If you learned operational patterns, update `AGENTS.md` briefly

## Critical Rules

- **One task per iteration, then exit**
- **Tests must pass before committing**
- **Follow the loaded rule files exactly**—they define project standards
- If you find bugs unrelated to your task, document in `IMPLEMENTATION_PLAN.md`
- Keep `IMPLEMENTATION_PLAN.md` current—future iterations depend on it
- Keep `AGENTS.md` operational only—no progress notes, no changelogs
