import type { McpConfig } from './config.js';

export class ApiError extends Error {
  public readonly status: number;
  public readonly body: unknown;

  constructor(status: number, message: string, body: unknown) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.body = body;
  }
}

export interface ApiRequest {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  path: string;
  body?: unknown;
}

const parseErrorMessage = (status: number, body: unknown): string => {
  if (
    body !== null &&
    typeof body === 'object' &&
    'error' in body &&
    typeof (body as { error: unknown }).error === 'string'
  ) {
    return (body as { error: string }).error;
  }
  return `Request failed with status ${status}`;
};

export type FetchFn = typeof fetch;

export const createApiClient = (config: McpConfig, fetchImpl: FetchFn = fetch) => {
  const buildUrl = (path: string): string => {
    const normalised = path.startsWith('/') ? path : `/${path}`;
    return `${config.apiBaseUrl}${normalised}`;
  };

  const apiFetch = async <T = unknown>({
    method = 'GET',
    path,
    body,
  }: ApiRequest): Promise<T> => {
    const headers: Record<string, string> = {
      Accept: 'application/json',
      Authorization: `Bearer ${config.bearerToken}`,
    };
    if (body !== undefined) {
      headers['Content-Type'] = 'application/json';
    }

    const response = await fetchImpl(buildUrl(path), {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
    });

    if (response.status === 204) {
      return undefined as T;
    }

    const text = await response.text();
    let parsed: unknown = undefined;
    if (text.length > 0) {
      try {
        parsed = JSON.parse(text);
      } catch {
        parsed = text;
      }
    }

    if (!response.ok) {
      throw new ApiError(response.status, parseErrorMessage(response.status, parsed), parsed);
    }

    return parsed as T;
  };

  return { apiFetch };
};

export type ApiClient = ReturnType<typeof createApiClient>;
