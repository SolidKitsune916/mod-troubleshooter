---
description: Follow TypeScript strict mode standards when writing TypeScript code
globs: ["src/**/*.{ts,tsx}", "*.config.ts"]
---
# TypeScript Standards

## Context
- TypeScript strict mode enabled
- No `any` types - use proper typing
- Leverage type inference where appropriate

## Requirements

### Type Definitions
- Define explicit interfaces for component props
- Use `type` for unions/intersections, `interface` for objects
- Export shared types from @types/ directory
- Use `import type` for type-only imports

### Type Safety
- Never use `any` - use `unknown` and narrow with type guards
- Use discriminated unions for state variants
- Prefer const assertions for literal types
- Use satisfies for type checking without widening

### Best Practices
- Let TypeScript infer simple types (useState(0), useState(''))
- Explicit types when inference isn't enough (useState<User | null>(null))
- Use generics for reusable functions and hooks
- Add JSDoc comments for exported functions

### Common Patterns

```tsx
// Discriminated unions for state
type RequestState<T> =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'success'; data: T }
  | { status: 'error'; error: Error };

// Const assertion for string literals
const ROLES = ['admin', 'user', 'guest'] as const;
type Role = typeof ROLES[number];

// Generic hook
function useLocalStorage<T>(key: string, initial: T): [T, (v: T) => void];

// Type guard
function isUser(obj: unknown): obj is User {
  return typeof obj === 'object' && obj !== null && 'id' in obj;
}
```

## Examples

<example>
```tsx
// Good - Explicit prop types, discriminated union
interface UserListProps {
  initialFilter?: string;
  onSelect: (user: User) => void;
}

type ListState<T> = 
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'success'; data: T[] }
  | { status: 'error'; error: Error };

export const UserList: React.FC<UserListProps> = ({ initialFilter = '', onSelect }) => {
  const [state, setState] = useState<ListState<User>>({ status: 'idle' });

  // TypeScript narrows based on status
  if (state.status === 'success') {
    return <ul>{state.data.map(u => /* ... */)}</ul>;
  }
};
```
</example>

<example type="invalid">
```tsx
// Bad - any types, no proper typing
const UserList = ({ onSelect }: any) => {
  const [users, setUsers] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  
  // Multiple boolean flags instead of discriminated union
  if (loading) return <Spinner />;
  if (error) return <Error />;
};
```
</example>
