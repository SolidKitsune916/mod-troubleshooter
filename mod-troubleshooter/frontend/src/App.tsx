import { useState, useCallback, useMemo } from 'react';

import { CollectionBrowser } from '@features/collections/index.ts';
import { SettingsPage } from '@features/settings/index.ts';
import { Header } from '@components/Header/index.ts';
import { Sidebar } from '@components/Sidebar/index.ts';
import { SkipLinks } from '@components/SkipLinks/index.ts';
import { KeyboardShortcutsHelp } from '@components/KeyboardShortcutsHelp/index.ts';
import {
  useViewerCollections,
  useSearch,
  useMobileMenu,
  useKeyboardShortcuts,
  createDefaultShortcuts,
} from '@hooks/index.ts';

import styles from './App.module.css';

/** Page type for simple state-based routing */
type Page = 'collections' | 'settings';

/** Main application component */
function App() {
  const [currentPage, setCurrentPage] = useState<Page>('collections');
  const { isSidebarOpen, toggleSidebar, closeSidebar } = useMobileMenu();
  const { searchQuery, setSearchQuery } = useSearch();

  const {
    data,
    loading,
    error,
    currentGame,
    selectedCollections,
    selectedCollectionsList,
    currentCollection,
    currentView,
    availableCategories,
    setCurrentGame,
    selectCollection,
    selectAll,
    deselectAll,
    showAllCollections,
    showSingleCollection,
  } = useViewerCollections();

  // Keyboard shortcuts
  const shortcuts = useMemo(() => createDefaultShortcuts({
    onGoToCollections: () => setCurrentPage('collections'),
    onGoToSettings: () => setCurrentPage('settings'),
    onFocusSearch: () => {
      // Focus the search input if it exists
      const searchInput = document.querySelector<HTMLInputElement>('[data-search-input]');
      searchInput?.focus();
    },
  }), []);

  const { showHelp, closeHelp, pendingKey, shortcuts: registeredShortcuts } = useKeyboardShortcuts({
    enabled: true,
    shortcuts,
  });

  const handleCategoryClick = useCallback((categoryName: string) => {
    const categoryId = `category-${categoryName.replace(/[^a-z0-9]/gi, '-').toLowerCase()}`;
    const element = document.getElementById(categoryId);
    if (element) {
      element.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
    closeSidebar();
  }, [closeSidebar]);

  // Loading state
  if (loading) {
    return (
      <div className={styles.app}>
        <SkipLinks />
        <Header
          currentGame={currentGame}
          onGameChange={setCurrentGame}
          onMenuToggle={toggleSidebar}
        />
        <div className={styles.container}>
          <div className={styles.loadingState} role="status" aria-label="Loading collections">
            <div className={styles.loadingSpinner} />
            <p>Loading collections...</p>
          </div>
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className={styles.app}>
        <SkipLinks />
        <Header
          currentGame={currentGame}
          onGameChange={setCurrentGame}
          onMenuToggle={toggleSidebar}
        />
        <div className={styles.container}>
          <div className={styles.errorState} role="alert">
            <h2>Error loading data</h2>
            <p>{error}</p>
            <p>Make sure data files exist in the public/data folder or the API is running.</p>
          </div>
        </div>
      </div>
    );
  }

  // No data state
  if (!data) {
    return null;
  }

  return (
    <div className={styles.app}>
      <SkipLinks />

      {/* Keyboard shortcuts help overlay */}
      <KeyboardShortcutsHelp
        isOpen={showHelp}
        onClose={closeHelp}
        shortcuts={registeredShortcuts}
        pendingKey={pendingKey}
      />

      <Header
        currentGame={currentGame}
        onGameChange={setCurrentGame}
        onMenuToggle={toggleSidebar}
        collectionCount={data.collections.length}
        totalMods={data.totalMods}
      />

      <div className={styles.container}>
        <Sidebar
          collections={data.collections}
          selectedCollections={selectedCollections}
          onSelectCollection={selectCollection}
          onSelectAll={selectAll}
          onDeselectAll={deselectAll}
          onShowAllCollections={showAllCollections}
          onShowSingleCollection={showSingleCollection}
          isAllCollectionsActive={currentView === 'all'}
          currentCollectionId={currentView !== 'all' ? currentView : undefined}
          categories={availableCategories}
          onCategoryClick={handleCategoryClick}
          isMobileOpen={isSidebarOpen}
          onMobileClose={closeSidebar}
        />

        <main id="main-content" className={styles.main} role="main" tabIndex={-1}>
          {currentPage === 'collections' && (
            <CollectionBrowser
              gameId={currentGame}
              collections={selectedCollectionsList}
              currentCollection={currentCollection}
              currentView={currentView}
              searchQuery={searchQuery}
              onSearchChange={setSearchQuery}
            />
          )}
          {currentPage === 'settings' && (
            <SettingsPage onBack={() => setCurrentPage('collections')} />
          )}
        </main>
      </div>
    </div>
  );
}

export default App;
