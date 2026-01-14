import { useState } from 'react';

import { CollectionBrowser } from '@features/collections/index.ts';
import { SettingsPage } from '@features/settings/index.ts';

/** Page type for simple state-based routing */
type Page = 'collections' | 'settings';

/** Main application component */
function App() {
  const [currentPage, setCurrentPage] = useState<Page>('collections');

  return (
    <div className="min-h-screen bg-bg-primary">
      {/* Skip link for accessibility */}
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:absolute focus:top-4 focus:left-4
          px-4 py-2 bg-accent text-white rounded-sm z-50
          focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2"
      >
        Skip to main content
      </a>

      <header className="border-b border-border bg-bg-card">
        <div className="max-w-6xl mx-auto px-8 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold text-text-primary">
                Mod Troubleshooter
              </h1>
              <p className="text-sm text-text-secondary">
                Visualize, analyze, and troubleshoot Skyrim SE mod collections
              </p>
            </div>

            <nav aria-label="Main navigation">
              <ul className="flex items-center gap-2">
                <li>
                  <button
                    onClick={() => setCurrentPage('collections')}
                    aria-current={currentPage === 'collections' ? 'page' : undefined}
                    className={`min-h-11 px-4 py-2 rounded-sm font-medium
                      focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
                      transition-colors motion-reduce:transition-none
                      ${
                        currentPage === 'collections'
                          ? 'bg-accent text-white'
                          : 'text-text-secondary hover:text-text-primary hover:bg-bg-secondary'
                      }`}
                  >
                    Collections
                  </button>
                </li>
                <li>
                  <button
                    onClick={() => setCurrentPage('settings')}
                    aria-current={currentPage === 'settings' ? 'page' : undefined}
                    className={`min-h-11 px-4 py-2 rounded-sm font-medium
                      focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
                      transition-colors motion-reduce:transition-none
                      ${
                        currentPage === 'settings'
                          ? 'bg-accent text-white'
                          : 'text-text-secondary hover:text-text-primary hover:bg-bg-secondary'
                      }`}
                  >
                    Settings
                  </button>
                </li>
              </ul>
            </nav>
          </div>
        </div>
      </header>

      <main id="main-content" className="max-w-6xl mx-auto p-8">
        {currentPage === 'collections' && <CollectionBrowser />}
        {currentPage === 'settings' && (
          <SettingsPage onBack={() => setCurrentPage('collections')} />
        )}
      </main>
    </div>
  );
}

export default App;
