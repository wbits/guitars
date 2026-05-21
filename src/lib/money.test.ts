import { describe, expect, it } from 'vitest';
import { formatMoney, majorToMinor, minorToMajor } from './money';

describe('money', () => {
  describe('majorToMinor / minorToMajor', () => {
    it('round-trips integer values', () => {
      for (const value of [0, 1, 12, 199, 1999, 100000]) {
        expect(minorToMajor(majorToMinor(value))).toBe(value);
      }
    });

    it('round-trips two-decimal values', () => {
      const cases: Array<[number, number]> = [
        [19.99, 1999],
        [0.01, 1],
        [1234.56, 123456],
      ];
      for (const [major, expectedMinor] of cases) {
        expect(majorToMinor(major)).toBe(expectedMinor);
        expect(minorToMajor(expectedMinor)).toBe(major);
      }
    });

    it('rejects non-integer minor units', () => {
      expect(() => minorToMajor(1.5)).toThrow();
    });

    it('rejects non-finite major units', () => {
      expect(() => majorToMinor(Number.NaN)).toThrow();
      expect(() => majorToMinor(Number.POSITIVE_INFINITY)).toThrow();
    });
  });

  describe('formatMoney', () => {
    it('formats EUR with comma decimals and dot thousands', () => {
      // "EUR 1.999,00" - allow non-breaking-space variations from Intl
      const result = formatMoney(199900, 'EUR');
      expect(result.startsWith('EUR ')).toBe(true);
      expect(result).toContain('1.999');
      expect(result).toContain(',00');
    });

    it('formats USD with comma thousands and dot decimals', () => {
      const result = formatMoney(199900, 'USD');
      expect(result.startsWith('USD ')).toBe(true);
      expect(result).toContain('1,999');
      expect(result).toContain('.00');
    });

    it('always renders exactly two fraction digits', () => {
      expect(formatMoney(0, 'USD')).toBe('USD 0.00');
      expect(formatMoney(50, 'USD')).toBe('USD 0.50');
    });
  });
});
