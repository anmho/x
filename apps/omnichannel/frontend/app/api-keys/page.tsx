'use client';

import { FormEvent, useEffect, useMemo, useState } from 'react';
import { KeyRound, Plus } from 'lucide-react';
import { AppNav } from '@/app/_components/app-nav';
import { PageIntro } from '@/app/_components/blocks';
import { HeadlessSelect } from '@/app/_components/headless-select';
import { AccessKey, listAccessKeys, mintAccessKey, revokeAccessKey } from '@/lib/access-keys';

type Application = {
  id: string;
  name: string;
};

const APP_STORAGE_KEY = 'notifications:apps';
const DEFAULT_APPS: Application[] = [
  { id: 'app_notif_prod', name: 'notifications-api-prod' },
  { id: 'app_notif_stage', name: 'notifications-api-staging' },
];
const SCOPES = ['notifications:read', 'notifications:write', 'notifications:admin'];

function readStored<T>(key: string, fallback: T): T {
  try {
    const value = window.localStorage.getItem(key);
    if (!value) return fallback;
    return JSON.parse(value) as T;
  } catch {
    return fallback;
  }
}

function toAccessEnvironment(value: string): 'dev' | 'staging' | 'prod' {
  if (value === 'staging') return 'staging';
  if (value === 'prod') return 'prod';
  return 'dev';
}

function scopePairs(scopes: string[]) {
  return scopes.map((entry) => {
    const [service, scope] = entry.split(':');
    return { service, scope };
  });
}

function formatDate(value: string) {
  return new Date(value).toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  });
}

