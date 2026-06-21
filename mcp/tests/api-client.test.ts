import { describe, expect, it, vi } from 'vitest';
import { ApiError, createApiClient } from '../src/api-client.js';

const config = {
  apiBaseUrl: 'https://api.example.com',
  bearerToken: 'test-token',
  githubRepo: 'wbits/guitars',
};

describe('createApiClient', () => {
  it('sends bearer token and parses JSON success responses', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ id: 'g1' }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    );

    const { apiFetch } = createApiClient(config, fetchMock);
    const result = await apiFetch<{ id: string }>({ path: '/guitar/g1' });

    expect(result).toEqual({ id: 'g1' });
    expect(fetchMock).toHaveBeenCalledWith(
      'https://api.example.com/guitar/g1',
      expect.objectContaining({
        method: 'GET',
        headers: expect.objectContaining({
          Authorization: 'Bearer test-token',
        }),
      }),
    );
  });

  it('throws ApiError with API error body message', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ error: 'not found' }), {
        status: 404,
        headers: { 'Content-Type': 'application/json' },
      }),
    );

    const { apiFetch } = createApiClient(config, fetchMock);

    await expect(apiFetch({ path: '/guitar/missing' })).rejects.toMatchObject({
      name: 'ApiError',
      status: 404,
      message: 'not found',
    } satisfies Partial<ApiError>);
  });

  it('handles 204 No Content', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 204 }));
    const { apiFetch } = createApiClient(config, fetchMock);
    await expect(apiFetch({ method: 'DELETE', path: '/guitar/x' })).resolves.toBeUndefined();
  });
});
