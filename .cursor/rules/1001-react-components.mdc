---
description: Follow React component standards when creating or editing components
globs: ["src/components/**/*.tsx", "src/features/**/*.tsx"]
---
# React Component Standards

## Context
- React functional components with TypeScript
- WCAG 2.2 AA compliance required
- Tailwind CSS for styling

## Requirements

### Component Design
- Define props interface above component with explicit types
- Use React.FC<Props> for component typing
- Use useId() for form field IDs (accessibility)
- Keep components focused - split if over 150 lines
- Use composition pattern for complex UIs

### State Management
- Use discriminated unions for complex state:
  ```tsx
  type State = 
    | { status: 'idle' }
    | { status: 'loading' }
    | { status: 'success'; data: T }
    | { status: 'error'; error: Error };
  ```
- Handle all async states: loading, error, success
- Compute derived values directly, don't store in state
- Use Context for data passing through 3+ levels

### Accessibility Requirements
- Semantic HTML: button for actions, a for navigation
- aria-label on icon-only buttons
- aria-live="polite" for dynamic content updates
- role="alert" for error messages
- Visible focus indicators (:focus-visible)
- Keyboard event handlers for custom interactive elements

### Loading & Error States
- Show skeleton loaders for content, spinners for actions
- Display user-friendly error messages with recovery actions
- Use ErrorBoundary for component-level error catching

## Examples

<example>
```tsx
interface SearchResultsProps {
  query: string;
}

export const SearchResults: React.FC<SearchResultsProps> = ({ query }) => {
  const { data, isLoading, error } = useSearch(query);
  
  if (isLoading) return <SearchSkeleton />;
  if (error) return (
    <div role="alert" className="text-red-600">
      <p>Failed to load results. <button onClick={refetch}>Try again</button></p>
    </div>
  );
  if (!data?.length) return (
    <div className="text-center py-8">
      <p>No results found for "{query}"</p>
      <p className="text-gray-500">Try adjusting your search terms</p>
    </div>
  );
  
  return (
    <>
      <div aria-live="polite" className="sr-only">
        {data.length} results found
      </div>
      <ul role="list">{/* results */}</ul>
    </>
  );
};
```
</example>

<example type="invalid">
```tsx
// Bad - No loading/error states, poor accessibility
const SearchResults = ({ query }) => {
  const { data } = useSearch(query);
  return <div>{data.map(item => <div>{item.name}</div>)}</div>;
};
```
</example>
