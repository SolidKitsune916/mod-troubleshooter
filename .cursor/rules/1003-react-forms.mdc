---
description: Follow form and validation standards when creating or editing forms
globs: ["src/**/*Form*.tsx", "src/**/*form*.tsx", "src/components/forms/**/*.tsx"]
---
# Forms & Validation Standards

## Context
- React Hook Form with Zod validation
- WCAG 2.2 AA compliance for form accessibility
- Real-time inline validation with clear error messages

## Requirements

### Form Structure
- Use React Hook Form with zodResolver
- Define Zod schema for all form validation
- Infer TypeScript types from Zod schema with `z.infer<>`
- Group related fields with fieldset and legend
- Mark required fields with asterisk (*) and aria-required

### Accessibility
- Every input MUST have an associated label (htmlFor/id)
- Use aria-invalid="true" on fields with errors
- Use aria-describedby to link error messages to inputs
- Error messages must have role="alert"
- Use autocomplete attributes for user data fields
- Focus first error field on form submission failure

### Validation UX
- Validate on blur (not every keystroke)
- Display errors inline below the field
- Provide clear, actionable error messages
- Show success indicators sparingly
- Preserve entered data on validation failure

### Form Enhancements
- Show loading state during submission (disable submit button)
- Use appropriate input types (email, tel, url, number)
- Add password visibility toggle
- Show character count for limited fields
- Auto-save drafts for long forms

## Examples

<example>
```tsx
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';

const schema = z.object({
  email: z.string().email('Please enter a valid email address'),
  password: z.string()
    .min(8, 'Password must be at least 8 characters')
    .regex(/[A-Z]/, 'Must contain an uppercase letter'),
});

type FormData = z.infer<typeof schema>;

export const LoginForm: React.FC = () => {
  const { register, handleSubmit, formState: { errors, isSubmitting } } = 
    useForm<FormData>({ resolver: zodResolver(schema) });

  return (
    <form onSubmit={handleSubmit(onSubmit)} noValidate>
      <div className="form-field">
        <label htmlFor="email">Email *</label>
        <input
          id="email"
          type="email"
          autoComplete="email"
          aria-required="true"
          aria-invalid={!!errors.email}
          aria-describedby={errors.email ? 'email-error' : undefined}
          {...register('email')}
        />
        {errors.email && (
          <span id="email-error" role="alert" className="text-red-600">
            {errors.email.message}
          </span>
        )}
      </div>
      
      <button type="submit" disabled={isSubmitting}>
        {isSubmitting ? 'Signing in...' : 'Sign in'}
      </button>
    </form>
  );
};
```
</example>

<example type="invalid">
```tsx
// Bad - No validation, no accessibility, no error handling
const LoginForm = () => {
  const [email, setEmail] = useState('');
  return (
    <form>
      <input placeholder="Email" value={email} onChange={e => setEmail(e.target.value)} />
      <button>Submit</button>
    </form>
  );
};
```
</example>
