import { useState } from 'react';
import type { ViewerCollection } from '@/types';
import './Sidebar.css';

interface SidebarProps {
  collections: ViewerCollection[];
  selectedCollections: Set<string>;
  onSelectCollection: (id: string) => void;
  onSelectAll: () => void;
  onDeselectAll: () => void;
  onShowAllCollections: () => void;
  onShowSingleCollection: (id: string) => void;
  isAllCollectionsActive: boolean;
  currentCollectionId?: string;
  categories?: string[];
  onCategoryClick?: (categoryName: string) => void;
  isMobileOpen?: boolean;
  onMobileClose?: () => void;
}

export function Sidebar({
  collections,
  selectedCollections,
  onSelectCollection,
  onSelectAll,
  onDeselectAll,
  onShowAllCollections,
  onShowSingleCollection,
  isAllCollectionsActive,
  currentCollectionId,
  categories,
  onCategoryClick,
  isMobileOpen,
  onMobileClose,
}: SidebarProps) {
  const [isCollapsed, setIsCollapsed] = useState(false);

  const handleToggle = () => {
    if (window.innerWidth <= 768) {
      onMobileClose?.();
    } else {
      setIsCollapsed(!isCollapsed);
    }
  };

  return (
    <>
      {isMobileOpen && (
        <div
          className="sidebar-overlay"
          onClick={onMobileClose}
          aria-hidden="true"
        />
      )}
      <aside
        className={`sidebar ${isCollapsed ? 'collapsed' : ''} ${isMobileOpen ? 'open' : ''}`}
        id="sidebar-nav"
        role="navigation"
        aria-label="Collections navigation"
      >
        <div className="sidebar-header">
          <button
            className="sidebar-toggle"
            onClick={handleToggle}
            aria-label={isMobileOpen ? 'Close sidebar' : 'Toggle sidebar'}
            aria-expanded={!isCollapsed}
          >
            {isMobileOpen ? (
              <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z" />
              </svg>
            ) : (
              <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                <path d="M3 18h18v-2H3v2zm0-5h18v-2H3v2zm0-7v2h18V6H3z" />
              </svg>
            )}
            <span className="sidebar-toggle-text">
              {isMobileOpen ? 'Close' : 'Collections'}
            </span>
          </button>
        </div>
        {!isCollapsed && (
          <div className="sidebar-content">
            <button
              className={`all-collections-btn ${isAllCollectionsActive ? 'active' : ''}`}
              onClick={onShowAllCollections}
              aria-pressed={isAllCollectionsActive}
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                <path d="M4 6H2v14c0 1.1.9 2 2 2h14v-2H4V6zm16-4H8c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm0 14H8V4h12v12z" />
              </svg>
              All Collections
            </button>

            {categories && categories.length > 0 && (
              <div className="category-jump-section">
                <h2>Jump to Category</h2>
                <div className="category-pills">
                  {categories.map((cat) => (
                    <button
                      key={cat}
                      className="category-pill"
                      onClick={() => onCategoryClick?.(cat)}
                      type="button"
                    >
                      {cat}
                    </button>
                  ))}
                </div>
              </div>
            )}

            <div className="selection-controls">
              <button className="selection-btn" onClick={onSelectAll} aria-label="Select all collections">
                Select All
              </button>
              <button className="selection-btn" onClick={onDeselectAll} aria-label="Deselect all collections">
                Deselect All
              </button>
            </div>

            <h2>Collections</h2>
            <div className="collections-list">
              {collections.map((collection) => {
                const collectionId = collection.slug || collection.id;
                const isSelected = selectedCollections.has(collection.id);
                const isActive = currentCollectionId === collection.id;
                const modCount = collection.modCount || collection.mods?.length || 0;
                return (
                  <div
                    key={collectionId}
                    className={`collection-nav-item ${isActive ? 'active' : ''}`}
                    data-collection={collectionId}
                  >
                    <input
                      type="checkbox"
                      className="collection-checkbox"
                      id={`check-${collectionId}`}
                      checked={isSelected}
                      onChange={() => onSelectCollection(collection.id)}
                      aria-label={`Select ${collection.name}`}
                    />
                    <label
                      htmlFor={`check-${collectionId}`}
                      className="collection-label"
                      onClick={(e) => {
                        e.stopPropagation();
                        onShowSingleCollection(collection.id);
                      }}
                    >
                      {collection.tileImage?.url && (
                        <img
                          src={collection.tileImage.url}
                          alt=""
                          className="nav-thumb"
                          loading="lazy"
                          onError={(e) => {
                            (e.target as HTMLImageElement).style.display = 'none';
                          }}
                        />
                      )}
                      <div className="nav-info">
                        <span className="nav-name">{collection.name}</span>
                        <span className="nav-count">{modCount} mods</span>
                      </div>
                    </label>
                  </div>
                );
              })}
            </div>
          </div>
        )}
      </aside>
    </>
  );
}
