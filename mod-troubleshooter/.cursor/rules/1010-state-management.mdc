---
description: Follow state management patterns when managing application state
globs: ["src/store/**/*.ts", "src/context/**/*.tsx", "src/**/*Context*.tsx"]
---
# State Management Standards

## Context
- Use appropriate state management for scope
- TanStack Query for server state
- Zustand or Context for client state
- Local state for component-specific data

## Requirements

### State Selection
| State Type | Solution |
|------------|----------|
| Component UI state | useState |
| Derived/computed | useMemo |
| Server data | TanStack Query |
| Global client state | Zustand or Context |
| Form state | React Hook Form |
| URL state | URL params |

### Context API
- Create typed context with explicit interface
- Provide custom hook for consumption
- Throw error if used outside provider
- Keep context providers close to where needed

### Zustand
- Define typed store interface
- Use selectors to prevent unnecessary re-renders
- Keep actions inside store definition
- Use persist middleware for persistent state

### Best Practices
- Don't store derived values in state
- Colocate state as close as possible to usage
- Avoid prop drilling beyond 3 levels
- Use URL for shareable state (filters, pagination)

## Examples

<example>
```tsx
// Context API - Type-safe auth context
interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  login: (credentials: Credentials) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  
  const value: AuthContextType = {
    user,
    isAuthenticated: !!user,
    login: async (credentials) => {
      const user = await authService.login(credentials);
      setUser(user);
    },
    logout: () => setUser(null),
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}

// Zustand - Global UI state
interface UIState {
  sidebarOpen: boolean;
  theme: 'light' | 'dark';
  toggleSidebar: () => void;
  setTheme: (theme: 'light' | 'dark') => void;
}

export const useUIStore = create<UIState>()(
  persist(
    (set) => ({
      sidebarOpen: true,
      theme: 'light',
      toggleSidebar: () => set((s) => ({ sidebarOpen: !s.sidebarOpen })),
      setTheme: (theme) => set({ theme }),
    }),
    { name: 'ui-storage' }
  )
);

// Usage with selector (prevents re-renders)
const sidebarOpen = useUIStore((s) => s.sidebarOpen);
```
</example>

<example type="invalid">
```tsx
// Bad - Untyped context, no error handling
const AuthContext = createContext(null);

const useAuth = () => useContext(AuthContext); // Can return null!

// Bad - Storing derived state
const [items, setItems] = useState([]);
const [filteredItems, setFilteredItems] = useState([]);
useEffect(() => {
  setFilteredItems(items.filter(i => i.active));
}, [items]);
```
</example>
