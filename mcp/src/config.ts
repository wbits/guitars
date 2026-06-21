import { z } from 'zod';

const configSchema = z.object({
  apiBaseUrl: z
    .string({ required_error: 'GUITARS_API_BASE_URL is required' })
    .min(1, 'GUITARS_API_BASE_URL is required')
    .transform((url) => url.replace(/\/+$/, '')),
  bearerToken: z
    .string({ required_error: 'GUITARS_BEARER_TOKEN is required' })
    .min(1, 'GUITARS_BEARER_TOKEN is required'),
  githubRepo: z.string().min(1).default('wbits/guitars'),
});

export type McpConfig = z.infer<typeof configSchema>;

export const loadConfig = (env: NodeJS.ProcessEnv = process.env): McpConfig => {
  const parsed = configSchema.safeParse({
    apiBaseUrl: env.GUITARS_API_BASE_URL,
    bearerToken: env.GUITARS_BEARER_TOKEN,
    githubRepo: env.GITHUB_REPO ?? 'wbits/guitars',
  });

  if (!parsed.success) {
    const message = parsed.error.issues.map((issue) => issue.message).join('; ');
    throw new Error(`Invalid MCP configuration: ${message}`);
  }

  return parsed.data;
};
