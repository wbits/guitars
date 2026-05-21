import { getToken } from '@/lib/token';

const API_BASE_URL: string = (() => {
  const raw = import.meta.env.VITE_GUITARS_API_BASE_URL;
  if (typeof raw !== 'string') return '';
  return raw.replace(/\/+$/, '');
})();

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

export class MissingTokenError extends Error {
  constructor() {
    super('No bearer token configured');
    this.name = 'MissingTokenError';
  }
}

export interface ApiRequest {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE';
  path: string;
  body?: unknown;
  signal?: AbortSignal;
}

const buildUrl = (path: string): string => {
  const normalised = path.startsWith('/') ? path : `/${path}`;
  return `${API_BASE_URL}${normalised}`;
};

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

export const apiFetch = async <T = unknown>({
  method = 'GET',
  path,
  body,
  signal,
}: ApiRequest): Promise<T> => {
  const token = getToken();
  if (!token) {
    throw new MissingTokenError();
  }

  const headers: Record<string, string> = {
    Accept: 'application/json',
    Authorization: `Bearer ${token}`,
  };
  if (body !== undefined) {
    headers['Content-Type'] = 'application/json';
  }

  const response = await fetch(buildUrl(path), {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
    signal,
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

export const apiBaseUrl = (): string => API_BASE_URL;
