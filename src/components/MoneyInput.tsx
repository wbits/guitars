import { forwardRef, useEffect, useState } from 'react';
import { majorToMinor, minorToMajor } from '@/lib/money';

interface MoneyInputProps {
  id?: string;
  /** Value in MINOR units (cents). */
  value: number | undefined;
  /** Receives a value in MINOR units (cents), or undefined when blank. */
  onChange: (minor: number | undefined) => void;
  onBlur?: () => void;
  disabled?: boolean;
  placeholder?: string;
  'aria-invalid'?: boolean;
  'aria-describedby'?: string;
}

const formatMajor = (minor: number | undefined): string => {
  if (minor === undefined || minor === null || Number.isNaN(minor)) return '';
  return minorToMajor(minor).toFixed(2);
};

/**
 * Displays decimal "major" amounts (e.g. 19.99) but emits integer minor units
 * (e.g. 1999) to the parent. Accepts either `.` or `,` as a decimal separator.
 */
export const MoneyInput = forwardRef<HTMLInputElement, MoneyInputProps>(
  ({ id, value, onChange, onBlur, disabled, placeholder, ...aria }, ref) => {
    const [text, setText] = useState<string>(() => formatMajor(value));

    useEffect(() => {
      // Sync from outside only when the underlying value diverges from what
      // the user is typing (e.g. when the form is reset).
      const currentMinor = parseToMinor(text);
      if (currentMinor !== value) {
        setText(formatMajor(value));
      }
      // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [value]);

    return (
      <input
        ref={ref}
        id={id}
        type="text"
        inputMode="decimal"
        className="input"
        placeholder={placeholder ?? '0.00'}
        value={text}
        disabled={disabled}
        onChange={(e) => {
          const next = e.target.value;
          setText(next);
          onChange(parseToMinor(next));
        }}
        onBlur={() => {
          setText(formatMajor(value));
          onBlur?.();
        }}
        aria-invalid={aria['aria-invalid']}
        aria-describedby={aria['aria-describedby']}
      />
    );
  },
);

MoneyInput.displayName = 'MoneyInput';

const parseToMinor = (text: string): number | undefined => {
  const trimmed = text.trim().replace(/,/g, '.');
  if (trimmed === '') return undefined;
  const num = Number(trimmed);
  if (!Number.isFinite(num)) return undefined;
  return majorToMinor(num);
};
