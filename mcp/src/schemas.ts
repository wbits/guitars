import { z } from 'zod';
import {
  CURRENCIES,
  guitarInputSchema,
  guitarListSchema,
  guitarSchema,
  type Guitar,
  type GuitarInput,
} from './contracts/guitar.js';
import { majorToMinor } from './money.js';

const optionalFields = {
  pictures: z.array(z.string().url()).optional().describe('Picture URLs (http/https)'),
  coverPictureIndex: z.number().int().min(0).optional().describe('Index into pictures for thumbnail'),
  serialNumber: z.string().optional(),
  color: z.string().optional(),
  country: z.string().optional(),
  factory: z.string().optional(),
  description: z.string().optional(),
};

export const guitarToolFieldsSchema = z.object({
  brand: z.string().describe('Guitar brand, e.g. Fender'),
  typeName: z.string().describe('Model or type name, e.g. Stratocaster'),
  buildYear: z.number().int().describe('Year the guitar was built'),
  price: z.number().nonnegative().describe('Price in major units (e.g. 1999.00 EUR)'),
  priceCurrency: z.enum(CURRENCIES).describe('EUR or USD'),
  ...optionalFields,
});

export const createGuitarToolArgsSchema = guitarToolFieldsSchema;

export const updateGuitarToolArgsSchema = guitarToolFieldsSchema.extend({
  id: z.string().min(1).describe('Guitar id to update'),
});

export const getGuitarToolArgsSchema = z.object({
  id: z.string().min(1).describe('Guitar id'),
});

export type GuitarToolFields = z.infer<typeof guitarToolFieldsSchema>;

export const toolFieldsToGuitarInput = (fields: GuitarToolFields): GuitarInput => {
  const { price, ...rest } = fields;
  return guitarInputSchema.parse({
    ...rest,
    pictures: rest.pictures ?? [],
    coverPictureIndex: rest.coverPictureIndex ?? 0,
    priceAmount: majorToMinor(price),
    priceCurrency: rest.priceCurrency,
  });
};

export const parseGuitar = (raw: unknown): Guitar => guitarSchema.parse(raw);

export const parseGuitarList = (raw: unknown): Guitar[] => guitarListSchema.parse(raw);

export { CURRENCIES, type Guitar, type GuitarInput };
