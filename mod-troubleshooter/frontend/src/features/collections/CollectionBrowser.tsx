import { useState, useMemo } from 'react';
import type { GameId, ViewerCollection, ViewerMod } from '@/types';

import { useCollection } from '@hooks/useCollections.ts';
import { ApiError } from '@services/api.ts';
import { LoadOrderView } from '@features/loadorder/index.ts';
import { ConflictView } from '@features/conflicts/index.ts';
import { HeroBanner } from '@components/HeroBanner/index.ts';
import { SearchBar } from '@components/SearchBar/index.ts';
import { deduplicateMods, groupModsByCategory } from '@/utils/dataLoader';

import { CollectionSearch } from './CollectionSearch.tsx';
import { CollectionHeader } from './CollectionHeader.tsx';
import styles from './CollectionBrowser.module.css';

/** View modes available in the collection browser */
type ViewMode = 'mods' | 'loadorder' | 'conflicts';

interface CollectionBrowserProps {
  gameId: GameId;
  collections: ViewerCollection[];
  currentCollection: ViewerCollection | null;
  currentView: 'all' | string;
  searchQuery: string;
  onSearchChange: (query: string) => void;
}

/** Loading skeleton for collection */
const CollectionSkeleton: React.FC = () => (
  <div className={styles.skeleton}>
    <div className={styles.skeletonHeader}>
      <div className={styles.skeletonImage} />
      <div className={styles.skeletonContent}>
        <div className={styles.skeletonLine} style={{ width: '66%' }} />
        <div className={styles.skeletonLine} style={{ width: '50%' }} />
        <div className={styles.skeletonLine} style={{ width: '100%' }} />
      </div>
    </div>
    <div className={styles.skeletonList}>
      {[1, 2, 3].map((i) => (
        <div key={i} className={styles.skeletonItem}>
          <div className={styles.skeletonThumb} />
          <div className={styles.skeletonItemContent}>
            <div className={styles.skeletonLine} style={{ width: '33%' }} />
            <div className={styles.skeletonLine} style={{ width: '25%' }} />
            <div className={styles.skeletonLine} style={{ width: '100%' }} />
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
    <div role="alert" className={styles.errorDisplay}>
      <p className={styles.errorMessage}>{message}</p>
      <button onClick={onRetry} className={styles.retryButton}>
        Try Again
      </button>
    </div>
  );
};

/** View mode tab navigation */
interface ViewModeTabsProps {
  currentMode: ViewMode;
  onModeChange: (mode: ViewMode) => void;
}

const ViewModeTabs: React.FC<ViewModeTabsProps> = ({ currentMode, onModeChange }) => (
  <nav aria-label="Collection view modes" className={styles.viewModeTabs}>
    <ul className={styles.tabList}>
      <li>
        <button
          onClick={() => onModeChange('mods')}
          aria-current={currentMode === 'mods' ? 'page' : undefined}
          className={`${styles.tabButton} ${currentMode === 'mods' ? styles.tabButtonActive : ''}`}
        >
          Mod Files
        </button>
      </li>
      <li>
        <button
          onClick={() => onModeChange('loadorder')}
          aria-current={currentMode === 'loadorder' ? 'page' : undefined}
          className={`${styles.tabButton} ${currentMode === 'loadorder' ? styles.tabButtonActive : ''}`}
        >
          Load Order
        </button>
      </li>
      <li>
        <button
          onClick={() => onModeChange('conflicts')}
          aria-current={currentMode === 'conflicts' ? 'page' : undefined}
          className={`${styles.tabButton} ${currentMode === 'conflicts' ? styles.tabButtonActive : ''}`}
        >
          Conflicts
        </button>
      </li>
    </ul>
  </nav>
);

/** Category section for grouped mod display */
interface CategorySectionProps {
  categoryName: string;
  mods: ViewerMod[];
}

const CategorySection: React.FC<CategorySectionProps> = ({ categoryName, mods }) => {
  const categoryId = `category-${categoryName.replace(/[^a-z0-9]/gi, '-').toLowerCase()}`;
  
  return (
    <section id={categoryId} className={styles.categorySection} aria-labelledby={`${categoryId}-title`}>
      <h3 id={`${categoryId}-title`} className={styles.categorySectionTitle}>
        {categoryName}
        <span className={styles.categorySectionCount}>({mods.length})</span>
      </h3>
      <div className={styles.modGrid}>
        {mods.map((mod) => (
          <article key={mod.modId || `${mod.name}-${mod.author}`} className={styles.modCard}>
            {mod.pictureUrl ? (
              <img
                src={mod.pictureUrl}
                alt=""
                className={styles.modCardImage}
                loading="lazy"
              />
            ) : (
              <div className={styles.modCardImagePlaceholder} aria-hidden="true">
                No Image
              </div>
            )}
            <div className={styles.modCardContent}>
              <h4 className={styles.modCardTitle}>{mod.name}</h4>
              <div className={styles.modCardMeta}>
                {mod.version && <span>v{mod.version}</span>}
                <span>by {mod.author || mod.uploader?.name || 'Unknown'}</span>
              </div>
              {mod.summary && (
                <p className={styles.modCardSummary}>{mod.summary}</p>
              )}
              {mod.optional && (
                <span className={styles.optionalBadge}>Optional</span>
              )}
            </div>
          </article>
        ))}
      </div>
    </section>
  );
};

/** Main collection browser container */
export const CollectionBrowser: React.FC<CollectionBrowserProps> = ({
  gameId,
  collections,
  currentCollection,
  currentView,
  searchQuery,
  onSearchChange,
}) => {
  const [apiSlug, setApiSlug] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<ViewMode>('mods');

  // API-based collection fetching (for advanced features)
  const {
    data: apiCollection,
    isLoading: apiLoading,
    error: apiError,
    refetch: apiRefetch,
  } = useCollection(apiSlug ?? '', !!apiSlug);

  const handleApiSearch = (newSlug: string) => {
    setApiSlug(newSlug);
    setViewMode('mods');
  };

  // Get mods based on current view
  const displayMods = useMemo(() => {
    if (currentView === 'all') {
      // Combine and deduplicate mods from all selected collections
      const allMods = collections.flatMap((c) => c.mods);
      return deduplicateMods(allMods);
    } else if (currentCollection) {
      return currentCollection.mods;
    }
    return [];
  }, [currentView, collections, currentCollection]);

  // Filter mods based on search query
  const filteredMods = useMemo(() => {
    if (!searchQuery.trim()) return displayMods;
    const query = searchQuery.toLowerCase();
    return displayMods.filter(
      (mod) =>
        mod.name.toLowerCase().includes(query) ||
        mod.summary?.toLowerCase().includes(query) ||
        mod.author?.toLowerCase().includes(query) ||
        mod.uploader?.name?.toLowerCase().includes(query) ||
        mod.category?.toLowerCase().includes(query)
    );
  }, [displayMods, searchQuery]);

  // Group mods by category
  const groupedMods = useMemo(() => {
    return groupModsByCategory(filteredMods);
  }, [filteredMods]);

  // Get revision info for API features
  const revisionNumber = apiCollection?.latestPublishedRevision?.revisionNumber ?? 0;

  return (
    <div className={styles.collectionBrowser}>
      <HeroBanner gameId={gameId} />
      <SearchBar value={searchQuery} onChange={onSearchChange} />

      <div className={styles.contentArea}>
        {/* API search for advanced features */}
        <div className={styles.apiSection}>
          <CollectionSearch onSearch={handleApiSearch} isLoading={apiLoading} />

          {apiLoading && <CollectionSkeleton />}

          {apiError && !apiLoading && (
            <ErrorDisplay error={apiError} onRetry={() => apiRefetch()} />
          )}

          {apiCollection && !apiLoading && !apiError && (
            <>
              <div aria-live="polite" className={styles.srOnly}>
                Loaded {apiCollection.name} with{' '}
                {apiCollection.latestPublishedRevision?.modFiles.length ?? 0} mods
              </div>
              <CollectionHeader collection={apiCollection} />
              <ViewModeTabs currentMode={viewMode} onModeChange={setViewMode} />
              {viewMode === 'loadorder' && apiSlug && revisionNumber > 0 && (
                <LoadOrderView slug={apiSlug} revision={revisionNumber} />
              )}
              {viewMode === 'conflicts' && apiSlug && revisionNumber > 0 && (
                <ConflictView slug={apiSlug} revision={revisionNumber} />
              )}
            </>
          )}
        </div>

        {/* Viewer mode - show collections from JSON data */}
        {collections.length > 0 && !apiCollection && (
          <div className={styles.viewerSection}>
            <div className={styles.viewerHeader}>
              <h2 className={styles.viewerTitle}>
                {currentView === 'all'
                  ? `Browsing ${collections.length} Collection${collections.length !== 1 ? 's' : ''}`
                  : currentCollection?.name ?? 'Collection'}
              </h2>
              <p className={styles.viewerSubtitle}>
                {filteredMods.length} mod{filteredMods.length !== 1 ? 's' : ''}
                {searchQuery && ` matching "${searchQuery}"`}
              </p>
            </div>

            {groupedMods.length > 0 ? (
              <div className={styles.categoriesContainer}>
                {groupedMods.map(([category, mods]) => (
                  <CategorySection key={category} categoryName={category} mods={mods} />
                ))}
              </div>
            ) : (
              <div className={styles.emptyState}>
                <p className={styles.emptyStateTitle}>
                  {searchQuery ? 'No mods match your search' : 'No mods to display'}
                </p>
                <p className={styles.emptyStateSubtitle}>
                  {searchQuery
                    ? 'Try adjusting your search terms'
                    : 'Select collections from the sidebar to view their mods'}
                </p>
              </div>
            )}
          </div>
        )}

        {/* Empty state when no collections and no API search */}
        {collections.length === 0 && !apiSlug && !apiLoading && !apiError && (
          <div className={styles.emptyState}>
            <p className={styles.emptyStateTitle}>No collections selected</p>
            <p className={styles.emptyStateSubtitle}>
              Select collections from the sidebar or search for a collection above
            </p>
          </div>
        )}
      </div>
    </div>
  );
};
