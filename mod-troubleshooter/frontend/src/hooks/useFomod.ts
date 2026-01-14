import { useQuery } from '@tanstack/react-query';

import { analyzeFomod } from '@services/fomodService.ts';

/** Query keys for FOMOD operations */
export const fomodKeys = {
  all: ['fomod'] as const,
  analyze: (game: string, modId: number, fileId: number) =>
    ['fomod', 'analyze', game, modId, fileId] as const,
};

/** Hook to analyze a FOMOD from a mod file */
export function useFomodAnalysis(
  game: string,
  modId: number,
  fileId: number,
  enabled = true,
) {
  return useQuery({
    queryKey: fomodKeys.analyze(game, modId, fileId),
    queryFn: () => analyzeFomod(game, modId, fileId),
    enabled: enabled && Boolean(game) && modId > 0 && fileId > 0,
    staleTime: 24 * 60 * 60 * 1000, // 24 hours - FOMOD data is static
  });
}
