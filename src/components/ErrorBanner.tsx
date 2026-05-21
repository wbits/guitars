import { ApiError, MissingTokenError } from '@/api/client';
import { Link } from 'react-router-dom';

interface ErrorBannerProps {
  error: unknown;
  title?: string;
}

const describe = (error: unknown): string => {
  if (error instanceof MissingTokenError) {
    return 'No bearer token configured.';
  }
  if (error instanceof ApiError) {
    if (error.status === 401) return 'The configured bearer token was rejected.';
    if (error.status === 404) return 'The requested guitar could not be found.';
    return error.message;
  }
  if (error instanceof Error) return error.message;
  if (typeof error === 'string') return error;
  return 'An unexpected error occurred.';
};

export const ErrorBanner = ({ error, title }: ErrorBannerProps) => {
  if (!error) return null;
  const message = describe(error);
  const showTokenLink =
    error instanceof MissingTokenError ||
    (error instanceof ApiError && error.status === 401);

  return (
    <div
      role="alert"
      className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800"
    >
      {title ? <p className="mb-1 font-semibold">{title}</p> : null}
      <p>{message}</p>
      {showTokenLink ? (
        <p className="mt-2">
          <Link to="/settings" className="font-medium underline">
            Configure a token in settings
          </Link>
        </p>
      ) : null}
    </div>
  );
};
