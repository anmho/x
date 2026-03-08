'use client';

import { FormEvent, useEffect, useMemo, useState } from 'react';
import { Megaphone, Play, Plus, Rocket, Trash2 } from 'lucide-react';
import { AppNav } from '@/app/_components/app-nav';
import { NotificationChannel, templateApi, Template } from '@/lib/api';
import { dispatchCampaign } from '@/lib/omnichannel-dispatch';
import {
  OmnichannelCampaign,
  loadCampaigns,
  loadRecipients,
  saveCampaigns,
} from '@/lib/omnichannel-store';

const STATUS_ORDER: OmnichannelCampaign['status'][] = ['draft', 'scheduled', 'live', 'completed'];

const EMPTY_FORM = {
  name: '',
  segment: 'audience',
  channel: 'email' as NotificationChannel,
  subject: '',
  body: '',
  templateId: '',
};

export default function OmnichannelCampaignsPage() {
  const [campaigns, setCampaigns] = useState<OmnichannelCampaign[]>([]);
  const [hydrated, setHydrated] = useState(false);
  const [templates, setTemplates] = useState<Template[]>([]);
  const [form, setForm] = useState(EMPTY_FORM);
  const [busyCampaignId, setBusyCampaignId] = useState<string | null>(null);
  const [feedback, setFeedback] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setCampaigns(loadCampaigns());
    void loadTemplates();
    setHydrated(true);
  }, []);

  useEffect(() => {
    if (!hydrated) return;
    saveCampaigns(campaigns);
  }, [campaigns, hydrated]);

  async function loadTemplates() {
    try {
      const response = await templateApi.list();
      setTemplates(response.data || []);
    } catch {
      setTemplates([]);
    }
  }

  function applyTemplate(templateId: string) {
    const template = templates.find((item) => item.id === templateId);
    if (!template) return;
    setForm((prev) => ({
      ...prev,
      templateId: template.id,
      subject: template.subject,
      body: template.body,
    }));
  }

  function addCampaign(event: FormEvent) {
    event.preventDefault();
    if (!form.name.trim()) return;
    if (!form.templateId && (!form.subject.trim() || !form.body.trim())) {
      setError('Provide subject/body or select a template.');
      return;
    }

    const now = new Date().toISOString();
    setCampaigns((prev) => [
      {
        id: `camp-${Date.now()}`,
        name: form.name.trim(),
        segment: form.segment.trim() || 'audience',
        channel: form.channel,
        status: 'draft',
        subject: form.subject.trim(),
        body: form.body,
        template_id: form.templateId || undefined,
        updated_at: now,
      },
      ...prev,
    ]);
    setForm(EMPTY_FORM);
    setError(null);
  }

  function advanceStatus(id: string) {
    setCampaigns((prev) =>
      prev.map((campaign) => {
        if (campaign.id !== id) return campaign;
        const currentIndex = STATUS_ORDER.indexOf(campaign.status);
        const nextIndex = Math.min(currentIndex + 1, STATUS_ORDER.length - 1);
        return { ...campaign, status: STATUS_ORDER[nextIndex], updated_at: new Date().toISOString() };
      }),
    );
  }

  function removeCampaign(id: string) {
    setCampaigns((prev) => prev.filter((campaign) => campaign.id !== id));
  }

  async function launchCampaign(campaign: OmnichannelCampaign) {
    try {
      setBusyCampaignId(campaign.id);
      const recipients = loadRecipients();
      const result = await dispatchCampaign(campaign, recipients);
      setFeedback(result.message);
      setError(result.submitted > 0 ? null : result.message);
      const now = new Date().toISOString();
      setCampaigns((prev) =>
        prev.map((item) =>
          item.id === campaign.id
            ? {
                ...item,
                status: result.submitted > 0 ? 'live' : item.status,
                last_run_at: now,
                last_run_count: result.submitted,
                updated_at: now,
              }
            : item,
        ),
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to launch campaign');
      setFeedback(null);
    } finally {
      setBusyCampaignId(null);
    }
  }

  const liveCount = useMemo(
    () => campaigns.filter((campaign) => campaign.status === 'live').length,
    [campaigns],
  );

  return (
    <div className="min-h-screen app-shell">
      <AppNav active="campaigns" />

      <main className="mx-auto max-w-6xl px-4 py-6 sm:px-6">
        <div className="mb-4 flex items-center justify-between gap-3">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Campaigns</h1>
            <p className="mt-1 text-sm text-zinc-500">Create campaigns and dispatch them to matching active recipients by segment and channel.</p>
          </div>
          <div className="rounded-md border border-zinc-800 bg-zinc-950 px-3 py-2 text-sm text-zinc-300">
            Live campaigns: <span className="font-medium text-zinc-100">{liveCount}</span>
          </div>
        </div>

        {(feedback || error) && (
          <div className="mb-4 space-y-2">
            {feedback && <p className="rounded-md border border-emerald-500/30 bg-emerald-500/10 px-3 py-2 text-sm text-emerald-300">{feedback}</p>}
            {error && <p className="rounded-md border border-rose-500/30 bg-rose-500/10 px-3 py-2 text-sm text-rose-300">{error}</p>}
          </div>
        )}

        <section className="grid grid-cols-1 gap-4 lg:grid-cols-[400px_1fr]">
          <article className="rounded-xl border border-zinc-800 bg-zinc-950 p-4">
            <h2 className="inline-flex items-center gap-2 text-sm font-medium text-zinc-200">
              <Plus className="h-4 w-4" />
              New Campaign
            </h2>
            <form onSubmit={addCampaign} className="mt-3 space-y-3">
              <input
                value={form.name}
                onChange={(event) => setForm((prev) => ({ ...prev, name: event.target.value }))}
                placeholder="campaign name"
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
              />
              <input
                value={form.segment}
                onChange={(event) => setForm((prev) => ({ ...prev, segment: event.target.value }))}
                placeholder="target segment"
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
              />
              <select
                value={form.channel}
                onChange={(event) => setForm((prev) => ({ ...prev, channel: event.target.value as NotificationChannel }))}
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
              >
                <option value="email">email</option>
                <option value="sms">sms</option>
                <option value="push">push</option>
                <option value="webhook">webhook</option>
                <option value="app">app</option>
                <option value="imessage">imessage</option>
              </select>
              <select
                value={form.templateId}
                onChange={(event) => {
                  const value = event.target.value;
                  if (!value) {
                    setForm((prev) => ({ ...prev, templateId: '' }));
                    return;
                  }
                  applyTemplate(value);
                }}
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
              >
                <option value="">No template</option>
                {templates.map((template) => (
                  <option key={template.id} value={template.id}>{template.name}</option>
                ))}
              </select>
              <input
                value={form.subject}
                onChange={(event) => setForm((prev) => ({ ...prev, subject: event.target.value }))}
                placeholder="subject"
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
              />
              <textarea
                value={form.body}
                onChange={(event) => setForm((prev) => ({ ...prev, body: event.target.value }))}
                rows={5}
                placeholder="body (HTML)"
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
              />
              <button
                type="submit"
                className="inline-flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm hover:bg-zinc-800"
              >
                <Plus className="h-3.5 w-3.5" />
                Create
              </button>
            </form>
          </article>

          <article className="rounded-xl border border-zinc-800 bg-zinc-950">
            <div className="border-b border-zinc-800 px-4 py-3">
              <h2 className="inline-flex items-center gap-2 text-sm font-medium text-zinc-200">
                <Megaphone className="h-4 w-4" />
                Campaign Board
              </h2>
            </div>
            <div className="divide-y divide-zinc-800">
              {campaigns.map((campaign) => (
                <article key={campaign.id} className="flex items-center justify-between gap-3 px-4 py-3">
                  <div className="min-w-0">
                    <p className="truncate text-sm text-zinc-200">{campaign.name}</p>
                    <p className="text-xs text-zinc-500">
                      {campaign.channel} • segment:{' '}
                      {campaign.segment}
                      {campaign.last_run_count ? ` • last run: ${campaign.last_run_count}` : ''}
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="rounded-full border border-zinc-700 bg-zinc-900 px-2 py-0.5 text-xs text-zinc-300">
                      {campaign.status}
                    </span>
                    <button
                      type="button"
                      onClick={() => advanceStatus(campaign.id)}
                      className="inline-flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-900 px-2 py-1 text-xs text-zinc-300 hover:bg-zinc-800"
                      disabled={campaign.status === 'completed'}
                    >
                      <Rocket className="h-3.5 w-3.5" />
                      Advance
                    </button>
                    <button
                      type="button"
                      onClick={() => void launchCampaign(campaign)}
                      className="inline-flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-900 px-2 py-1 text-xs text-zinc-300 hover:bg-zinc-800"
                      disabled={busyCampaignId === campaign.id}
                    >
                      <Play className="h-3.5 w-3.5" />
                      {busyCampaignId === campaign.id ? 'Running' : 'Launch'}
                    </button>
                    <button
                      type="button"
                      onClick={() => removeCampaign(campaign.id)}
                      className="inline-flex items-center gap-1 text-xs text-rose-300 hover:text-rose-200"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                      Delete
                    </button>
                  </div>
                </article>
              ))}
              {campaigns.length === 0 && (
                <p className="px-4 py-8 text-sm text-zinc-500">No campaigns yet.</p>
              )}
            </div>
          </article>
        </section>
      </main>
    </div>
  );
}
