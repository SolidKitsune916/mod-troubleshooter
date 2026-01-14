import { useQuery } from '@tanstack/react-query';

import { analyzeCollectionConflicts } from '@services/conflictService.ts';

/** Query keys for conflict operations */
export const conflictKeys = {
  all: ['conflicts'] as const,
  analyze: (slug: string, revision: number, includeHashes: boolean) =>
    ['conflicts', 'analyze', slug, revision, includeHashes] as const,
};

/** Hook to analyze conflicts for a collection revision */
export function useConflictAnalysis(
  slug: string,
  revision: number,
  includeHashes = false,
  enabled = true,
) {
  return useQuery({
    queryKey: conflictKeys.analyze(slug, revision, includeHashes),
    queryFn: () => analyzeCollectionConflicts(slug, revision, includeHashes),
    enabled: enabled && Boolean(slug) && revision > 0,
    staleTime: 24 * 60 * 60 * 1000, // 24 hours - conflict analysis is stable
  });
}
