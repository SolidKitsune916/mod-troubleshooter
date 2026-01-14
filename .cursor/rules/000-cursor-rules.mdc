---
description: Use ALWAYS when asked to CREATE A RULE or UPDATE A RULE or taught a lesson from the user that should be retained as a new rule for Cursor
globs: [".cursor/rules/*.mdc"]
---
# Cursor Rules Format
## Core Structure

```mdc
---
description: ACTION when TRIGGER to OUTCOME
globs: *.mdc
---

# Rule Title

## Context
- When to apply this rule
- Prerequisites or conditions

## Requirements
- Concise, actionable items
- Each requirement must be testable

## Examples
<example>
Good concise example with explanation
</example>

<example type="invalid">
Invalid concise example with explanation
</example>
```

## File Organization

### Location
- Path: `.cursor/rules/`
- Extension: `.mdc`

### Naming Convention
PREFIX-name.mdc where PREFIX is:
- 0XX: Core standards
- 1XX: Tool configs
- 3XX: Testing standards
- 1XXX: Language rules (TypeScript, CSS)
- 2XXX: Framework rules (React, Vite)
- 8XX: Workflows
- 9XX: Templates
- _name.mdc: Private rules

### Glob Pattern Examples
- Core standards: .cursor/rules/*.mdc
- TypeScript: src/**/*.{ts,tsx}
- React components: src/components/**/*.tsx
- Hooks: src/hooks/**/*.ts
- Testing: **/*.test.{ts,tsx}
- Styles: src/**/*.css
- Configuration: *.config.{ts,js,json}

## Required Fields

### Frontmatter
- description: ACTION TRIGGER OUTCOME format
- globs: `glob pattern for files and folders`

### Body
- context: Usage conditions
- requirements: Actionable items
- examples: Both valid and invalid

## Formatting Guidelines

- Use Concise Markdown primarily
- XML tags limited to: <example>, <danger>, <required>, <critical>
- Always indent content within XML tags by 2 spaces
- Keep rules as short as possible
- Use Emojis where appropriate to convey meaning
- Keep examples as short as possible

## AI Optimization Tips

1. Use precise, deterministic ACTION TRIGGER OUTCOME format
2. Provide concise positive and negative examples
3. Optimize for AI context window efficiency
4. Remove any non-essential or redundant information
5. Use standard glob patterns without quotes

<critical>
  - NEVER include verbose explanations that increase AI token overhead
  - Keep file as short and to the point as possible
  - Frontmatter can ONLY have the fields description and globs
</critical>
