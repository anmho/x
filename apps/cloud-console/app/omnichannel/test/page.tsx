'use client';

import { FormEvent, useEffect, useMemo, useState } from 'react';
import { BellRing, CheckCircle2, Clock3, RefreshCw, Send, XCircle } from 'lucide-react';
import { AppNav } from '@/app/_components/app-nav';
import { notificationApi, Notification, NotificationChannel } from '@/lib/api';
import { NotificationLivePreview } from '@/app/_components/notification-live-preview';

function statusPill(status: Notification['status']) {
  const classes: Record<Notification['status'], string> = {
    pending: 'border-amber-500/30 bg-amber-500/10 text-amber-300',
    processing: 'border-blue-500/30 bg-blue-500/10 text-blue-300',
    sent: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-300',
    failed: 'border-rose-500/30 bg-rose-500/10 text-rose-300',
    cancelled: 'border-zinc-600 bg-zinc-800/70 text-zinc-300',
  };

  return <span className={`rounded-full border px-2 py-0.5 text-xs ${classes[status]}`}>{status}</span>;
}

function statusIcon(status: Notification['status']) {
  if (status === 'sent') return <CheckCircle2 className="h-3.5 w-3.5 text-emerald-300" />;
  if (status === 'failed') return <XCircle className="h-3.5 w-3.5 text-rose-300" />;
  if (status === 'processing') return <Clock3 className="h-3.5 w-3.5 animate-spin text-blue-300" />;
  if (status === 'cancelled') return <XCircle className="h-3.5 w-3.5 text-zinc-300" />;
  return <Clock3 className="h-3.5 w-3.5 text-amber-300" />;
}

export default function OmnichannelTestPage() {
  const [channel, setChannel] = useState<NotificationChannel>('email');
  const [recipient, setRecipient] = useState('user@example.com');
  const [subject, setSubject] = useState('Omnichannel channel test');
  const [body, setBody] = useState('<p>Test from Omnichannel test center</p>');
  const [submitting, setSubmitting] = useState(false);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  async function loadNotifications() {
    try {
      setLoading(true);
      const response = await notificationApi.list();
      setNotifications(response.data ?? []);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load tests');
      setSuccess(null);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadNotifications();
  }, []);

  async function submitTest(event: FormEvent) {
    event.preventDefault();
    try {
      setSubmitting(true);
      await notificationApi.create({
        channel,
        recipient,
        recipient_email: recipient,
        subject,
        body,
        metadata: {
          source: 'omnichannel.test',
        },
      });
      setSuccess(`Submitted ${channel} test`);
      setError(null);
      await loadNotifications();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to submit test');
      setSuccess(null);
    } finally {
      setSubmitting(false);
    }
  }

  const recentTests = useMemo(() => notifications.slice(0, 10), [notifications]);

  return (
    <div className="min-h-screen app-shell">
      <AppNav active="omnichannelTest" />

      <main className="mx-auto max-w-6xl px-4 py-6 sm:px-6">
        <div className="mb-4 flex items-center justify-between gap-3">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Omnichannel Test Center</h1>
            <p className="mt-1 text-sm text-zinc-500">Run quick test notifications and verify delivery state.</p>
          </div>
          <button
            type="button"
            onClick={() => void loadNotifications()}
            className="inline-flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-200 hover:bg-zinc-800"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            Refresh
          </button>
        </div>

        {(error || success) && (
          <div className="mb-4 space-y-2">
            {error && <p className="rounded-md border border-rose-500/30 bg-rose-500/10 px-3 py-2 text-sm text-rose-300">{error}</p>}
            {success && <p className="rounded-md border border-emerald-500/30 bg-emerald-500/10 px-3 py-2 text-sm text-emerald-300">{success}</p>}
          </div>
        )}

        <section className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          <article className="rounded-xl border border-zinc-800 bg-zinc-950 p-4">
            <h2 className="inline-flex items-center gap-2 text-sm font-medium text-zinc-200">
              <BellRing className="h-4 w-4" />
              Run Test
            </h2>
            <form onSubmit={submitTest} className="mt-3 space-y-3">
              <select
                value={channel}
                onChange={(event) => setChannel(event.target.value as NotificationChannel)}
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
              >
                <option value="email">email</option>
                <option value="sms">sms</option>
                <option value="push">push</option>
                <option value="webhook">webhook</option>
                <option value="app">app (emulator)</option>
                <option value="imessage">imessage (emulator)</option>
              </select>
              <input
                value={recipient}
                onChange={(event) => setRecipient(event.target.value)}
                placeholder="recipient"
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
              />
              <input
                value={subject}
                onChange={(event) => setSubject(event.target.value)}
                placeholder="subject"
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
              />
              <textarea
                value={body}
                onChange={(event) => setBody(event.target.value)}
                rows={5}
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
              />
              <button
                type="submit"
                disabled={submitting}
                className="inline-flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm hover:bg-zinc-800 disabled:opacity-40"
              >
                <Send className="h-3.5 w-3.5" />
                {submitting ? 'Submitting...' : 'Send Test'}
              </button>
              <p className="text-xs text-zinc-500">
                `app` channel posts to `OMNICHANNEL_APP_RELAY_URL` when configured.
              </p>
            </form>

            <div className="mt-4">
              <p className="mb-2 text-xs uppercase tracking-wide text-zinc-500">Live Preview</p>
              <NotificationLivePreview
                subject={subject}
                body={body}
                recipient={recipient}
              />
            </div>
          </article>

          <article className="rounded-xl border border-zinc-800 bg-zinc-950">
            <div className="border-b border-zinc-800 px-4 py-3">
              <h2 className="text-sm font-medium text-zinc-200">Recent Test Activity</h2>
            </div>
            {loading ? (
              <div className="space-y-2 p-4">
                {Array.from({ length: 5 }).map((_, index) => (
                  <div key={index} className="h-10 animate-pulse rounded-md bg-zinc-900" />
                ))}
              </div>
            ) : recentTests.length === 0 ? (
              <p className="px-4 py-10 text-sm text-zinc-500">No notifications yet.</p>
            ) : (
              <div className="divide-y divide-zinc-800">
                {recentTests.map((item) => (
                  <article key={item.id} className="flex items-center justify-between gap-3 px-4 py-3">
                    <div className="min-w-0">
                      <p className="truncate text-sm text-zinc-200">{item.subject}</p>
                      <p className="truncate text-xs text-zinc-500">{item.recipient_email}</p>
                    </div>
                    <div className="inline-flex items-center gap-1.5">
                      {statusIcon(item.status)}
                      {statusPill(item.status)}
                    </div>
                  </article>
                ))}
              </div>
            )}
          </article>
        </section>
      </main>
    </div>
  );
}
