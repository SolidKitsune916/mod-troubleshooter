---
description: Follow service and API integration standards when creating services
globs: ["src/services/**/*.ts", "src/api/**/*.ts"]
---
# Services & API Standards

## Context
- TypeScript service modules for API integration
- TanStack Query for server state management
- Type-safe API responses with Zod validation

## Requirements

### Service Structure
- Create service modules in src/services/ directory
- Name services descriptively: userService, authService
- Export typed functions, not classes
- Use Zod schemas to validate API responses
- Return typed data, throw typed errors

### API Integration
- Use fetch or axios with typed responses
- Centralize base URL and headers configuration
- Implement request/response interceptors for auth
- Handle errors consistently with custom error types

### TanStack Query Integration
- Create custom hooks that wrap useQuery/useMutation
- Define query keys as constants
- Set appropriate staleTime and cacheTime
- Implement optimistic updates for mutations

### Error Handling
- Define ApiError class with status and message
- Transform API errors to user-friendly messages
- Include retry logic for transient failures
- Log errors for debugging (not in production)

## Examples

<example>
```tsx
// services/userService.ts
import { z } from 'zod';

const UserSchema = z.object({
  id: z.string(),
  name: z.string(),
  email: z.string().email(),
});

const UsersResponseSchema = z.array(UserSchema);

export type User = z.infer<typeof UserSchema>;

export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = 'ApiError';
  }
}

export async function fetchUsers(): Promise<User[]> {
  const response = await fetch('/api/users');
  
  if (!response.ok) {
    throw new ApiError(response.status, 'Failed to fetch users');
  }
  
  const data = await response.json();
  return UsersResponseSchema.parse(data);
}

// hooks/useUsers.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

export const userKeys = {
  all: ['users'] as const,
  detail: (id: string) => ['users', id] as const,
};

export function useUsers() {
  return useQuery({
    queryKey: userKeys.all,
    queryFn: fetchUsers,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

export function useCreateUser() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: createUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.all });
    },
  });
}
```
</example>

<example type="invalid">
```tsx
// Bad - No types, no error handling, no validation
export async function getUsers() {
  const res = await fetch('/api/users');
  return res.json();
}

// Bad - Direct fetch in component
const Users = () => {
  const [users, setUsers] = useState([]);
  useEffect(() => {
    fetch('/api/users').then(r => r.json()).then(setUsers);
  }, []);
};
```
</example>
