import { useEffect, useState, type ReactNode } from 'react';
import { Link } from 'react-router-dom';
import { hasToken, RUNTIME_TOKEN_KEY } from '@/lib/token';

interface AuthGateProps {
  children: ReactNode;
}

/**
 * Renders children only when a bearer token is configured. Otherwise displays
 * a friendly explanation and a link to /settings. Re-checks on storage events
 * so pasting a token in /settings unblocks the rest of the app immediately.
 */
export const AuthGate = ({ children }: AuthGateProps) => {
  const [tokenPresent, setTokenPresent] = useState<boolean>(() => hasToken());

  useEffect(() => {
    const recheck = () => setTokenPresent(hasToken());

    window.addEventListener('storage', recheck);
    window.addEventListener('guitars:token-changed', recheck);
    return () => {
      window.removeEventListener('storage', recheck);
      window.removeEventListener('guitars:token-changed', recheck);
    };
  }, []);

  if (tokenPresent) return <>{children}</>;

  return (
    <div className="mx-auto max-w-xl p-8">
      <div className="rounded-lg border border-amber-300 bg-amber-50 p-6">
        <h2 className="text-lg font-semibold text-amber-900">
          No bearer token configured
        </h2>
        <p className="mt-2 text-sm text-amber-900">
          The GuitarCollection API requires a bearer token on every request. This
          build does not have one baked in, and the runtime override
          (<code className="rounded bg-amber-100 px-1">sessionStorage</code> key{' '}
          <code className="rounded bg-amber-100 px-1">{RUNTIME_TOKEN_KEY}</code>) is
          empty.
        </p>
        <p className="mt-4">
          <Link to="/settings" className="btn-primary">
            Configure token
          </Link>
        </p>
      </div>
    </div>
  );
};
