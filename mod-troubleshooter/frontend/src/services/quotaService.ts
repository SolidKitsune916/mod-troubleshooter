import { z } from 'zod';
import { fetchApi } from './api.ts';

/**
 * Schema for quota response from the backend.
 */
export const QuotaResponseSchema = z.object({
  hourlyLimit: z.number(),
  hourlyRemaining: z.number(),
  dailyLimit: z.number(),
  dailyRemaining: z.number(),
  available: z.boolean(),
});

export type QuotaResponse = z.infer<typeof QuotaResponseSchema>;

/**
 * Fetches the current Nexus API quota information.
 * @returns The current quota information
 */
export async function fetchQuota(): Promise<QuotaResponse> {
  return fetchApi('/quota', QuotaResponseSchema);
}
