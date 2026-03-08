'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { NotificationChannel, notificationApi } from '@/lib/api';
import { ArrowLeft, Send } from 'lucide-react';
import { AppNav } from '@/app/_components/app-nav';
import { NotificationLivePreview } from '@/app/_components/notification-live-preview';

export default function CreateNotification() {
  const router = useRouter();
  const [channel, setChannel] = useState<NotificationChannel>('email');
  const [recipientEmail, setRecipientEmail] = useState('');
  const [subject, setSubject] = useState('');
  const [body, setBody] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      await notificationApi.create({
        channel,
        recipient_email: recipientEmail,
        subject,
        body,
      });
      setSuccess(true);
      setTimeout(() => router.push('/deployments'), 1200);
    } catch (err: any) {
      setError(err.response?.data?.message || err.message || 'Failed to send notification');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen app-shell">
      <AppNav active="notifications" />

      <main className="mx-auto max-w-6xl px-4 py-6 sm:px-6">
        <Link href="/omnichannel" className="mb-4 inline-flex items-center gap-1 text-sm text-zinc-400 hover:text-zinc-200">
          <ArrowLeft className="h-4 w-4" /> Back
        </Link>

        <section className="grid grid-cols-1 gap-4 xl:grid-cols-[1.05fr_0.95fr]">
          <article className="rounded-xl border border-zinc-800 bg-zinc-950 p-5">
            <h1 className="text-xl font-semibold tracking-tight">Send Notification</h1>
            <p className="mt-1 text-sm text-zinc-500">Template management lives in Templates. This flow is for direct test and manual sends.</p>

            {success && <p className="mt-3 rounded-md border border-emerald-500/30 bg-emerald-500/10 px-3 py-2 text-sm text-emerald-300">Notification sent. Redirecting...</p>}
            {error && <p className="mt-3 rounded-md border border-rose-500/30 bg-rose-500/10 px-3 py-2 text-sm text-rose-300">{error}</p>}

            <form onSubmit={handleSubmit} className="mt-4 space-y-4 text-sm">
              <Field label="Channel">
                <select
                  value={channel}
                  onChange={(e) => setChannel(e.target.value as NotificationChannel)}
                  className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2"
                >
                  <option value="email">email</option>
                  <option value="sms">sms</option>
                  <option value="push">push</option>
                  <option value="webhook">webhook</option>
                  <option value="app">app (emulator)</option>
                  <option value="imessage">imessage (emulator)</option>
                </select>
              </Field>

              <Field label={channel === 'webhook' ? 'Webhook URL' : channel === 'app' || channel === 'imessage' ? 'Device / Handle' : 'Recipient'}>
                <input
                  type={channel === 'email' ? 'email' : 'text'}
                  required
                  value={recipientEmail}
                  onChange={(e) => setRecipientEmail(e.target.value)}
                  className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2"
                  placeholder={channel === 'webhook' ? 'https://example.com/webhook' : channel === 'app' || channel === 'imessage' ? 'device-or-user-id' : 'user@example.com'}
                />
              </Field>

              <Field label="Subject">
                <input
                  type="text"
                  required
                  value={subject}
                  onChange={(e) => setSubject(e.target.value)}
                  className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2"
                />
              </Field>

              <Field label="Message Body (HTML)">
                <textarea
                  required
                  value={body}
                  onChange={(e) => setBody(e.target.value)}
                  rows={10}
                  className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 font-mono"
                />
              </Field>

              <div className="flex justify-end">
                <button
                  type="submit"
                  disabled={loading}
                  className="inline-flex items-center gap-2 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 font-medium hover:bg-zinc-800 disabled:opacity-60"
                >
                  <Send className="h-4 w-4" />
                  {loading ? 'Sending...' : 'Send Notification'}
                </button>
              </div>
            </form>
          </article>

          <article className="rounded-xl border border-zinc-800 bg-zinc-950 p-5">
            <h2 className="text-sm font-medium uppercase tracking-wide text-zinc-400">Live Preview</h2>
            <p className="mt-1 text-xs text-zinc-500">Real-time rendering for email, push, iMessage, SMS, and webhook payload views.</p>
            <div className="mt-3">
              <NotificationLivePreview
                subject={subject}
                body={body}
                recipient={recipientEmail}
              />
            </div>
          </article>
        </section>
      </main>
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="mb-1 block text-zinc-400">{label}</span>
      {children}
    </label>
  );
}
