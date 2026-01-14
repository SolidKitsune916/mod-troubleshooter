import { useQuery } from '@tanstack/react-query';

import { analyzeCollectionLoadOrder } from '@services/loadorderService.ts';

/** Query keys for load order operations */
export const loadOrderKeys = {
  all: ['loadorder'] as const,
  analyze: (slug: string, revision: number) =>
    ['loadorder', 'analyze', slug, revision] as const,
};

/** Hook to analyze load order for a collection revision */
export function useLoadOrderAnalysis(
  slug: string,
  revision: number,
  enabled = true,
) {
  return useQuery({
    queryKey: loadOrderKeys.analyze(slug, revision),
    queryFn: () => analyzeCollectionLoadOrder(slug, revision),
    enabled: enabled && Boolean(slug) && revision > 0,
    staleTime: 24 * 60 * 60 * 1000, // 24 hours - load order analysis is stable
  });
}
