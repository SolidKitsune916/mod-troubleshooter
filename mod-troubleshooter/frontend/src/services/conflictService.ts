import { ConflictAnalyzeResponseSchema } from '@/types/index.ts';
import type { ConflictAnalyzeResponse } from '@/types/index.ts';

import { fetchApi } from './api.ts';

/** Analyze conflicts for a collection revision */
export async function analyzeCollectionConflicts(
  slug: string,
  revision: number,
  includeHashes = false,
): Promise<ConflictAnalyzeResponse> {
  const queryParams = includeHashes ? '?includeHashes=true' : '';
  return fetchApi(
    `/collections/${encodeURIComponent(slug)}/revisions/${revision}/conflicts${queryParams}`,
    ConflictAnalyzeResponseSchema,
  );
}

/** Mod reference for manual conflict analysis */
interface ModReference {
  modId: string;
  modName: string;
  game: string;
  nexusModId: number;
  fileId: number;
}

/** Request payload for manual conflict analysis */
interface AnalyzeConflictsRequest {
  mods: ModReference[];
  includeContentHashes?: boolean;
}

/** Analyze conflicts for a custom list of mods */
export async function analyzeConflicts(
  mods: ModReference[],
  includeContentHashes = false,
): Promise<ConflictAnalyzeResponse> {
  const body: AnalyzeConflictsRequest = { mods, includeContentHashes };

  return fetchApi('/conflicts/analyze', ConflictAnalyzeResponseSchema, {
    method: 'POST',
    body: JSON.stringify(body),
  });
}
