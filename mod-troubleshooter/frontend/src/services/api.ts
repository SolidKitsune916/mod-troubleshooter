import { z } from 'zod';

/** Custom error class for API errors */
export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

/** API response envelope schema */
const ResponseEnvelopeSchema = <T extends z.ZodTypeAny>(dataSchema: T) =>
  z.object({
    success: z.boolean(),
    data: dataSchema.optional(),
    error: z.string().optional(),
  });

/** Base API configuration */
const API_BASE_URL = '/api';

/** Generic fetch wrapper with type safety and error handling */
async function fetchApi<T>(
  endpoint: string,
  schema: z.ZodType<T>,
  options?: RequestInit,
): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;

  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!response.ok) {
    const errorText = await response.text().catch(() => 'Unknown error');
    throw new ApiError(response.status, `API error: ${errorText}`);
  }

  const json: unknown = await response.json();
  const envelope = ResponseEnvelopeSchema(schema).parse(json);

  if (!envelope.success || envelope.error) {
    throw new ApiError(500, envelope.error ?? 'Unknown error');
  }

  if (envelope.data === undefined) {
    throw new ApiError(500, 'No data in response');
  }

  return envelope.data;
}

/** Export the fetch wrapper for use in service modules */
export { fetchApi };
