import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

import {
  fetchSettings,
  updateSettings,
  validateApiKey,
} from '@services/settingsService.ts';

/** Query keys for settings */
export const settingsKeys = {
  all: ['settings'] as const,
  detail: () => ['settings', 'detail'] as const,
};

/** Hook to fetch current settings */
export function useSettings() {
  return useQuery({
    queryKey: settingsKeys.detail(),
    queryFn: fetchSettings,
    staleTime: 30 * 1000, // 30 seconds - settings don't change often
  });
}

/** Hook to update settings */
export function useUpdateSettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (nexusApiKey: string) => updateSettings(nexusApiKey),
    onSuccess: () => {
      // Invalidate settings cache after update
      queryClient.invalidateQueries({ queryKey: settingsKeys.all });
    },
  });
}

/** Hook to validate an API key */
export function useValidateApiKey() {
  return useMutation({
    mutationFn: (nexusApiKey: string) => validateApiKey(nexusApiKey),
  });
}
