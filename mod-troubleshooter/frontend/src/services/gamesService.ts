import { SupportedGamesSchema } from '@/types/api.ts';
import type { SupportedGame } from '@/types/api.ts';

import { fetchApi } from './api.ts';

/**
 * Fetch the list of supported games from the backend.
 * @returns Array of supported games with their IDs, labels, and Nexus domain names
 */
export async function fetchGames(): Promise<SupportedGame[]> {
  return fetchApi('/games', SupportedGamesSchema);
}
