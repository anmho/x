'use client';

import { FormEvent, useEffect, useMemo, useState } from 'react';
import { Network, Plus, Search, Trash2, Users } from 'lucide-react';
import { AppNav } from '@/app/_components/app-nav';
import { NotificationChannel } from '@/lib/api';
import {
  OmnichannelRecipient,
  loadRecipients,
  saveRecipients,
} from '@/lib/omnichannel-store';

export default function OmnichannelRecipientsPage() {
  const [recipients, setRecipients] = useState<OmnichannelRecipient[]>([]);
  const [hydrated, setHydrated] = useState(false);
  const [destination, setDestination] = useState('');
  const [channel, setChannel] = useState<NotificationChannel>('email');
  const [segment, setSegment] = useState('audience');
  const [query, setQuery] = useState('');
  const [segmentFilter, setSegmentFilter] = useState('all');

  useEffect(() => {
    setRecipients(loadRecipients());
    setHydrated(true);
  }, []);

  useEffect(() => {
    if (!hydrated) return;
    saveRecipients(recipients);
  }, [recipients, hydrated]);

  function addRecipient(event: FormEvent) {
    event.preventDefault();
    if (!destination.trim()) return;
    setRecipients((prev) => [
      {
        id: `r-${Date.now()}`,
        destination: destination.trim(),
        channel,
        segment: segment.trim() || 'audience',
        status: 'active',
        created_at: new Date().toISOString(),
      },
      ...prev,
    ]);
    setDestination('');
    setChannel('email');
    setSegment('audience');
  }

  function removeRecipient(id: string) {
    setRecipients((prev) => prev.filter((item) => item.id !== id));
  }

  function toggleStatus(id: string) {
    setRecipients((prev) =>
      prev.map((item) =>
        item.id === id
          ? { ...item, status: item.status === 'active' ? 'paused' : 'active' }
          : item,
      ),
    );
  }

  const segments = useMemo(() => {
    const unique = new Set(recipients.map((recipient) => recipient.segment));
    return Array.from(unique).sort();
  }, [recipients]);

  const filteredRecipients = useMemo(() => {
    return recipients.filter((recipient) => {
      if (segmentFilter !== 'all' && recipient.segment !== segmentFilter) return false;
      if (!query) return true;
      const text = `${recipient.destination} ${recipient.segment} ${recipient.channel}`.toLowerCase();
      return text.includes(query.toLowerCase());
    });
  }, [recipients, query, segmentFilter]);

  return (
    <div className="min-h-screen app-shell">
      <AppNav active="recipients" />

      <main className="mx-auto max-w-6xl px-4 py-6 sm:px-6">
        <div className="mb-4">
          <h1 className="text-2xl font-semibold tracking-tight">Recipients</h1>
          <p className="mt-1 text-sm text-zinc-500">Persistent destination lists grouped by segment and channel for campaigns and cron jobs.</p>
        </div>

        <section className="grid grid-cols-1 gap-4 lg:grid-cols-[360px_1fr]">
          <article className="rounded-xl border border-zinc-800 bg-zinc-950 p-4">
            <h2 className="inline-flex items-center gap-2 text-sm font-medium text-zinc-200">
              <Plus className="h-4 w-4" />
              Add Recipient
            </h2>
            <form onSubmit={addRecipient} className="mt-3 space-y-3">
              <input
                value={destination}
                onChange={(event) => setDestination(event.target.value)}
                placeholder="email, webhook URL, or app device id"
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
              />
              <select
                value={channel}
                onChange={(event) => setChannel(event.target.value as NotificationChannel)}
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
              >
                <option value="email">email</option>
                <option value="sms">sms</option>
                <option value="push">push</option>
                <option value="webhook">webhook</option>
                <option value="app">app</option>
                <option value="imessage">imessage</option>
              </select>
              <input
                value={segment}
                onChange={(event) => setSegment(event.target.value)}
                placeholder="segment"
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
              />
              <button
                type="submit"
                className="inline-flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm hover:bg-zinc-800"
              >
                <Plus className="h-3.5 w-3.5" />
                Add
              </button>
            </form>

            <div className="mt-4 rounded-md border border-zinc-800 bg-zinc-900/30 p-3">
              <p className="inline-flex items-center gap-1 text-xs uppercase tracking-wide text-zinc-500">
                <Users className="h-3.5 w-3.5" />
                Segments
              </p>
              <div className="mt-2 flex flex-wrap gap-1.5">
                {segments.map((value) => (
                  <span key={value} className="rounded-full border border-zinc-700 bg-zinc-900 px-2 py-0.5 text-xs text-zinc-300">
                    {value}
                  </span>
                ))}
                {segments.length === 0 && (
                  <span className="text-xs text-zinc-500">No segments yet</span>
                )}
              </div>
            </div>
          </article>

          <article className="rounded-xl border border-zinc-800 bg-zinc-950">
            <div className="border-b border-zinc-800 px-4 py-3">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <h2 className="inline-flex items-center gap-2 text-sm font-medium text-zinc-200">
                  <Network className="h-4 w-4" />
                  Recipient Directory
                </h2>
                <div className="flex items-center gap-2">
                  <label className="relative block">
                    <Search className="pointer-events-none absolute left-2 top-2 h-3.5 w-3.5 text-zinc-500" />
                    <input
                      value={query}
                      onChange={(event) => setQuery(event.target.value)}
                      placeholder="Search"
                      className="w-40 rounded-md border border-zinc-700 bg-zinc-900 py-1.5 pl-7 pr-2 text-xs text-zinc-200"
                    />
                  </label>
                  <select
                    value={segmentFilter}
                    onChange={(event) => setSegmentFilter(event.target.value)}
                    className="rounded-md border border-zinc-700 bg-zinc-900 px-2 py-1.5 text-xs text-zinc-200"
                  >
                    <option value="all">all segments</option>
                    {segments.map((value) => (
                      <option key={value} value={value}>{value}</option>
                    ))}
                  </select>
                </div>
              </div>
            </div>
            <div className="divide-y divide-zinc-800">
              {filteredRecipients.map((recipient) => (
                <article key={recipient.id} className="flex items-center justify-between gap-3 px-4 py-3">
                  <div className="min-w-0">
                    <p className="truncate text-sm text-zinc-200">{recipient.destination}</p>
                    <p className="text-xs text-zinc-500">
                      {recipient.channel} • {recipient.segment}
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    <button
                      type="button"
                      onClick={() => toggleStatus(recipient.id)}
                      className={`rounded-full border px-2 py-0.5 text-xs ${
                        recipient.status === 'active'
                          ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-300'
                          : 'border-amber-500/30 bg-amber-500/10 text-amber-300'
                      }`}
                    >
                      {recipient.status}
                    </button>
                    <button
                      type="button"
                      onClick={() => removeRecipient(recipient.id)}
                      className="inline-flex items-center gap-1 text-xs text-rose-300 hover:text-rose-200"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                      Remove
                    </button>
                  </div>
                </article>
              ))}
              {filteredRecipients.length === 0 && (
                <p className="px-4 py-8 text-sm text-zinc-500">No recipients match current filters.</p>
              )}
            </div>
          </article>
        </section>
      </main>
    </div>
  );
}
