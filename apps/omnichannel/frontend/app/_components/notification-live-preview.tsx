'use client';

import { useMemo, useState } from 'react';

type PreviewMode = 'email' | 'push' | 'imessage' | 'sms' | 'webhook';

const PREVIEW_MODES: Array<{ id: PreviewMode; label: string }> = [
  { id: 'email', label: 'Email' },
  { id: 'push', label: 'Push' },
  { id: 'imessage', label: 'iMessage' },
  { id: 'sms', label: 'SMS' },
  { id: 'webhook', label: 'Webhook' },
];

function escapeHtml(value: string) {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;');
}

function htmlToText(input: string) {
  return input.replace(/<[^>]*>/g, ' ').replace(/\s+/g, ' ').trim();
}

function textSnippet(input: string, max = 140) {
  if (!input) return 'No message body yet.';
  if (input.length <= max) return input;
  return `${input.slice(0, max - 1)}…`;
}

export function NotificationLivePreview({
  subject,
  body,
  recipient,
}: {
  subject: string;
  body: string;
  recipient?: string;
}) {
  const [mode, setMode] = useState<PreviewMode>('email');
  const plainBody = useMemo(() => htmlToText(body), [body]);
  const previewSubject = subject.trim() || 'Untitled notification';
  const previewRecipient = recipient?.trim() || 'recipient@example.com';
  const previewSnippet = textSnippet(plainBody);

  const emailDocument = useMemo(() => {
    const safeSubject = escapeHtml(previewSubject);
    const renderedBody = body.trim() || '<p style="color:#6b7280">No HTML body provided yet.</p>';
    return `<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>${safeSubject}</title>
    <style>
      body { margin: 0; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #f4f4f5; color: #111827; }
      .frame { max-width: 680px; margin: 20px auto; border-radius: 12px; overflow: hidden; border: 1px solid #e4e4e7; background: #fff; }
      .header { padding: 18px 20px; border-bottom: 1px solid #e4e4e7; }
      .subject { margin: 0; font-size: 18px; font-weight: 600; }
      .meta { margin-top: 6px; color: #6b7280; font-size: 12px; }
      .content { padding: 20px; line-height: 1.5; }
    </style>
  </head>
  <body>
    <div class="frame">
      <div class="header">
        <h1 class="subject">${safeSubject}</h1>
        <p class="meta">To: ${escapeHtml(previewRecipient)}</p>
      </div>
      <div class="content">${renderedBody}</div>
    </div>
  </body>
</html>`;
  }, [body, previewRecipient, previewSubject]);

  return (
    <div className="rounded-xl border border-zinc-800 bg-zinc-950 p-3">
      <div className="mb-3 flex flex-wrap gap-1.5">
        {PREVIEW_MODES.map((item) => (
          <button
            key={item.id}
            type="button"
            onClick={() => setMode(item.id)}
            className={`rounded-md border px-2.5 py-1 text-xs transition-colors ${
              mode === item.id
                ? 'border-zinc-600 bg-zinc-800 text-zinc-100'
                : 'border-zinc-700 bg-zinc-900 text-zinc-400 hover:text-zinc-200'
            }`}
          >
            {item.label}
          </button>
        ))}
      </div>

      {mode === 'email' && (
        <iframe
          title="Live email preview"
          sandbox=""
          srcDoc={emailDocument}
          className="h-[460px] w-full rounded-lg border border-zinc-800 bg-white"
        />
      )}

      {mode === 'push' && (
        <div className="mx-auto w-full max-w-[320px] rounded-3xl border border-zinc-700 bg-zinc-900 p-3 shadow-inner">
          <p className="text-[11px] uppercase tracking-wide text-zinc-500">Now</p>
          <div className="mt-2 rounded-2xl border border-zinc-700 bg-zinc-800/70 p-3">
            <p className="text-xs font-medium text-zinc-200">Project X Notifications</p>
            <p className="mt-1 text-sm font-semibold text-zinc-100">{previewSubject}</p>
            <p className="mt-1 text-xs text-zinc-400">{previewSnippet}</p>
          </div>
        </div>
      )}

      {mode === 'imessage' && (
        <div className="mx-auto w-full max-w-[340px] rounded-3xl border border-zinc-700 bg-zinc-950 p-4">
          <p className="text-center text-xs text-zinc-500">{previewRecipient}</p>
          <div className="mt-4 space-y-2">
            <div className="max-w-[78%] rounded-2xl rounded-bl-md bg-zinc-800 px-3 py-2 text-xs text-zinc-300">
              Previewing iMessage thread
            </div>
            <div className="ml-auto max-w-[78%] rounded-2xl rounded-br-md bg-blue-600 px-3 py-2 text-sm text-white">
              <p className="font-medium">{previewSubject}</p>
              <p className="mt-1 text-xs text-blue-100">{previewSnippet}</p>
            </div>
          </div>
        </div>
      )}

      {mode === 'sms' && (
        <div className="mx-auto w-full max-w-[340px] rounded-3xl border border-zinc-700 bg-zinc-950 p-4">
          <p className="text-xs text-zinc-500">SMS to {previewRecipient}</p>
          <div className="mt-3 max-w-[82%] rounded-2xl rounded-bl-md bg-emerald-600 px-3 py-2 text-sm text-white">
            <p className="font-medium">{previewSubject}</p>
            <p className="mt-1 text-xs text-emerald-100">{previewSnippet}</p>
          </div>
        </div>
      )}

      {mode === 'webhook' && (
        <pre className="overflow-x-auto rounded-lg border border-zinc-800 bg-zinc-900 p-3 text-xs text-zinc-300">
{JSON.stringify(
  {
    channel: 'webhook',
    recipient: previewRecipient,
    subject: previewSubject,
    body: previewSnippet,
    metadata: { source: 'live-preview' },
  },
  null,
  2,
)}
        </pre>
      )}
    </div>
  );
}
