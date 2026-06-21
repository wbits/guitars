import { execFile } from 'node:child_process';
import { promisify } from 'node:util';
import { z } from 'zod';
import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js';
import type { McpConfig } from '../config.js';

const execFileAsync = promisify(execFile);

export const crawlToolArgsSchema = z.object({
  guitarId: z.string().optional().describe('Optional single guitar id to crawl'),
});

export type ExecFileFn = typeof execFileAsync;

export const triggerMarketCrawl = async (
  config: McpConfig,
  args: z.infer<typeof crawlToolArgsSchema>,
  execFileImpl: ExecFileFn = execFileAsync,
): Promise<{ message: string }> => {
  const ghArgs = ['workflow', 'run', 'crawl.yml', '--repo', config.githubRepo];
  if (args.guitarId) {
    ghArgs.push('-f', `guitar_id=${args.guitarId}`);
  }

  const { stdout, stderr } = await execFileImpl('gh', ghArgs, {
    env: process.env,
  });

  const output = [stdout, stderr].filter(Boolean).join('\n').trim();
  return {
    message: output.length > 0 ? output : `Triggered market crawl for ${config.githubRepo}`,
  };
};

export const registerCrawlTool = (
  server: McpServer,
  config: McpConfig,
  execFileImpl?: ExecFileFn,
): void => {
  server.registerTool(
    'trigger_market_crawl',
    {
      description:
        'Dispatch the market crawl GitHub Actions workflow (wbits/guitars). Requires gh CLI authenticated (gh auth login or GH_TOKEN). Collection must have marketCrawlEnabled.',
      inputSchema: {
        guitarId: crawlToolArgsSchema.shape.guitarId,
      },
    },
    async (args) => {
      try {
        const parsed = crawlToolArgsSchema.parse(args);
        const result = await triggerMarketCrawl(config, parsed, execFileImpl);
        return {
          content: [{ type: 'text' as const, text: JSON.stringify(result, null, 2) }],
        };
      } catch (error) {
        const message =
          error instanceof Error
            ? error.message.includes('ENOENT')
              ? 'gh CLI not found on PATH. Install GitHub CLI and authenticate, or set GH_TOKEN.'
              : error.message
            : String(error);
        return {
          content: [{ type: 'text' as const, text: JSON.stringify({ error: message }, null, 2) }],
          isError: true as const,
        };
      }
    },
  );
};
