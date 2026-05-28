'use client';

import { FormEvent, Suspense, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';

function LoginContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const next = searchParams.get('next') || '/';

  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const response = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ password }),
      });

      if (!response.ok) {
        const data = (await response.json()) as { error?: string };
        throw new Error(data.error || 'Login failed');
      }

      router.push(next);
      router.refresh();
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Login failed';
      setError(message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen app-shell">
      <main className="mx-auto flex min-h-screen max-w-md items-center px-4">
        <section className="w-full rounded-xl border border-zinc-800 bg-zinc-950 p-6">
          <h1 className="text-xl font-semibold tracking-tight">Sign in</h1>
          <p className="mt-1 text-sm text-zinc-400">Enter console password to continue.</p>

          {error && (
            <p className="mt-3 rounded-md border border-rose-500/30 bg-rose-500/10 px-3 py-2 text-sm text-rose-300">
              {error}
            </p>
          )}

          <form className="mt-4 space-y-3" onSubmit={onSubmit}>
            <label className="block">
              <span className="mb-1 block text-sm text-zinc-400">Password</span>
              <input
                type="password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2"
              />
            </label>

            <button
              type="submit"
              disabled={loading}
              className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm font-medium hover:bg-zinc-800 disabled:opacity-60"
            >
              {loading ? 'Signing in...' : 'Sign in'}
            </button>
          </form>
        </section>
      </main>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen app-shell">
          <main className="mx-auto flex min-h-screen max-w-md items-center px-4">
            <section className="w-full rounded-xl border border-zinc-800 bg-zinc-950 p-6">
              <h1 className="text-xl font-semibold tracking-tight">Sign in</h1>
            </section>
          </main>
        </div>
      }
    >
      <LoginContent />
    </Suspense>
  );
}
