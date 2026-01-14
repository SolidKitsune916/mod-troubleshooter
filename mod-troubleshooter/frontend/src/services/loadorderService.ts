import { LoadOrderAnalyzeResponseSchema } from '@/types/index.ts';
import type { LoadOrderAnalyzeResponse } from '@/types/index.ts';

import { fetchApi } from './api.ts';

/** Analyze load order for a collection revision */
export async function analyzeCollectionLoadOrder(
  slug: string,
  revision: number,
): Promise<LoadOrderAnalyzeResponse> {
  return fetchApi(
    `/collections/${encodeURIComponent(slug)}/revisions/${revision}/loadorder`,
    LoadOrderAnalyzeResponseSchema,
  );
}

/** Request payload for manual load order analysis */
interface AnalyzeLoadOrderRequest {
  plugins: PluginReference[];
}

/** Plugin reference for manual analysis */
interface PluginReference {
  filename: string;
  game?: string;
  modId?: number;
  fileId?: number;
}

/** Analyze a custom list of plugins */
export async function analyzeLoadOrder(
  plugins: PluginReference[],
): Promise<LoadOrderAnalyzeResponse> {
  const body: AnalyzeLoadOrderRequest = { plugins };

  return fetchApi('/loadorder/analyze', LoadOrderAnalyzeResponseSchema, {
    method: 'POST',
    body: JSON.stringify(body),
  });
}
