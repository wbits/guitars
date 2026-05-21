/**
 * The API stores prices as integer minor units (e.g. cents). The UI displays
 * and accepts decimal "major" units. These helpers convert between the two
 * without floating-point surprises.
 */

const MINOR_PER_MAJOR = 100;

export const majorToMinor = (major: number): number => {
  if (!Number.isFinite(major)) {
    throw new Error('majorToMinor: value must be a finite number');
  }
  // Use Math.round on a rescaled value to avoid 1.005 -> 100 style drift.
  return Math.round(major * MINOR_PER_MAJOR);
};

export const minorToMajor = (minor: number): number => {
  if (!Number.isInteger(minor)) {
    throw new Error('minorToMajor: value must be an integer');
  }
  return minor / MINOR_PER_MAJOR;
};

const localeFor = (currency: string): string => {
  switch (currency) {
    case 'EUR':
      return 'de-DE'; // "1.999,00 €" - we strip the symbol below
    case 'USD':
      return 'en-US';
    default:
      return 'en-US';
  }
};

/**
 * Formats a minor-unit amount as "<CURRENCY> <amount>", e.g. "EUR 1.999,00".
 * The currency code is rendered before the value (rather than using the
 * locale's native symbol position) so prices stay unambiguous in the UI.
 */
export const formatMoney = (minor: number, currency: string): string => {
  const major = minorToMajor(minor);
  const formatted = new Intl.NumberFormat(localeFor(currency), {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(major);
  return `${currency} ${formatted}`;
};
