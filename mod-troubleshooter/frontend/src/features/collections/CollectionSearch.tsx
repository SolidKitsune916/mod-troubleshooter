import { useState, useId, type FormEvent } from 'react';

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
    <form onSubmit={handleSubmit} className="flex flex-col gap-2">
      <label htmlFor={inputId} className="text-sm font-medium text-text-secondary">
        Collection URL or Slug
      </label>
      <div className="flex gap-2">
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
          className="flex-1 min-h-11 px-4 py-2 rounded-sm
            bg-bg-secondary text-text-primary placeholder:text-text-muted
            border border-border
            focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
            disabled:opacity-50 disabled:cursor-not-allowed"
        />
        <button
          type="submit"
          disabled={isLoading || !input.trim()}
          className="min-h-11 min-w-11 px-6 py-2 rounded-sm
            bg-primary text-white font-medium
            hover:bg-primary-dark
            focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
            disabled:opacity-50 disabled:cursor-not-allowed
            transition-colors motion-reduce:transition-none"
        >
          {isLoading ? 'Loading...' : 'Load Collection'}
        </button>
      </div>
      {error && (
        <span id={errorId} role="alert" className="text-sm text-error">
          {error}
        </span>
      )}
    </form>
  );
};
