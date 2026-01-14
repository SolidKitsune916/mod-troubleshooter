---
description: Follow React hooks standards when creating or modifying custom hooks
globs: ["src/hooks/**/*.ts", "src/**/*use*.ts"]
---
# React Hooks Standards

## Context
- Custom React hooks with TypeScript
- Hooks encapsulate reusable stateful logic
- Follow React hooks rules and conventions

## Requirements

### Naming & Structure
- Prefix all hooks with `use` (e.g., useAuth, useFetch)
- Define explicit return type interface
- Add JSDoc documentation with usage examples
- Keep hooks focused on single responsibility

### Implementation
- Always include cleanup in useEffect return function
- Handle race conditions in async effects with cancelled flag
- Use useCallback for functions passed to children
- Use useMemo only for expensive computations with measured need
- Implement proper error handling with typed errors

### Common Patterns
- Return object with named properties (not positional tuple for >2 values)
- Include refetch/retry functions for data hooks
- Provide loading, error, and data states for async hooks
- Use AbortController for cancelable fetch requests

## Examples

<example>
```tsx
interface UseFetchResult<T> {
  data: T | null;
  isLoading: boolean;
  error: Error | null;
  refetch: () => void;
}

/**
 * Fetches data from a URL with loading and error states
 * @example
 * const { data, isLoading, error } = useFetch<User[]>('/api/users');
 */
export function useFetch<T>(url: string): UseFetchResult<T> {
  const [data, setData] = useState<T | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchData = useCallback(async () => {
    const controller = new AbortController();
    setIsLoading(true);
    setError(null);
    
    try {
      const response = await fetch(url, { signal: controller.signal });
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const json = await response.json();
      setData(json);
    } catch (e) {
      if (e instanceof Error && e.name !== 'AbortError') {
        setError(e);
      }
    } finally {
      setIsLoading(false);
    }
    
    return () => controller.abort();
  }, [url]);

  useEffect(() => {
    const cleanup = fetchData();
    return () => { cleanup.then(fn => fn?.()); };
  }, [fetchData]);

  return { data, isLoading, error, refetch: fetchData };
}
```
</example>

<example type="invalid">
```tsx
// Bad - No cleanup, no error handling, no types
function useFetch(url) {
  const [data, setData] = useState();
  
  useEffect(() => {
    fetch(url).then(r => r.json()).then(setData);
  }, [url]);
  
  return data;
}
```
</example>
