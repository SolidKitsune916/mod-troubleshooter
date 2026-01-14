import {
  SettingsSchema,
  ValidateKeyResponseSchema,
} from '@/types/settings.ts';
import type { Settings, ValidateKeyResponse } from '@/types/settings.ts';

import { fetchApi, fetchApiMessage } from './api.ts';

/** Fetch current settings */
export async function fetchSettings(): Promise<Settings> {
  return fetchApi('/settings', SettingsSchema);
}

/** Update settings with new API key */
export async function updateSettings(nexusApiKey: string): Promise<string> {
  return fetchApiMessage('/settings', {
    method: 'POST',
    body: JSON.stringify({ nexusApiKey }),
  });
}

/** Validate an API key without saving */
export async function validateApiKey(
  nexusApiKey: string,
): Promise<ValidateKeyResponse> {
  return fetchApi('/settings/validate', ValidateKeyResponseSchema, {
    method: 'POST',
    body: JSON.stringify({ nexusApiKey }),
  });
}
