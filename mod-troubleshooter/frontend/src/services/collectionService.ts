import { CollectionSchema, RevisionDetailsSchema } from '@/types/index.ts';
import type { Collection, RevisionDetails } from '@/types/index.ts';

import { fetchApi } from './api.ts';

/** Fetch collection with latest revision mods */
export async function fetchCollection(slug: string): Promise<Collection> {
  return fetchApi(`/collections/${encodeURIComponent(slug)}`, CollectionSchema);
}

/** Fetch collection revision history */
export async function fetchCollectionRevisions(
  slug: string,
): Promise<Collection> {
  return fetchApi(
    `/collections/${encodeURIComponent(slug)}/revisions`,
    CollectionSchema,
  );
}

/** Fetch specific collection revision with mods */
export async function fetchCollectionRevisionMods(
  slug: string,
  revisionNumber: number,
): Promise<RevisionDetails> {
  return fetchApi(
    `/collections/${encodeURIComponent(slug)}/revisions/${revisionNumber}`,
    RevisionDetailsSchema,
  );
}
