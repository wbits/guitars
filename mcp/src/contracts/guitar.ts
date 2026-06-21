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

export const guitarFieldsSchema = z.object({
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
  coverPictureIndex: z.number().int().min(0).default(0),
  serialNumber: optionalTrimmed,
  color: optionalTrimmed,
  country: optionalTrimmed,
  factory: optionalTrimmed,
  description: optionalTrimmed,
});

const refineCoverPicture = (
  data: { pictures: string[]; coverPictureIndex: number },
  ctx: z.RefinementCtx,
) => {
  if (data.pictures.length > 0 && data.coverPictureIndex >= data.pictures.length) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: 'Cover picture must refer to an uploaded picture',
      path: ['coverPictureIndex'],
    });
  }
};

/** POST/PUT body shape (no id or owner). Keep in sync with Go domain validation. */
export const guitarInputSchema = guitarFieldsSchema.superRefine(refineCoverPicture);

export type GuitarInput = z.infer<typeof guitarInputSchema>;

export const guitarSchema = guitarFieldsSchema
  .extend({
    id: z.string().min(1),
    owner: z.string().min(1).optional(),
  })
  .superRefine(refineCoverPicture);

export type Guitar = z.infer<typeof guitarSchema>;

export const guitarListSchema = z.array(guitarSchema);
