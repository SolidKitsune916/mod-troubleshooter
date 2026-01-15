import { useState, useId, type FormEvent } from 'react';
import styles from './CollectionSearch.module.css';

interface CollectionSearchProps {
  onSearch: (slug: string) => void;
  isLoading?: boolean;
}

/** Extracts collection slug from URL or returns slug as-is */
function parseCollectionSlug(input: string): string {
  const trimmed = input.trim();

  // Match Nexus collection URL pattern: nexusmods.com/.../collections/{slug}
  const urlMatch = trimmed.match(
    /nexusmods\.com\/[^/]+\/collections\/([^/?#]+)/i,
  );
  if (urlMatch) {
    return urlMatch[1];
  }

  return trimmed;
}

/** Search form for entering a collection URL or slug */
export const CollectionSearch: React.FC<CollectionSearchProps> = ({
  onSearch,
  isLoading = false,
}) => {
  const [input, setInput] = useState('');
  const [error, setError] = useState<string | null>(null);
  const inputId = useId();
  const errorId = useId();

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);

    const slug = parseCollectionSlug(input);
    if (!slug) {
      setError('Please enter a collection URL or slug');
      return;
    }

    onSearch(slug);
  };

  return (
    <form onSubmit={handleSubmit} className={styles.form}>
      <label htmlFor={inputId} className={styles.label}>
        Collection URL or Slug
      </label>
      <div className={styles.inputGroup}>
        <input
          id={inputId}
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="https://nexusmods.com/skyrimspecialedition/collections/... or slug"
          aria-required="true"
          aria-invalid={!!error}
          aria-describedby={error ? errorId : undefined}
          disabled={isLoading}
          className={styles.input}
        />
        <button
          type="submit"
          disabled={isLoading || !input.trim()}
          className={styles.submitButton}
          aria-busy={isLoading}
        >
          <span className={styles.buttonContent}>
            {isLoading && (
              <span
                className={styles.spinner}
                aria-hidden="true"
              />
            )}
            {isLoading ? 'Loading' : 'Load Collection'}
          </span>
        </button>
      </div>
      {error && (
        <span id={errorId} role="alert" className={styles.error}>
          {error}
        </span>
      )}
    </form>
  );
};
