import { describe, expect, it, vi } from 'vitest';
import { triggerMarketCrawl } from '../src/tools/crawl.js';

const config = {
  apiBaseUrl: 'https://api.example.com',
  bearerToken: 'token',
  githubRepo: 'wbits/guitars',
};

describe('triggerMarketCrawl', () => {
  it('dispatches workflow without guitar id', async () => {
    const execFileImpl = vi.fn().mockResolvedValue({ stdout: 'ok', stderr: '' });
    await triggerMarketCrawl(config, {}, execFileImpl);

    expect(execFileImpl).toHaveBeenCalledWith(
      'gh',
      ['workflow', 'run', 'crawl.yml', '--repo', 'wbits/guitars'],
      expect.any(Object),
    );
  });

  it('passes guitar_id input when provided', async () => {
    const execFileImpl = vi.fn().mockResolvedValue({ stdout: 'ok', stderr: '' });
    await triggerMarketCrawl(config, { guitarId: 'abc-123' }, execFileImpl);

    expect(execFileImpl).toHaveBeenCalledWith(
      'gh',
      ['workflow', 'run', 'crawl.yml', '--repo', 'wbits/guitars', '-f', 'guitar_id=abc-123'],
      expect.any(Object),
    );
  });
});
