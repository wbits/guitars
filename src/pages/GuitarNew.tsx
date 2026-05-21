import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { GuitarForm } from '@/components/GuitarForm';
import { useCreateGuitar } from '@/api/guitars';
import { ApiError } from '@/api/client';

export const GuitarNew = () => {
  const navigate = useNavigate();
  const create = useCreateGuitar();
  const [serverError, setServerError] = useState<string | null>(null);

  return (
    <section className="space-y-4">
      <header>
        <h1 className="text-2xl font-semibold">Add guitar</h1>
        <p className="text-sm text-slate-600">
          Create a new guitar in the collection.
        </p>
      </header>

      <div className="rounded-md border border-slate-200 bg-white p-6 shadow-sm">
        <GuitarForm
          submitting={create.isPending}
          submitLabel="Create"
          serverError={serverError}
          onCancel={() => navigate(-1)}
          onSubmit={async (values) => {
            setServerError(null);
            try {
              const created = await create.mutateAsync(values);
              navigate(`/guitars/${created.id}`);
            } catch (err) {
              if (err instanceof ApiError) {
                setServerError(err.message);
              } else if (err instanceof Error) {
                setServerError(err.message);
              } else {
                setServerError('Unknown error');
              }
            }
          }}
        />
      </div>
    </section>
  );
};
