import { describe, expect, it } from 'vitest';
import { guitarInputSchema } from './guitar';

const validPayload = () => ({
  brand: 'Fender',
  typeName: 'Stratocaster',
  buildYear: 1996,
  priceAmount: 199900,
  priceCurrency: 'EUR' as const,
  pictures: ['https://example.com/strat.jpg'],
  serialNumber: 'SN-1234',
  description: 'A fine guitar.',
});

describe('guitarInputSchema', () => {
  it('accepts a valid payload', () => {
    const result = guitarInputSchema.safeParse(validPayload());
    expect(result.success).toBe(true);
  });

  it('rejects blank brand', () => {
    const result = guitarInputSchema.safeParse({ ...validPayload(), brand: '   ' });
    expect(result.success).toBe(false);
  });

  it('rejects blank typeName', () => {
    const result = guitarInputSchema.safeParse({ ...validPayload(), typeName: '' });
    expect(result.success).toBe(false);
  });

  it('rejects buildYear below 1800', () => {
    const result = guitarInputSchema.safeParse({ ...validPayload(), buildYear: 1799 });
    expect(result.success).toBe(false);
  });

  it('rejects buildYear above currentYear + 1', () => {
    const future = new Date().getFullYear() + 2;
    const result = guitarInputSchema.safeParse({ ...validPayload(), buildYear: future });
    expect(result.success).toBe(false);
  });

  it('rejects non-integer buildYear', () => {
    const result = guitarInputSchema.safeParse({ ...validPayload(), buildYear: 1996.5 });
    expect(result.success).toBe(false);
  });

  it('rejects negative priceAmount', () => {
    const result = guitarInputSchema.safeParse({ ...validPayload(), priceAmount: -1 });
    expect(result.success).toBe(false);
  });

  it('rejects non-integer priceAmount', () => {
    const result = guitarInputSchema.safeParse({ ...validPayload(), priceAmount: 199.5 });
    expect(result.success).toBe(false);
  });

  it('rejects unsupported currencies', () => {
    const result = guitarInputSchema.safeParse({ ...validPayload(), priceCurrency: 'GBP' });
    expect(result.success).toBe(false);
  });

  it('rejects non-URL pictures', () => {
    const result = guitarInputSchema.safeParse({
      ...validPayload(),
      pictures: ['not a url'],
    });
    expect(result.success).toBe(false);
  });

  it('rejects non-absolute (non http(s)) URLs', () => {
    const result = guitarInputSchema.safeParse({
      ...validPayload(),
      pictures: ['ftp://example.com/x.jpg'],
    });
    expect(result.success).toBe(false);
  });

  it('accepts empty pictures array', () => {
    const result = guitarInputSchema.safeParse({ ...validPayload(), pictures: [] });
    expect(result.success).toBe(true);
  });

  it('defaults pictures to [] when omitted', () => {
    const { pictures: _omit, ...rest } = validPayload();
    void _omit;
    const result = guitarInputSchema.safeParse(rest);
    expect(result.success).toBe(true);
    if (result.success) expect(result.data.pictures).toEqual([]);
  });

  it('treats blank optional fields as undefined', () => {
    const result = guitarInputSchema.safeParse({
      ...validPayload(),
      serialNumber: '   ',
      description: '',
    });
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.serialNumber).toBeUndefined();
      expect(result.data.description).toBeUndefined();
    }
  });
});
