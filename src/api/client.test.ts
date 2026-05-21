import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const ENV_KEY = 'VITE_GUITARS_BEARER_TOKEN';

beforeEach(() => {
  vi.stubEnv(ENV_KEY, 'test-token');
  vi.resetModules();
});

afterEach(() => {
  vi.unstubAllEnvs();
  vi.restoreAllMocks();
});

describe('apiFetch', () => {
  it('adds Authorization, Accept and Content-Type headers', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ ok: true }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    );
    vi.stubGlobal('fetch', fetchMock);

    const { apiFetch } = await import('./client');
    const result = await apiFetch<{ ok: boolean }>({
      method: 'POST',
      path: '/guitar',
      body: { brand: 'Fender' },
    });

    expect(result).toEqual({ ok: true });
    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [, init] = fetchMock.mock.calls[0];
    expect(init.method).toBe('POST');
    expect(init.headers).toMatchObject({
      Authorization: 'Bearer test-token',
      Accept: 'application/json',
      'Content-Type': 'application/json',
    });
    expect(init.body).toBe(JSON.stringify({ brand: 'Fender' }));
  });

  it('omits Content-Type when no body is sent', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response('[]', { status: 200, headers: { 'Content-Type': 'application/json' } }),
    );
    vi.stubGlobal('fetch', fetchMock);

    const { apiFetch } = await import('./client');
    await apiFetch({ path: '/guitar' });

    const [, init] = fetchMock.mock.calls[0];
    expect(init.headers).not.toHaveProperty('Content-Type');
    expect(init.body).toBeUndefined();
  });

  it('returns undefined for 204 responses without touching the body', async () => {
    const noBody = new Response(null, { status: 204 });
    const textSpy = vi.spyOn(noBody, 'text');
    const fetchMock = vi.fn().mockResolvedValue(noBody);
    vi.stubGlobal('fetch', fetchMock);

    const { apiFetch } = await import('./client');
    const result = await apiFetch({ method: 'DELETE', path: '/guitar/abc' });

    expect(result).toBeUndefined();
    expect(textSpy).not.toHaveBeenCalled();
  });

  it('throws ApiError with parsed { error } body on non-2xx', async () => {
    const makeResponse = () =>
      new Response(JSON.stringify({ error: 'invalid brand' }), {
        status: 400,
        headers: { 'Content-Type': 'application/json' },
      });
    const fetchMock = vi.fn().mockImplementation(async () => makeResponse());
    vi.stubGlobal('fetch', fetchMock);

    const { apiFetch, ApiError } = await import('./client');

    await expect(apiFetch({ method: 'POST', path: '/guitar', body: {} })).rejects.toMatchObject({
      name: 'ApiError',
      status: 400,
      message: 'invalid brand',
      body: { error: 'invalid brand' },
    });

    let caught: unknown;
    try {
      await apiFetch({ method: 'POST', path: '/guitar', body: {} });
    } catch (e) {
      caught = e;
    }
    expect(caught).toBeInstanceOf(ApiError);
  });

  it('throws MissingTokenError when no token is configured', async () => {
    vi.stubEnv(ENV_KEY, '');
    sessionStorage.clear();
    const fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);

    const { apiFetch, MissingTokenError } = await import('./client');

    await expect(apiFetch({ path: '/guitar' })).rejects.toBeInstanceOf(MissingTokenError);
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
