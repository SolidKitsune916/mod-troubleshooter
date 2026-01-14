import { FomodAnalyzeResponseSchema } from '@/types/index.ts';
import type { FomodAnalyzeResponse } from '@/types/index.ts';

import { fetchApi } from './api.ts';

/** Request payload for FOMOD analysis */
interface AnalyzeFomodRequest {
  game: string;
  modId: number;
  fileId: number;
}

/** Analyze a FOMOD from a mod file */
export async function analyzeFomod(
  game: string,
  modId: number,
  fileId: number,
): Promise<FomodAnalyzeResponse> {
  const body: AnalyzeFomodRequest = { game, modId, fileId };

  return fetchApi('/fomod/analyze', FomodAnalyzeResponseSchema, {
    method: 'POST',
    body: JSON.stringify(body),
  });
}