export default function ApiKeysPage() {
  const [apps, setApps] = useState<Application[]>(DEFAULT_APPS);
  const [keys, setKeys] = useState<AccessKey[]>([]);
  const [newOwner, setNewOwner] = useState('console');
  const [newKeyAppId, setNewKeyAppId] = useState(DEFAULT_APPS[0].id);
  const [newEnvironment, setNewEnvironment] = useState<'dev' | 'staging' | 'prod'>('dev');
  const [newScopes, setNewScopes] = useState<string[]>(['notifications:write']);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [createdKey, setCreatedKey] = useState<string | null>(null);

  useEffect(() => {
    setApps(readStored<Application[]>(APP_STORAGE_KEY, DEFAULT_APPS));
  }, []);

  async function refreshKeys() {
    try {
      setLoading(true);
      const data = await listAccessKeys();
      setKeys(data);
      setError(null);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to load keys';
      setError(message);
      setSuccess(null);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void refreshKeys();
  }, []);

  function toggleScope(scope: string) {
    setNewScopes((prev) => (prev.includes(scope) ? prev.filter((item) => item !== scope) : [...prev, scope]));
  }

  const selectedApplication = useMemo(
    () => apps.find((app) => app.id === newKeyAppId) ?? apps[0],
    [apps, newKeyAppId],
  );

  async function createKey(event: FormEvent) {
    event.preventDefault();
    if (!selectedApplication || !newOwner.trim() || newScopes.length === 0) return;

    try {
      setSubmitting(true);
      const minted = await mintAccessKey({
        application: selectedApplication.name,
        owner: newOwner.trim(),
        environment: toAccessEnvironment(newEnvironment),
        service_scopes: scopePairs(newScopes),
      });
      setKeys((prev) => [minted, ...prev]);
      setCreatedKey(minted.key || null);
      setSuccess(`Minted key ${minted.key_prefix}`);
      setError(null);
      setNewScopes(['notifications:write']);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to mint key';
      setError(message);
      setSuccess(null);
    } finally {
      setSubmitting(false);
    }
  }

  async function revokeKey(id: string) {
    try {
      const revoked = await revokeAccessKey(id);
      setKeys((prev) => prev.map((entry) => (entry.id === id ? revoked : entry)));
      setSuccess(`Revoked ${revoked.key_prefix}`);
      setError(null);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to revoke key';
      setError(message);
      setSuccess(null);
    }
  }

  return (
    <AppNav active="apiKeys">
      <main className="mx-auto max-w-6xl px-4 py-8 sm:px-6">
        <PageIntro
          title="API Key Management"
          description="Mint and revoke scoped keys for each application."
        />

        {error && <p className="mt-3 rounded-md border border-rose-500/30 bg-rose-500/10 px-3 py-2 text-sm text-rose-200">{error}</p>}
        {success && <p className="mt-3 rounded-md border border-emerald-500/30 bg-emerald-500/10 px-3 py-2 text-sm text-emerald-200">{success}</p>}
        {createdKey && (
          <p className="mt-3 rounded-md border border-amber-500/30 bg-amber-500/10 px-3 py-2 font-mono text-xs text-amber-200">
            {createdKey}
          </p>
        )}

        <section className="mt-6 rounded-xl border border-zinc-800 bg-zinc-950">
          <div className="border-b border-zinc-800 px-4 py-3">
            <h2 className="inline-flex items-center gap-2 text-sm font-medium text-zinc-200">
              <KeyRound className="h-4 w-4" /> Mint key
            </h2>
          </div>
          <form onSubmit={createKey} className="space-y-3 p-4">
            <input
              value={newOwner}
              onChange={(event) => setNewOwner(event.target.value)}
              placeholder="owner"
              className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
            />
            <HeadlessSelect
              value={newKeyAppId}
              onValueChange={setNewKeyAppId}
              options={apps.map((app) => ({ value: app.id, label: app.name }))}
              ariaLabel="Select application"
              placeholder="Select application"
            />
            <HeadlessSelect
              value={newEnvironment}
              onValueChange={(value) => setNewEnvironment(value as 'dev' | 'staging' | 'prod')}
              options={[
                { value: 'dev', label: 'dev' },
                { value: 'staging', label: 'staging' },
                { value: 'prod', label: 'prod' },
              ]}
              ariaLabel="Select environment"
            />
            <div className="flex flex-wrap gap-1.5">
              {SCOPES.map((scope) => (
                <label
                  key={scope}
                  className={`cursor-pointer rounded-full border px-2 py-1 text-xs ${
                    newScopes.includes(scope)
                      ? 'border-zinc-600 bg-zinc-800 text-zinc-100'
                      : 'border-zinc-700 bg-zinc-900 text-zinc-400'
                  }`}
                >
                  <input
                    type="checkbox"
                    className="hidden"
                    checked={newScopes.includes(scope)}
                    onChange={() => toggleScope(scope)}
                  />
                  {scope}
                </label>
              ))}
            </div>
            <button
              type="submit"
              disabled={submitting}
              className="inline-flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-50"
            >
              <Plus className="h-3.5 w-3.5" /> Mint key
            </button>
          </form>
        </section>

        <section className="mt-4 rounded-xl border border-zinc-800 bg-zinc-950">
          <div className="border-b border-zinc-800 px-4 py-3">
            <h2 className="text-sm font-medium text-zinc-200">Keys</h2>
          </div>
          {loading ? (
            <div className="px-4 py-3 text-sm text-zinc-400">Loading keys...</div>
          ) : (
            <div className="divide-y divide-zinc-800">
              {keys.map((entry) => (
                <article key={entry.id} className="px-4 py-3">
                  <div className="flex items-start justify-between gap-2">
                    <div>
                      <p className="text-sm font-medium">{entry.owner}</p>
                      <p className="text-xs text-zinc-500">{entry.application} • {entry.environment}</p>
                      <p className="mt-1 font-mono text-xs text-zinc-300">{entry.key_prefix}</p>
                      <p className="mt-1 text-xs text-zinc-500">{formatDate(entry.created_at)}</p>
                    </div>
                    <span className={`rounded-full px-2 py-0.5 text-xs ${entry.status === 'active' ? 'bg-emerald-500/20 text-emerald-300' : 'bg-rose-500/20 text-rose-300'}`}>
                      {entry.status}
                    </span>
                  </div>
                  <div className="mt-2 flex flex-wrap gap-1">
                    {entry.service_scopes.map((scope) => (
                      <span key={`${entry.id}:${scope.service}:${scope.scope}`} className="rounded-full border border-zinc-700 bg-zinc-900 px-2 py-0.5 text-xs text-zinc-300">
                        {scope.service}:{scope.scope}
                      </span>
                    ))}
                  </div>
                  {entry.status === 'active' && (
                    <button
                      type="button"
                      onClick={() => void revokeKey(entry.id)}
                      className="mt-2 text-xs text-rose-300 hover:text-rose-200"
                    >
                      Revoke key
                    </button>
                  )}
                </article>
              ))}
            </div>
          )}
        </section>
      </main>
    </AppNav>
  );
}
