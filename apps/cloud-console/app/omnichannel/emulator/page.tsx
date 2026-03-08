'use client';

import { useEffect, useState } from 'react';
import { RefreshCw, Smartphone, Trash2 } from 'lucide-react';
import { AppNav } from '@/app/_components/app-nav';

type EmulatorMessage = {
  id: string;
  channel: string;
  destination: string;
  title: string;
  body: string;
  sent_at: string;
};

function formatDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString();
}

export default function OmnichannelEmulatorPage() {
  const [messages, setMessages] = useState<EmulatorMessage[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  async function loadMessages() {
    try {
      setLoading(true);
      const response = await fetch('/api/emulator/messages', { cache: 'no-store' });
      if (!response.ok) throw new Error(`Failed with ${response.status}`);
      const data = await response.json();
      setMessages(data.data || []);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load emulator messages');
    } finally {
      setLoading(false);
    }
  }

  async function clearMessages() {
    await fetch('/api/emulator/messages', { method: 'DELETE' });
    await loadMessages();
  }

  useEffect(() => {
    void loadMessages();
  }, []);

  return (
    <div className="min-h-screen app-shell">
      <AppNav active="emulator" />

      <main className="mx-auto max-w-6xl px-4 py-6 sm:px-6">
        <div className="mb-4 flex items-center justify-between gap-3">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">App Emulator Inbox</h1>
            <p className="mt-1 text-sm text-zinc-500">Receives `app` and `push` relay payloads for local emulator testing.</p>
          </div>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => void loadMessages()}
              className="inline-flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-200 hover:bg-zinc-800"
            >
              <RefreshCw className="h-3.5 w-3.5" />
              Refresh
            </button>
            <button
              type="button"
              onClick={() => void clearMessages()}
              className="inline-flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-200 hover:bg-zinc-800"
            >
              <Trash2 className="h-3.5 w-3.5" />
              Clear
            </button>
          </div>
        </div>

        {error && (
          <p className="mb-4 rounded-md border border-rose-500/30 bg-rose-500/10 px-3 py-2 text-sm text-rose-300">{error}</p>
        )}

        <section className="rounded-xl border border-zinc-800 bg-zinc-950">
          <div className="border-b border-zinc-800 px-4 py-3">
            <h2 className="inline-flex items-center gap-2 text-sm font-medium text-zinc-200">
              <Smartphone className="h-4 w-4" />
              Received Messages
            </h2>
          </div>

          {loading ? (
            <div className="space-y-2 p-4">
              {Array.from({ length: 5 }).map((_, idx) => (
                <div key={idx} className="h-14 animate-pulse rounded-md bg-zinc-900" />
              ))}
            </div>
          ) : messages.length === 0 ? (
            <p className="px-4 py-12 text-sm text-zinc-500">No emulator messages yet.</p>
          ) : (
            <div className="divide-y divide-zinc-800">
              {messages.map((message) => (
                <article key={message.id} className="px-4 py-3">
                  <div className="flex items-center justify-between gap-3">
                    <p className="text-sm font-medium text-zinc-100">{message.title || '(no title)'}</p>
                    <span className="rounded-full border border-zinc-700 bg-zinc-900 px-2 py-0.5 text-xs text-zinc-300">
                      {message.channel}
                    </span>
                  </div>
                  <p className="mt-1 text-xs text-zinc-500">destination: {message.destination || 'n/a'}</p>
                  <p className="mt-2 text-sm text-zinc-300">{message.body || '(no body)'}</p>
                  <p className="mt-2 text-xs text-zinc-600">{formatDate(message.sent_at)}</p>
                </article>
              ))}
            </div>
          )}
        </section>
      </main>
    </div>
  );
}
