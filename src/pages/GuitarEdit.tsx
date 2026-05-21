import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { GuitarForm } from '@/components/GuitarForm';
import { useGuitar, useUpdateGuitar } from '@/api/guitars';
import { ApiError } from '@/api/client';
import { ErrorBanner } from '@/components/ErrorBanner';

export const GuitarEdit = () => {
  const { id = '' } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const guitar = useGuitar(id);
  const update = useUpdateGuitar(id);
  const [serverError, setServerError] = useState<string | null>(null);

  if (guitar.isLoading) return <p className="text-sm text-slate-600">Loading…</p>;
  if (guitar.isError || !guitar.data) {
    return <ErrorBanner error={guitar.error ?? 'Guitar not found'} title="Could not load guitar" />;
  }

  return (
    <section className="space-y-4">
      <header>
        <h1 className="text-2xl font-semibold">Edit guitar</h1>
        <p className="text-sm text-slate-600">
          Update the details for <span className="font-medium">{guitar.data.brand}</span>{' '}
          <span className="text-slate-500">({guitar.data.typeName})</span>.
        </p>
      </header>

      <div className="rounded-md border border-slate-200 bg-white p-6 shadow-sm">
        <GuitarForm
          initialValues={guitar.data}
          submitting={update.isPending}
          submitLabel="Save changes"
          serverError={serverError}
          onCancel={() => navigate(`/guitars/${id}`)}
          onSubmit={async (values) => {
            setServerError(null);
            try {
              const updated = await update.mutateAsync(values);
              navigate(`/guitars/${updated.id}`);
            } catch (err) {
              if (err instanceof ApiError) setServerError(err.message);
              else if (err instanceof Error) setServerError(err.message);
              else setServerError('Unknown error');
            }
          }}
        />
      </div>
    </section>
  );
};
