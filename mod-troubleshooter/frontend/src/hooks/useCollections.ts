import { useQuery } from '@tanstack/react-query';

import {
  fetchCollection,
  fetchCollectionRevisions,
  fetchCollectionRevisionMods,
} from '@services/collectionService.ts';

/** Query keys for collections */
export const collectionKeys = {
  all: ['collections'] as const,
  detail: (slug: string) => ['collections', slug] as const,
  revisions: (slug: string) => ['collections', slug, 'revisions'] as const,
  revisionMods: (slug: string, revision: number) =>
    ['collections', slug, 'revisions', revision] as const,
};

/** Hook to fetch a collection with latest revision mods */
export function useCollection(slug: string, enabled = true) {
  return useQuery({
    queryKey: collectionKeys.detail(slug),
    queryFn: () => fetchCollection(slug),
    enabled: enabled && Boolean(slug),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

/** Hook to fetch collection revision history */
export function useCollectionRevisions(slug: string, enabled = true) {
  return useQuery({
    queryKey: collectionKeys.revisions(slug),
    queryFn: () => fetchCollectionRevisions(slug),
    enabled: enabled && Boolean(slug),
    staleTime: 5 * 60 * 1000,
  });
}

/** Hook to fetch specific revision mods */
export function useCollectionRevisionMods(
  slug: string,
  revisionNumber: number,
  enabled = true,
) {
  return useQuery({
    queryKey: collectionKeys.revisionMods(slug, revisionNumber),
    queryFn: () => fetchCollectionRevisionMods(slug, revisionNumber),
    enabled: enabled && Boolean(slug) && revisionNumber > 0,
    staleTime: 5 * 60 * 1000,
  });
}
