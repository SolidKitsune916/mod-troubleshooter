import { useQuery } from '@tanstack/react-query';

import { fetchGames } from '@services/gamesService.ts';
import type { SupportedGame } from '@/types/api.ts';

/**
 * Query key for games data.
 */
export const gamesKeys = {
  all: ['games'] as const,
};

/**
 * Hook to fetch and manage the list of supported games.
 * Games are cached for 24 hours since they rarely change.
 */
export function useGames() {
  return useQuery<SupportedGame[], Error>({
    queryKey: gamesKeys.all,
    queryFn: fetchGames,
    staleTime: 24 * 60 * 60 * 1000, // 24 hours
    gcTime: 24 * 60 * 60 * 1000, // Keep in cache for 24 hours
  });
}

/**
 * Get a game by its ID from the list of supported games.
 */
export function getGameById(games: SupportedGame[], id: string): SupportedGame | undefined {
  return games.find((game) => game.id === id);
}

/**
 * Get the default game ID (first game in the list).
 */
export function getDefaultGameId(games: SupportedGame[]): string {
  return games[0]?.id ?? 'skyrim';
}
