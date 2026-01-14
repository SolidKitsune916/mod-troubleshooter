---
description: Apply quality of life UX patterns to all UI implementations
globs: ["src/**/*.tsx", "src/**/*.css"]
---
# Quality of Life UX Patterns

## Context
- QoL patterns improve user experience significantly
- Apply these patterns consistently across the application
- Focus on responsive, accessible, and performant UI

## Requirements

### Navigation & Scrolling
- Scroll to top on route change
- Restore scroll position on back navigation
- Show "back to top" button after 400px scroll
- Use sticky navigation with safe-area padding
- Support deep linking with URL parameters

### Loading & Feedback
- Show skeleton loaders for content (not spinners)
- Display loading state on buttons during async actions
- Provide toast notifications for action confirmations
- Position toasts consistently (top-right or bottom-right)
- Auto-dismiss success toasts after 4-6 seconds
- Persist error toasts until dismissed

### Empty & Error States
- Never show blank screens - always provide guidance
- Empty states: illustration + message + call-to-action
- Error states: user-friendly message + recovery action
- No results: suggest clearing filters or adjusting search

### Forms & Inputs
- Validate on blur, not every keystroke
- Show inline errors directly below fields
- Focus first error field on submission failure
- Display character count for limited fields
- Auto-save form drafts for long forms
- Warn before navigating away from unsaved changes

### Interactions
- Minimum touch targets: 44×44px
- Visible hover states on interactive elements
- Confirmation dialogs for destructive actions
- Undo option for deletions when possible
- Optimistic updates for toggle actions

### Responsive Design
- Use mobile-first breakpoint strategy
- Handle safe areas on notched devices
- Stack buttons vertically on mobile
- Collapse tables to cards on small screens

### Performance Perception
- Optimistic UI updates for quick feedback
- Progressive image loading (blur-up)
- Prefetch likely next pages on hover
- Show stale data immediately, refresh in background

### Theme & Motion
- Respect prefers-color-scheme for dark mode
- Respect prefers-reduced-motion for animations
- Provide manual theme toggle option
- Use CSS custom properties for theming

## Required Utilities

```css
/* Safe area handling */
.safe-padding {
  padding-left: env(safe-area-inset-left);
  padding-right: env(safe-area-inset-right);
  padding-bottom: env(safe-area-inset-bottom);
}

/* Fixed bottom elements */
.fixed-bottom {
  padding-bottom: calc(1rem + env(safe-area-inset-bottom));
}
```

```tsx
// Scroll to top on route change
function ScrollToTop() {
  const { pathname } = useLocation();
  useEffect(() => {
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }, [pathname]);
  return null;
}

// Reduced motion hook
function usePrefersReducedMotion() {
  const [prefersReduced, setPrefersReduced] = useState(
    () => window.matchMedia('(prefers-reduced-motion: reduce)').matches
  );
  useEffect(() => {
    const mq = window.matchMedia('(prefers-reduced-motion: reduce)');
    const handler = (e: MediaQueryListEvent) => setPrefersReduced(e.matches);
    mq.addEventListener('change', handler);
    return () => mq.removeEventListener('change', handler);
  }, []);
  return prefersReduced;
}
```

## Examples

<example>
```tsx
// Good - Empty state with guidance
const ContactList: React.FC = () => {
  const { data, isLoading } = useContacts();
  
  if (isLoading) return <ContactListSkeleton />;
  
  if (!data?.length) {
    return (
      <div className="text-center py-12">
        <UserPlusIcon className="mx-auto h-12 w-12 text-gray-400" />
        <h3 className="mt-2 text-lg font-medium">No contacts yet</h3>
        <p className="mt-1 text-gray-500">
          Get started by adding your first contact.
        </p>
        <Button className="mt-4" onClick={openAddContact}>
          Add Contact
        </Button>
      </div>
    );
  }
  
  return <ul>{/* contacts */}</ul>;
};
```
</example>

<example type="invalid">
```tsx
// Bad - Blank screen when no data
const ContactList = () => {
  const { data } = useContacts();
  return data?.length ? <ul>{/* contacts */}</ul> : null;
};
```
</example>
