---
description: Follow testing standards when creating or modifying tests
globs: ["src/**/*.test.{ts,tsx}", "src/**/*.spec.{ts,tsx}"]
---
# Testing Standards

## Context
- Vitest for test runner
- React Testing Library for component tests
- Test behavior, not implementation details

## Requirements

### Test Structure
- Use describe blocks to group related tests
- Use clear, descriptive test names
- Follow Arrange-Act-Assert pattern
- Keep tests focused on single concern
- Test both happy and sad paths

### React Testing Library
- Query by accessible roles first (getByRole)
- Use userEvent (not fireEvent) for interactions
- All userEvent methods are async - always await them
- Test what users see and do, not internal state
- Use screen queries for clearer tests

### Accessibility Testing
- Test keyboard navigation
- Verify ARIA attributes are correct
- Check focus management in modals
- Ensure error messages are announced

### Component Tests
- Test loading, error, and success states
- Test form validation behavior
- Mock API calls with MSW or vi.mock
- Create custom render with providers

## Examples

<example>
```tsx
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { LoginForm } from './LoginForm';

describe('LoginForm', () => {
  it('shows validation error for invalid email', async () => {
    const user = userEvent.setup();
    render(<LoginForm />);
    
    const emailInput = screen.getByLabelText(/email/i);
    const submitButton = screen.getByRole('button', { name: /sign in/i });
    
    await user.type(emailInput, 'invalid');
    await user.click(submitButton);
    
    expect(screen.getByRole('alert')).toHaveTextContent(/valid email/i);
    expect(emailInput).toHaveAttribute('aria-invalid', 'true');
  });

  it('submits form with valid data', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<LoginForm onSubmit={onSubmit} />);
    
    await user.type(screen.getByLabelText(/email/i), 'test@example.com');
    await user.type(screen.getByLabelText(/password/i), 'Password123');
    await user.click(screen.getByRole('button', { name: /sign in/i }));
    
    expect(onSubmit).toHaveBeenCalledWith({
      email: 'test@example.com',
      password: 'Password123',
    });
  });
});
```
</example>

<example type="invalid">
```tsx
// Bad - Testing implementation, not behavior
it('sets state correctly', () => {
  const { result } = renderHook(() => useForm());
  act(() => result.current.setEmail('test'));
  expect(result.current.email).toBe('test');
});

// Bad - Using fireEvent instead of userEvent
it('submits', () => {
  render(<Form />);
  fireEvent.change(screen.getByRole('textbox'), { target: { value: 'x' } });
  fireEvent.click(screen.getByRole('button'));
});
```
</example>
