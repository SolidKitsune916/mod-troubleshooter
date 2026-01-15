import type { CollectionsData, ViewerMod, GameId } from '@/types';
import { GAMES } from '@/types';

/**
 * Load collections data from JSON file for a specific game
 */
export async function loadCollectionsData(gameId: GameId = 'skyrim'): Promise<CollectionsData> {
  const game = GAMES.find(g => g.id === gameId);
  if (!game) {
    throw new Error(`Unknown game: ${gameId}`);
  }

  try {
    const response = await fetch(`/data/${game.filename}`);
    if (!response.ok) {
      throw new Error(`Failed to load data: ${response.statusText}`);
    }
    const data: CollectionsData = await response.json();
    return data;
  } catch (error) {
    console.error('Error loading collections data:', error);
    throw error;
  }
}

/**
 * Group mods by category
 */
export function groupModsByCategory(mods: ViewerMod[]): [string, ViewerMod[]][] {
  const categories = new Map<string, ViewerMod[]>();
  
  mods.forEach((mod) => {
    const category = mod.category || 'Uncategorized';
    if (!categories.has(category)) {
      categories.set(category, []);
    }
    categories.get(category)!.push(mod);
  });

  // Sort categories, with 'Uncategorized' at the end
  const sortedCategories = Array.from(categories.entries()).sort((a, b) => {
    const aLower = a[0].toLowerCase();
    const bLower = b[0].toLowerCase();
    if (aLower === 'uncategorized') return 1;
    if (bLower === 'uncategorized') return -1;
    return aLower.localeCompare(bLower);
  });

  return sortedCategories;
}

/**
 * Deduplicate mods by modId only, with fallback to name+author
 * When multiple fileIds exist for the same modId, keeps the first occurrence
 */
export function deduplicateMods(mods: ViewerMod[]): ViewerMod[] {
  const seen = new Set<string>();
  const unique: ViewerMod[] = [];

  mods.forEach((mod) => {
    // Primary key: modId only (one instance per mod, regardless of fileId)
    let key: string;
    
    if (mod.modId) {
      // modId present - use it as the deduplication key
      // This ensures one instance per mod, even if multiple fileIds exist
      key = String(mod.modId);
    } else {
      // No modId - fall back to name + author (less reliable but better than nothing)
      const name = (mod.name || '').toLowerCase().trim();
      const author = ((mod.uploader?.name || mod.author || '')).toLowerCase().trim();
      key = `name:${name}|author:${author}`;
    }
    
    if (!seen.has(key)) {
      seen.add(key);
      unique.push(mod);
    }
  });

  return unique;
}

/**
 * Get unique category names from mods, sorted alphabetically
 * with 'Uncategorized' at the end
 */
export function getUniqueCategories(mods: ViewerMod[]): string[] {
  const categories = new Set<string>();
  mods.forEach((mod) => {
    categories.add(mod.category || 'Uncategorized');
  });

  return Array.from(categories).sort((a, b) => {
    const aLower = a.toLowerCase();
    const bLower = b.toLowerCase();
    if (aLower === 'uncategorized') return 1;
    if (bLower === 'uncategorized') return -1;
    return aLower.localeCompare(bLower);
  });
}
