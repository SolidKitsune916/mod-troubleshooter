import { useState, useEffect, useMemo, useCallback } from 'react';

import type { CollectionsData, ViewerCollection, GameId } from '@/types';
import { loadCollectionsData, deduplicateMods, getUniqueCategories } from '@/utils/dataLoader';

export interface UseViewerCollectionsResult {
  data: CollectionsData | null;
  loading: boolean;
  error: string | null;
  currentGame: GameId;
  selectedCollections: Set<string>;
  selectedCollectionsList: ViewerCollection[];
  currentCollection: ViewerCollection | null;
  currentView: 'all' | string;
  availableCategories: string[];
  setCurrentGame: (game: GameId) => void;
  selectCollection: (id: string) => void;
  selectAll: () => void;
  deselectAll: () => void;
  showAllCollections: () => void;
  showSingleCollection: (id: string) => void;
}

/**
 * Hook to manage viewer collections with multi-collection selection
 */
export function useViewerCollections(): UseViewerCollectionsResult {
  const [data, setData] = useState<CollectionsData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedCollections, setSelectedCollections] = useState<Set<string>>(new Set());
  const [currentView, setCurrentView] = useState<'all' | string>('all');
  const [currentGame, setCurrentGame] = useState<GameId>('skyrim');

  // Load data when game changes
  useEffect(() => {
    setLoading(true);
    setError(null);
    loadCollectionsData(currentGame)
      .then((loadedData) => {
        setData(loadedData);
        // Select all collections by default
        const allIds = new Set(loadedData.collections.map((c) => c.id));
        setSelectedCollections(allIds);
        setCurrentView('all');
        setLoading(false);
      })
      .catch((err) => {
        setError(err.message);
        setLoading(false);
      });
  }, [currentGame]);

  // Derived: selected collections list
  const selectedCollectionsList = useMemo(() => {
    if (!data) return [];
    return data.collections.filter((c) => selectedCollections.has(c.id));
  }, [data, selectedCollections]);

  // Derived: current collection (when viewing single)
  const currentCollection = useMemo(() => {
    if (currentView === 'all' || !data) return null;
    return data.collections.find((c) => c.id === currentView) || null;
  }, [currentView, data]);

  // Derived: available categories based on current view
  const availableCategories = useMemo(() => {
    if (currentView === 'all') {
      const allMods = selectedCollectionsList.flatMap((c) => c.mods);
      const uniqueMods = deduplicateMods(allMods);
      return getUniqueCategories(uniqueMods);
    } else if (currentCollection) {
      return getUniqueCategories(currentCollection.mods);
    }
    return [];
  }, [currentView, selectedCollectionsList, currentCollection]);

  // Actions
  const selectCollection = useCallback((id: string) => {
    setSelectedCollections((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
    setCurrentView('all');
  }, []);

  const selectAll = useCallback(() => {
    if (!data) return;
    const allIds = new Set(data.collections.map((c) => c.id));
    setSelectedCollections(allIds);
    setCurrentView('all');
  }, [data]);

  const deselectAll = useCallback(() => {
    setSelectedCollections(new Set());
  }, []);

  const showAllCollections = useCallback(() => {
    setSelectedCollections((prev) => {
      if (prev.size === 0 && data) {
        return new Set(data.collections.map((c) => c.id));
      }
      return prev;
    });
    setCurrentView('all');
  }, [data]);

  const showSingleCollection = useCallback((id: string) => {
    setCurrentView(id);
  }, []);

  const handleGameChange = useCallback((game: GameId) => {
    setCurrentGame(game);
  }, []);

  return {
    data,
    loading,
    error,
    currentGame,
    selectedCollections,
    selectedCollectionsList,
    currentCollection,
    currentView,
    availableCategories,
    setCurrentGame: handleGameChange,
    selectCollection,
    selectAll,
    deselectAll,
    showAllCollections,
    showSingleCollection,
  };
}
