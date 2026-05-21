import { z } from 'zod';

export const CURRENCIES = ['EUR', 'USD'] as const;
export type Currency = (typeof CURRENCIES)[number];

const MIN_BUILD_YEAR = 1800;
const maxBuildYear = (): number => new Date().getFullYear() + 1;

const trimmedNonEmpty = (field: string) =>
  z
    .string({ required_error: `${field} is required` })
    .transform((s) => s.trim())
    .refine((s) => s.length > 0, { message: `${field} is required` });

const optionalTrimmed = z
  .string()
  .optional()
  .transform((s) => (s === undefined ? undefined : s.trim()))
  .transform((s) => (s === '' ? undefined : s));

const absoluteUrl = z
  .string()
  .trim()
  .url({ message: 'Must be a valid absolute URL' })
  .refine((u) => /^https?:\/\//i.test(u), {
    message: 'Picture URLs must start with http(s)://',
  });

/**
 * Schema for what the client sends to the API (POST/PUT body).
 * `id` is server-generated and intentionally not part of this shape.
 */
export const guitarInputSchema = z.object({
  brand: trimmedNonEmpty('Brand'),
  typeName: trimmedNonEmpty('Type'),
  buildYear: z
    .number({ invalid_type_error: 'Build year must be a number' })
    .int('Build year must be an integer')
    .min(MIN_BUILD_YEAR, `Build year must be >= ${MIN_BUILD_YEAR}`)
    .refine((y) => y <= maxBuildYear(), {
      message: 'Build year cannot be in the future',
    }),
  priceAmount: z
    .number({ invalid_type_error: 'Price must be a number' })
    .int('Price must be an integer in minor units')
    .min(0, 'Price must be >= 0'),
  priceCurrency: z.enum(CURRENCIES, {
    errorMap: () => ({ message: 'Currency must be EUR or USD' }),
  }),
  pictures: z.array(absoluteUrl).default([]),
  serialNumber: optionalTrimmed,
  description: optionalTrimmed,
});

export type GuitarInput = z.infer<typeof guitarInputSchema>;

export const guitarSchema = guitarInputSchema.extend({
  id: z.string().min(1),
});

export type Guitar = z.infer<typeof guitarSchema>;

export const guitarListSchema = z.array(guitarSchema);
