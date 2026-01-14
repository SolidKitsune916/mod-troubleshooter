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

/** API response envelope schema - matches backend Response struct */
const ResponseEnvelopeSchema = <T extends z.ZodTypeAny>(dataSchema: T) =>
  z.object({
    data: dataSchema.optional(),
    error: z.string().optional(),
    message: z.string().optional(),
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
    // Try to parse error response
    try {
      const errorJson = await response.json();
      const errorEnvelope = z.object({ error: z.string().optional() }).parse(errorJson);
      throw new ApiError(response.status, errorEnvelope.error ?? 'Unknown error');
    } catch (e) {
      if (e instanceof ApiError) throw e;
      const errorText = await response.text().catch(() => 'Unknown error');
      throw new ApiError(response.status, `API error: ${errorText}`);
    }
  }

  const json: unknown = await response.json();
  const envelope = ResponseEnvelopeSchema(schema).parse(json);

  if (envelope.error) {
    throw new ApiError(500, envelope.error);
  }

  if (envelope.data === undefined) {
    throw new ApiError(500, 'No data in response');
  }

  return envelope.data;
}

/** Fetch wrapper for message-only responses (success messages) */
async function fetchApiMessage(
  endpoint: string,
  options?: RequestInit,
): Promise<string> {
  const url = `${API_BASE_URL}${endpoint}`;

  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!response.ok) {
    try {
      const errorJson = await response.json();
      const errorEnvelope = z.object({ error: z.string().optional() }).parse(errorJson);
      throw new ApiError(response.status, errorEnvelope.error ?? 'Unknown error');
    } catch (e) {
      if (e instanceof ApiError) throw e;
      const errorText = await response.text().catch(() => 'Unknown error');
      throw new ApiError(response.status, `API error: ${errorText}`);
    }
  }

  const json: unknown = await response.json();
  const envelope = z.object({
    message: z.string().optional(),
    error: z.string().optional(),
  }).parse(json);

  if (envelope.error) {
    throw new ApiError(500, envelope.error);
  }

  return envelope.message ?? 'Success';
}

/** Export the fetch wrappers for use in service modules */
export { fetchApi, fetchApiMessage };
