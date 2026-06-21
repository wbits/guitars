import { describe, expect, it } from 'vitest';
import { loadConfig } from '../src/config.js';

describe('loadConfig', () => {
  it('parses required env vars and strips trailing slash from API URL', () => {
    const config = loadConfig({
      GUITARS_API_BASE_URL: 'https://api.example.com/',
      GUITARS_BEARER_TOKEN: 'token-123',
    });

    expect(config).toEqual({
      apiBaseUrl: 'https://api.example.com',
      bearerToken: 'token-123',
      githubRepo: 'wbits/guitars',
    });
  });

  it('uses GITHUB_REPO when set', () => {
    const config = loadConfig({
      GUITARS_API_BASE_URL: 'https://api.example.com',
      GUITARS_BEARER_TOKEN: 'token',
      GITHUB_REPO: 'org/repo',
    });

    expect(config.githubRepo).toBe('org/repo');
  });

  it('throws when required env vars are missing', () => {
    expect(() => loadConfig({})).toThrow(/GUITARS_API_BASE_URL/);
    expect(() =>
      loadConfig({
        GUITARS_API_BASE_URL: 'https://api.example.com',
      }),
    ).toThrow(/GUITARS_BEARER_TOKEN/);
  });
});
