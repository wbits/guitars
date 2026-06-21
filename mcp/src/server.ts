import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js';
import { createApiClient } from './api-client.js';
import type { McpConfig } from './config.js';
import { registerCrawlTool } from './tools/crawl.js';
import { registerGuitarTools } from './tools/guitars.js';

export const createGuitarsMcpServer = (config: McpConfig): McpServer => {
  const server = new McpServer({
    name: 'guitars-mcp',
    version: '0.1.0',
  });

  const client = createApiClient(config);
  registerGuitarTools(server, client);
  registerCrawlTool(server, config);

  return server;
};
