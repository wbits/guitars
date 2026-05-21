import { useMemo } from 'react';
import { Link } from 'react-router-dom';
import { useGuitars } from '@/api/guitars';
import { ErrorBanner } from '@/components/ErrorBanner';
import { formatMoney } from '@/lib/money';

export const GuitarList = () => {
  const { data, isLoading, isError, error, refetch, isFetching } = useGuitars();

  const sorted = useMemo(
    () =>
      [...(data ?? [])].sort((a, b) =>
        a.brand.localeCompare(b.brand, undefined, { sensitivity: 'base' }),
      ),
    [data],
  );

  return (
    <section className="space-y-4">
      <header className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Guitars</h1>
          <p className="text-sm text-slate-600">
            {isFetching ? 'Refreshing…' : `${sorted.length} guitar${sorted.length === 1 ? '' : 's'}`}
          </p>
        </div>
        <Link to="/guitars/new" className="btn-primary">
          Add guitar
        </Link>
      </header>

      {isLoading ? <p className="text-sm text-slate-600">Loading guitars…</p> : null}

      {isError ? (
        <div className="space-y-2">
          <ErrorBanner error={error} title="Could not load guitars" />
          <button type="button" onClick={() => refetch()} className="btn-secondary">
            Retry
          </button>
        </div>
      ) : null}

      {!isLoading && !isError && sorted.length === 0 ? (
        <div className="rounded-md border border-dashed border-slate-300 bg-white p-8 text-center">
          <p className="text-slate-700">No guitars in the collection yet.</p>
          <p className="mt-2 text-sm text-slate-500">
            Get started by adding your first one.
          </p>
          <Link to="/guitars/new" className="btn-primary mt-4 inline-flex">
            Add guitar
          </Link>
        </div>
      ) : null}

      {sorted.length > 0 ? (
        <div className="overflow-x-auto rounded-md border border-slate-200 bg-white">
          <table className="min-w-full divide-y divide-slate-200 text-sm">
            <thead className="bg-slate-50 text-left text-xs uppercase tracking-wide text-slate-500">
              <tr>
                <th className="px-4 py-2">Brand</th>
                <th className="px-4 py-2">Type</th>
                <th className="px-4 py-2">Build year</th>
                <th className="px-4 py-2">Price</th>
                <th className="px-4 py-2">Serial number</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {sorted.map((g) => (
                <tr key={g.id} className="hover:bg-slate-50">
                  <td className="px-4 py-2">
                    <Link to={`/guitars/${g.id}`} className="font-medium text-slate-900 hover:underline">
                      {g.brand}
                    </Link>
                  </td>
                  <td className="px-4 py-2">{g.typeName}</td>
                  <td className="px-4 py-2 tabular-nums">{g.buildYear}</td>
                  <td className="px-4 py-2 tabular-nums">
                    {formatMoney(g.priceAmount, g.priceCurrency)}
                  </td>
                  <td className="px-4 py-2 text-slate-600">{g.serialNumber ?? '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : null}
    </section>
  );
};
