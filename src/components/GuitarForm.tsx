import { useState } from 'react';
import { Controller, useFieldArray, useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import {
  CURRENCIES,
  guitarInputSchema,
  type GuitarInput,
} from '@/domain/guitar';
import { MoneyInput } from './MoneyInput';

export interface GuitarFormProps {
  initialValues?: Partial<GuitarInput>;
  onSubmit: (values: GuitarInput) => void | Promise<void>;
  onCancel?: () => void;
  submitting?: boolean;
  submitLabel?: string;
  /** Errors from the server, displayed underneath the form. */
  serverError?: string | null;
}

/**
 * react-hook-form's useFieldArray requires object entries, so we wrap each
 * picture URL in `{ url }` for the form state and then unwrap on submit via
 * a zod `.transform()`. The resolver still enforces the domain rules.
 */
const formSchema = guitarInputSchema
  .omit({ pictures: true })
  .extend({
    pictures: z.array(z.object({ url: z.string() })).default([]),
  })
  .transform((value) => ({
    ...value,
    pictures: value.pictures.map((p) => p.url),
  }))
  .pipe(guitarInputSchema);

type FormShape = {
  brand: string;
  typeName: string;
  buildYear: number;
  priceAmount: number | undefined;
  priceCurrency: 'EUR' | 'USD';
  pictures: { url: string }[];
  serialNumber?: string;
  description?: string;
};

const toFormShape = (v: Partial<GuitarInput> | undefined): FormShape => ({
  brand: v?.brand ?? '',
  typeName: v?.typeName ?? '',
  buildYear: v?.buildYear ?? new Date().getFullYear(),
  priceAmount: v?.priceAmount,
  priceCurrency: v?.priceCurrency ?? 'EUR',
  pictures: (v?.pictures ?? []).map((url) => ({ url })),
  serialNumber: v?.serialNumber ?? '',
  description: v?.description ?? '',
});

export const GuitarForm = ({
  initialValues,
  onSubmit,
  onCancel,
  submitting,
  submitLabel = 'Save',
  serverError,
}: GuitarFormProps) => {
  const [newPicture, setNewPicture] = useState('');

  const {
    register,
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<FormShape, unknown, GuitarInput>({
    resolver: zodResolver(formSchema) as never,
    defaultValues: toFormShape(initialValues),
    mode: 'onBlur',
  });

  const { fields, append, remove } = useFieldArray({ control, name: 'pictures' });

  const addPictureUrl = () => {
    const trimmed = newPicture.trim();
    if (trimmed === '') return;
    append({ url: trimmed });
    setNewPicture('');
  };

  const submit = handleSubmit(async (values) => {
    await onSubmit(values);
  });

  return (
    <form onSubmit={submit} className="space-y-5" noValidate>
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div>
          <label htmlFor="brand" className="label">
            Brand
          </label>
          <input
            id="brand"
            className="input"
            aria-invalid={Boolean(errors.brand)}
            {...register('brand')}
          />
          {errors.brand ? <p className="error-text">{errors.brand.message}</p> : null}
        </div>

        <div>
          <label htmlFor="typeName" className="label">
            Type
          </label>
          <input
            id="typeName"
            className="input"
            aria-invalid={Boolean(errors.typeName)}
            {...register('typeName')}
          />
          {errors.typeName ? <p className="error-text">{errors.typeName.message}</p> : null}
        </div>

        <div>
          <label htmlFor="buildYear" className="label">
            Build year
          </label>
          <input
            id="buildYear"
            type="number"
            className="input"
            aria-invalid={Boolean(errors.buildYear)}
            {...register('buildYear', { valueAsNumber: true })}
          />
          {errors.buildYear ? <p className="error-text">{errors.buildYear.message}</p> : null}
        </div>

        <div>
          <label htmlFor="serialNumber" className="label">
            Serial number
          </label>
          <input
            id="serialNumber"
            className="input"
            aria-invalid={Boolean(errors.serialNumber)}
            {...register('serialNumber')}
          />
          <p className="help">Optional.</p>
        </div>

        <div>
          <label htmlFor="priceAmount" className="label">
            Price
          </label>
          <Controller
            control={control}
            name="priceAmount"
            render={({ field }) => (
              <MoneyInput
                id="priceAmount"
                value={field.value}
                onChange={field.onChange}
                onBlur={field.onBlur}
                aria-invalid={Boolean(errors.priceAmount)}
              />
            )}
          />
          <p className="help">Enter the price in major units (e.g. 1999.00).</p>
          {errors.priceAmount ? (
            <p className="error-text">{errors.priceAmount.message}</p>
          ) : null}
        </div>

        <div>
          <label htmlFor="priceCurrency" className="label">
            Currency
          </label>
          <select
            id="priceCurrency"
            className="input"
            aria-invalid={Boolean(errors.priceCurrency)}
            {...register('priceCurrency')}
          >
            {CURRENCIES.map((c) => (
              <option key={c} value={c}>
                {c}
              </option>
            ))}
          </select>
          {errors.priceCurrency ? (
            <p className="error-text">{errors.priceCurrency.message}</p>
          ) : null}
        </div>
      </div>

      <div>
        <label htmlFor="description" className="label">
          Description
        </label>
        <textarea
          id="description"
          rows={3}
          className="input"
          {...register('description')}
        />
      </div>

      <fieldset>
        <legend className="label">Picture URLs</legend>
        {fields.length === 0 ? (
          <p className="help">No pictures added yet.</p>
        ) : (
          <ul className="space-y-2">
            {fields.map((field, index) => (
              <li key={field.id} className="flex items-center gap-2">
                <input
                  className="input"
                  aria-label={`Picture URL ${index + 1}`}
                  aria-invalid={Boolean(errors.pictures?.[index])}
                  {...register(`pictures.${index}.url` as const)}
                />
                <button
                  type="button"
                  onClick={() => remove(index)}
                  className="btn-secondary"
                >
                  Remove
                </button>
              </li>
            ))}
          </ul>
        )}
        {errors.pictures && !Array.isArray(errors.pictures) ? (
          <p className="error-text">{(errors.pictures as { message?: string }).message}</p>
        ) : null}
        <div className="mt-3 flex items-center gap-2">
          <input
            type="url"
            placeholder="https://example.com/image.jpg"
            className="input"
            value={newPicture}
            onChange={(e) => setNewPicture(e.target.value)}
            aria-label="Add picture URL"
          />
          <button type="button" onClick={addPictureUrl} className="btn-secondary">
            Add picture URL
          </button>
        </div>
      </fieldset>

      {serverError ? (
        <div
          role="alert"
          className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800"
        >
          <p className="font-semibold">The server rejected the request:</p>
          <p>{serverError}</p>
        </div>
      ) : null}

      <div className="flex items-center justify-end gap-2">
        {onCancel ? (
          <button type="button" onClick={onCancel} className="btn-secondary">
            Cancel
          </button>
        ) : null}
        <button type="submit" className="btn-primary" disabled={submitting}>
          {submitting ? 'Saving…' : submitLabel}
        </button>
      </div>
    </form>
  );
};
