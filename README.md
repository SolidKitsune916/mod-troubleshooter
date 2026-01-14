# Ralph + Cursor Rules Setup

## Directory Structure

Your project should look like this:

```
your-project/
├── .cursor/
│   └── rules/                    # Your rule files go here
│       ├── 1000-react-general.mdc
│       ├── 1001-react-components.mdc
│       ├── 1002-react-hooks.mdc
│       ├── 1003-react-forms.mdc
│       ├── 1004-accessibility-wcag.mdc
│       ├── 1005-qol-ux.mdc
│       ├── 1006-testing.mdc
│       ├── 1007-tailwindcss.mdc
│       ├── 1008-typescript.mdc
│       ├── 1009-services.mdc
│       └── 1010-state-management.mdc
├── docs/                         # Full reference guides
│   ├── React-TypeScript-Best-Practices-Updated.md
│   ├── WCAG-2_2-Guide-Updated.md
│   └── QoL-UX-Best-Practices-Updated.md
├── specs/                        # Feature specifications
│   └── (your specs here)
├── src/                          # Your application code
├── loop.sh                       # Ralph loop script
├── PROMPT.md                     # Ralph instructions
├── IMPLEMENTATION_PLAN.md        # Task list (Ralph manages this)
├── AGENTS.md                     # Operational knowledge
└── ralph.log                     # Loop output log (auto-created)
```

## Setup Steps

### 1. Copy Ralph Files

Copy these files to your project root:
- `loop.sh`
- `PROMPT.md`
- `IMPLEMENTATION_PLAN.md`
- `AGENTS.md`

### 2. Organize Cursor Rules

Move your `.mdc` files to `.cursor/rules/`:

```bash
mkdir -p .cursor/rules
mv *.mdc .cursor/rules/
```

### 3. Move Reference Docs

```bash
mkdir -p docs
mv *-Updated.md docs/
```

### 4. Update AGENTS.md

Edit `AGENTS.md` with your actual project commands:
- Your test command (`npm test`, `vitest`, `pytest`, etc.)
- Your typecheck command
- Your lint command
- Your dev server command

### 5. Create Initial Specs

Create spec files in `specs/` for what you want to build:

```bash
mkdir -p specs
```

Example spec (`specs/user-auth.md`):
```markdown
# User Authentication

## Overview
Implement user login and registration.

## Acceptance Criteria
- [ ] Login form with email/password
- [ ] Email validation (format check)
- [ ] Password requirements (8+ chars, 1 number)
- [ ] Error messages for invalid inputs
- [ ] Loading state during submission
- [ ] Redirect to dashboard on success
```

### 6. Populate Implementation Plan

Either manually or ask Claude to plan:

```bash
# Manual: Edit IMPLEMENTATION_PLAN.md directly

# Or ask Claude to create the plan:
echo "Study specs/* and create an IMPLEMENTATION_PLAN.md with prioritized tasks" | claude
```

### 7. Make Loop Executable

```bash
chmod +x loop.sh
```

### 8. Run Ralph

```bash
# From Cursor's terminal
./loop.sh 20
```

## How It Works

1. **Loop starts** → feeds PROMPT.md to Claude CLI
2. **Claude orients** → reads AGENTS.md, IMPLEMENTATION_PLAN.md, specs/*
3. **Claude loads rules** → reads relevant `.cursor/rules/*.mdc` based on task
4. **Claude implements** → one task, following loaded rules
5. **Claude validates** → runs typecheck, tests, lint
6. **Claude commits** → if validation passes
7. **Claude updates** → marks task done, notes learnings
8. **Loop repeats** → fresh context, picks next task

## Monitoring

In another terminal tab:

```bash
# Watch task progress
watch -n 5 'cat IMPLEMENTATION_PLAN.md'

# Watch commits
watch -n 5 'git log --oneline -10'

# Tail the log
tail -f ralph.log
```

## Stopping Ralph

- `Ctrl+C` to stop the loop
- `git reset --hard HEAD~1` to undo last commit if needed
- Edit IMPLEMENTATION_PLAN.md to reprioritize

## Tips

1. **Start small** - Begin with 3-5 tasks, not 50
2. **Clear acceptance criteria** - Vague specs = vague output
3. **Watch the first few iterations** - Adjust PROMPT.md if Ralph goes off-track
4. **Rules are steering** - If output quality is wrong, the rules need tuning
5. **Plan is disposable** - Delete and regenerate if it's wrong
