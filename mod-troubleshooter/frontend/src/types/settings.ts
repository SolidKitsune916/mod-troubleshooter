import { z } from 'zod';

/** Settings response schema from API */
export const SettingsSchema = z.object({
  nexusApiKey: z.string(),
  hasNexusKey: z.boolean(),
  keyConfigured: z.boolean(),
});

/** Update settings request schema */
export const UpdateSettingsSchema = z.object({
  nexusApiKey: z.string(),
});

/** Validate API key response schema */
export const ValidateKeyResponseSchema = z.object({
  valid: z.boolean(),
});

/** Type exports inferred from schemas */
export type Settings = z.infer<typeof SettingsSchema>;
export type UpdateSettingsRequest = z.infer<typeof UpdateSettingsSchema>;
export type ValidateKeyResponse = z.infer<typeof ValidateKeyResponseSchema>;
