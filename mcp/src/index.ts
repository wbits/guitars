import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { loadConfig } from './config.js';
import { createGuitarsMcpServer } from './server.js';

const log = (message: string): void => {
  console.error(`[guitars-mcp] ${message}`);
};

const main = async (): Promise<void> => {
  const config = loadConfig();
  const server = createGuitarsMcpServer(config);
  const transport = new StdioServerTransport();
  await server.connect(transport);
  log('stdio server connected');
};

main().catch((error: unknown) => {
  const message = error instanceof Error ? error.message : String(error);
  log(`fatal: ${message}`);
  process.exit(1);
});
