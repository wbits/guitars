import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js';
import { ApiError, type ApiClient } from '../api-client.js';
import {
  createGuitarToolArgsSchema,
  getGuitarToolArgsSchema,
  parseGuitar,
  parseGuitarList,
  toolFieldsToGuitarInput,
  updateGuitarToolArgsSchema,
} from '../schemas.js';

const jsonResult = (data: unknown) => ({
  content: [{ type: 'text' as const, text: JSON.stringify(data, null, 2) }],
});

const errorResult = (error: unknown) => {
  if (error instanceof ApiError) {
    return {
      content: [
        {
          type: 'text' as const,
          text: JSON.stringify(
            {
              error: error.message,
              status: error.status,
              body: error.body,
            },
            null,
            2,
          ),
        },
      ],
      isError: true as const,
    };
  }

  const message = error instanceof Error ? error.message : String(error);
  return {
    content: [{ type: 'text' as const, text: JSON.stringify({ error: message }, null, 2) }],
    isError: true as const,
  };
};

export const registerGuitarTools = (server: McpServer, client: ApiClient): void => {
  server.registerTool(
    'list_guitars',
    {
      description: 'List all guitars in the signed-in user collection',
      inputSchema: {},
    },
    async () => {
      try {
        const raw = await client.apiFetch<unknown>({ path: '/guitar' });
        return jsonResult(parseGuitarList(raw));
      } catch (error) {
        return errorResult(error);
      }
    },
  );

  server.registerTool(
    'get_guitar',
    {
      description: 'Get a single guitar by id',
      inputSchema: {
        id: getGuitarToolArgsSchema.shape.id,
      },
    },
    async (args) => {
      try {
        const { id } = getGuitarToolArgsSchema.parse(args);
        const raw = await client.apiFetch<unknown>({ path: `/guitar/${encodeURIComponent(id)}` });
        return jsonResult(parseGuitar(raw));
      } catch (error) {
        return errorResult(error);
      }
    },
  );

  server.registerTool(
    'create_guitar',
    {
      description:
        'Add a new guitar to the collection. Price is in major units (e.g. 1999.00). Full replace semantics apply on update.',
      inputSchema: {
        brand: createGuitarToolArgsSchema.shape.brand,
        typeName: createGuitarToolArgsSchema.shape.typeName,
        buildYear: createGuitarToolArgsSchema.shape.buildYear,
        price: createGuitarToolArgsSchema.shape.price,
        priceCurrency: createGuitarToolArgsSchema.shape.priceCurrency,
        pictures: createGuitarToolArgsSchema.shape.pictures,
        coverPictureIndex: createGuitarToolArgsSchema.shape.coverPictureIndex,
        serialNumber: createGuitarToolArgsSchema.shape.serialNumber,
        color: createGuitarToolArgsSchema.shape.color,
        country: createGuitarToolArgsSchema.shape.country,
        factory: createGuitarToolArgsSchema.shape.factory,
        description: createGuitarToolArgsSchema.shape.description,
      },
    },
    async (args) => {
      try {
        const fields = createGuitarToolArgsSchema.parse(args);
        const body = toolFieldsToGuitarInput(fields);
        const raw = await client.apiFetch<unknown>({ method: 'POST', path: '/guitar', body });
        return jsonResult(parseGuitar(raw));
      } catch (error) {
        return errorResult(error);
      }
    },
  );

  server.registerTool(
    'update_guitar',
    {
      description:
        'Replace an existing guitar (PUT). Provide all fields. Price is in major units (e.g. 1999.00).',
      inputSchema: {
        id: updateGuitarToolArgsSchema.shape.id,
        brand: updateGuitarToolArgsSchema.shape.brand,
        typeName: updateGuitarToolArgsSchema.shape.typeName,
        buildYear: updateGuitarToolArgsSchema.shape.buildYear,
        price: updateGuitarToolArgsSchema.shape.price,
        priceCurrency: updateGuitarToolArgsSchema.shape.priceCurrency,
        pictures: updateGuitarToolArgsSchema.shape.pictures,
        coverPictureIndex: updateGuitarToolArgsSchema.shape.coverPictureIndex,
        serialNumber: updateGuitarToolArgsSchema.shape.serialNumber,
        color: updateGuitarToolArgsSchema.shape.color,
        country: updateGuitarToolArgsSchema.shape.country,
        factory: updateGuitarToolArgsSchema.shape.factory,
        description: updateGuitarToolArgsSchema.shape.description,
      },
    },
    async (args) => {
      try {
        const { id, ...fields } = updateGuitarToolArgsSchema.parse(args);
        const body = toolFieldsToGuitarInput(fields);
        const raw = await client.apiFetch<unknown>({
          method: 'PUT',
          path: `/guitar/${encodeURIComponent(id)}`,
          body,
        });
        return jsonResult(parseGuitar(raw));
      } catch (error) {
        return errorResult(error);
      }
    },
  );
};
