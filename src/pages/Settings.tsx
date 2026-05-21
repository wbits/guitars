import { useState } from 'react';
import {
  clearRuntimeToken,
  getRuntimeToken,
  RUNTIME_TOKEN_KEY,
  setRuntimeToken,
} from '@/lib/token';

const notifyTokenChanged = () => {
  window.dispatchEvent(new Event('guitars:token-changed'));
};

const mask = (token: string): string => {
  if (token.length <= 8) return '••••';
  return `${token.slice(0, 4)}…${token.slice(-4)}`;
};

export const Settings = () => {
  const [current, setCurrent] = useState<string | null>(() => getRuntimeToken());
  const [draft, setDraft] = useState('');
  const [saved, setSaved] = useState(false);
  const buildTimeFallback =
    typeof import.meta.env.VITE_GUITARS_BEARER_TOKEN === 'string' &&
    import.meta.env.VITE_GUITARS_BEARER_TOKEN.trim() !== '';

  const save = (e: React.FormEvent) => {
    e.preventDefault();
    setRuntimeToken(draft);
    setCurrent(getRuntimeToken());
    setDraft('');
    setSaved(true);
    notifyTokenChanged();
    setTimeout(() => setSaved(false), 2000);
  };

  const clear = () => {
    clearRuntimeToken();
    setCurrent(null);
    notifyTokenChanged();
  };

  return (
    <section className="space-y-6">
      <header>
        <h1 className="text-2xl font-semibold">Settings</h1>
        <p className="text-sm text-slate-600">
          Configure the bearer token used to call the GuitarCollection API.
        </p>
      </header>

      <div className="rounded-md border border-amber-300 bg-amber-50 p-4 text-sm text-amber-900">
        <p className="font-semibold">Security caveat</p>
        <p className="mt-1">
          Tokens entered here are stored in this tab's{' '}
          <code className="rounded bg-amber-100 px-1">sessionStorage</code> (key{' '}
          <code className="rounded bg-amber-100 px-1">{RUNTIME_TOKEN_KEY}</code>) and
          take precedence over any token baked into the build at build time.
          Anything baked into the build is publicly visible in the static bundle.
        </p>
      </div>

      <div className="rounded-md border border-slate-200 bg-white p-6 shadow-sm">
        <h2 className="text-base font-semibold text-slate-900">Current token</h2>
        <p className="mt-1 text-sm text-slate-600">
          {current
            ? `Runtime token configured: ${mask(current)}`
            : buildTimeFallback
              ? 'Using the build-time token from VITE_GUITARS_BEARER_TOKEN.'
              : 'No token configured.'}
        </p>
        {current ? (
          <button type="button" onClick={clear} className="btn-secondary mt-3">
            Clear runtime token
          </button>
        ) : null}
      </div>

      <form onSubmit={save} className="rounded-md border border-slate-200 bg-white p-6 shadow-sm">
        <label htmlFor="token" className="label">
          New bearer token
        </label>
        <input
          id="token"
          type="password"
          autoComplete="off"
          className="input"
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          placeholder="paste token here"
        />
        <p className="help">
          Stored only in this tab's <code>sessionStorage</code>. Closing the tab
          discards it.
        </p>
        <div className="mt-3 flex items-center gap-3">
          <button type="submit" className="btn-primary" disabled={draft.trim() === ''}>
            Save token
          </button>
          {saved ? <span className="text-sm text-green-700">Saved.</span> : null}
        </div>
      </form>
    </section>
  );
};
