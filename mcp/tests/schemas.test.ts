import { describe, expect, it } from 'vitest';
import { majorToMinor } from '../src/money.js';
import { toolFieldsToGuitarInput } from '../src/schemas.js';

describe('toolFieldsToGuitarInput', () => {
  it('converts major-unit price to minor units for the API', () => {
    const input = toolFieldsToGuitarInput({
      brand: 'Fender',
      typeName: 'Stratocaster',
      buildYear: 1996,
      price: 1999,
      priceCurrency: 'EUR',
    });

    expect(input.priceAmount).toBe(majorToMinor(1999));
    expect(input.brand).toBe('Fender');
    expect(input.pictures).toEqual([]);
    expect(input.coverPictureIndex).toBe(0);
  });

  it('rejects invalid build year', () => {
    expect(() =>
      toolFieldsToGuitarInput({
        brand: 'Fender',
        typeName: 'Strat',
        buildYear: 1700,
        price: 100,
        priceCurrency: 'EUR',
      }),
    ).toThrow();
  });
});
