import { useState } from 'react';

import { useCollection } from '@hooks/useCollections.ts';
import { ApiError } from '@services/api.ts';

import { CollectionSearch } from './CollectionSearch.tsx';
import { CollectionHeader } from './CollectionHeader.tsx';
import { ModList } from './ModList.tsx';

/** Loading skeleton for collection */
const CollectionSkeleton: React.FC = () => (
  <div className="space-y-6 animate-pulse">
    <div className="flex gap-6 p-6 rounded-sm bg-bg-card border border-border">
      <div className="w-32 h-32 rounded-xs bg-bg-secondary" />
      <div className="flex-1 space-y-3">
        <div className="h-8 w-2/3 bg-bg-secondary rounded-xs" />
        <div className="h-4 w-1/2 bg-bg-secondary rounded-xs" />
        <div className="h-4 w-full bg-bg-secondary rounded-xs" />
      </div>
    </div>
    <div className="space-y-3">
      {[1, 2, 3].map((i) => (
        <div
          key={i}
          className="flex gap-4 p-4 rounded-sm bg-bg-card border border-border"
        >
          <div className="w-20 h-20 rounded-xs bg-bg-secondary" />
          <div className="flex-1 space-y-2">
            <div className="h-5 w-1/3 bg-bg-secondary rounded-xs" />
            <div className="h-4 w-1/4 bg-bg-secondary rounded-xs" />
            <div className="h-4 w-full bg-bg-secondary rounded-xs" />
          </div>
        </div>
      ))}
    </div>
  </div>
);

/** Error display component */
interface ErrorDisplayProps {
  error: Error;
  onRetry: () => void;
}

const ErrorDisplay: React.FC<ErrorDisplayProps> = ({ error, onRetry }) => {
  let message = 'An unexpected error occurred.';

  if (error instanceof ApiError) {
    if (error.status === 404) {
      message = 'Collection not found. Please check the URL or slug.';
    } else if (error.status === 401 || error.status === 403) {
      message = 'API key is missing or invalid. Please configure the backend.';
    } else if (error.status >= 500) {
      message = 'Server error. Please try again later.';
    } else {
      message = error.message;
    }
  }

  return (
    <div
      role="alert"
      className="p-6 rounded-sm bg-error/10 border border-error text-center"
    >
      <p className="text-error font-medium mb-4">{message}</p>
      <button
        onClick={onRetry}
        className="min-h-11 px-6 py-2 rounded-sm
          bg-error text-white font-medium
          hover:bg-error/80
          focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
          transition-colors motion-reduce:transition-none"
      >
        Try Again
      </button>
    </div>
  );
};

/** Main collection browser container */
export const CollectionBrowser: React.FC = () => {
  const [slug, setSlug] = useState<string | null>(null);

  const {
    data: collection,
    isLoading,
    error,
    refetch,
  } = useCollection(slug ?? '', !!slug);

  const handleSearch = (newSlug: string) => {
    setSlug(newSlug);
  };

  const modFiles = collection?.latestPublishedRevision?.modFiles ?? [];

  return (
    <div className="space-y-6">
      <CollectionSearch onSearch={handleSearch} isLoading={isLoading} />

      {isLoading && <CollectionSkeleton />}

      {error && !isLoading && (
        <ErrorDisplay error={error} onRetry={() => refetch()} />
      )}

      {collection && !isLoading && !error && (
        <>
          <div aria-live="polite" className="sr-only">
            Loaded {collection.name} with {modFiles.length} mods
          </div>
          <CollectionHeader collection={collection} />
          <ModList modFiles={modFiles} />
        </>
      )}

      {!slug && !isLoading && !error && (
        <div className="text-center py-12 text-text-secondary">
          <p className="text-lg mb-2">Enter a collection to get started</p>
          <p className="text-sm text-text-muted">
            Paste a Nexus Mods collection URL or enter the collection slug
          </p>
        </div>
      )}
    </div>
  );
};
