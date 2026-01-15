import { useQuery } from '@tanstack/react-query';
import { fetchQuota, type QuotaResponse } from '@services/quotaService.ts';

/**
 * Query key for quota data.
 */
export const quotaKeys = {
  all: ['quota'] as const,
};

/**
 * Hook to fetch and manage Nexus API quota information.
 * Polls every 30 seconds when the window is focused.
 */
export function useQuota() {
  return useQuery<QuotaResponse, Error>({
    queryKey: quotaKeys.all,
    queryFn: fetchQuota,
    staleTime: 30 * 1000, // 30 seconds
    refetchInterval: 60 * 1000, // Refetch every minute
    refetchOnWindowFocus: true,
  });
}

/**
 * Calculate quota percentage remaining.
 */
export function getQuotaPercentage(quota: QuotaResponse, type: 'hourly' | 'daily'): number {
  if (!quota.available) return 0;

  const limit = type === 'hourly' ? quota.hourlyLimit : quota.dailyLimit;
  const remaining = type === 'hourly' ? quota.hourlyRemaining : quota.dailyRemaining;

  if (limit === 0) return 0;
  return Math.round((remaining / limit) * 100);
}

/**
 * Get quota status level based on remaining percentage.
 */
export function getQuotaStatus(percentage: number): 'good' | 'warning' | 'critical' {
  if (percentage >= 50) return 'good';
  if (percentage >= 20) return 'warning';
  return 'critical';
}
