import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const ENV_KEY = 'VITE_GUITARS_BEARER_TOKEN';

const reload = async () => {
  vi.resetModules();
  return import('./token');
};

describe('token resolution', () => {
  beforeEach(() => {
    sessionStorage.clear();
    vi.unstubAllEnvs();
  });

  afterEach(() => {
    sessionStorage.clear();
    vi.unstubAllEnvs();
  });

  it('returns null when nothing is configured', async () => {
    vi.stubEnv(ENV_KEY, '');
    const { getToken, hasToken } = await reload();
    expect(getToken()).toBeNull();
    expect(hasToken()).toBe(false);
  });

  it('falls back to the build-time env var', async () => {
    vi.stubEnv(ENV_KEY, 'build-time-token');
    const { getToken, hasToken } = await reload();
    expect(getToken()).toBe('build-time-token');
    expect(hasToken()).toBe(true);
  });

  it('runtime sessionStorage override wins over the build-time token', async () => {
    vi.stubEnv(ENV_KEY, 'build-time-token');
    const { getToken, setRuntimeToken, RUNTIME_TOKEN_KEY } = await reload();
    setRuntimeToken('runtime-token');
    expect(sessionStorage.getItem(RUNTIME_TOKEN_KEY)).toBe('runtime-token');
    expect(getToken()).toBe('runtime-token');
  });

  it('clearRuntimeToken removes the override and falls back', async () => {
    vi.stubEnv(ENV_KEY, 'build-time-token');
    const { getToken, setRuntimeToken, clearRuntimeToken } = await reload();
    setRuntimeToken('runtime-token');
    expect(getToken()).toBe('runtime-token');
    clearRuntimeToken();
    expect(getToken()).toBe('build-time-token');
  });

  it('ignores whitespace-only runtime tokens', async () => {
    vi.stubEnv(ENV_KEY, '');
    const { getToken, setRuntimeToken } = await reload();
    setRuntimeToken('   ');
    expect(getToken()).toBeNull();
  });
});
