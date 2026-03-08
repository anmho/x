'use client';

import { FormEvent, useEffect, useMemo, useState } from 'react';
import { Calendar, Clock3, Play, Plus, Trash2 } from 'lucide-react';
import { AppNav } from '@/app/_components/app-nav';
import { dispatchCampaign } from '@/lib/omnichannel-dispatch';
import {
  OmnichannelCampaign,
  OmnichannelCronJob,
  loadCampaigns,
  loadCronJobs,
  loadRecipients,
  saveCampaigns,
  saveCronJobs,
} from '@/lib/omnichannel-store';

const EMPTY_FORM = {
  name: '',
  schedule: '0 9 * * *',
  campaignId: '',
};

export default function OmnichannelCronJobsPage() {
  const [jobs, setJobs] = useState<OmnichannelCronJob[]>([]);
  const [campaigns, setCampaigns] = useState<OmnichannelCampaign[]>([]);
  const [hydrated, setHydrated] = useState(false);
  const [form, setForm] = useState(EMPTY_FORM);
  const [runningJobId, setRunningJobId] = useState<string | null>(null);
  const [feedback, setFeedback] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadedCampaigns = loadCampaigns();
    setCampaigns(loadedCampaigns);
    const loadedJobs = loadCronJobs();
    setJobs(loadedJobs);
    setForm((prev) => ({
      ...prev,
      campaignId: loadedCampaigns[0]?.id || '',
    }));
    setHydrated(true);
  }, []);

  useEffect(() => {
    if (!hydrated) return;
    saveCronJobs(jobs);
  }, [jobs, hydrated]);

  function addJob(event: FormEvent) {
    event.preventDefault();
    if (!form.name.trim() || !form.campaignId) return;
    const now = new Date().toISOString();
    setJobs((prev) => [
      {
        id: `job-${Date.now()}`,
        name: form.name.trim(),
        schedule: form.schedule.trim(),
        campaign_id: form.campaignId,
        status: 'active',
        updated_at: now,
      },
      ...prev,
    ]);
    setForm((prev) => ({ ...prev, name: '', schedule: '0 9 * * *' }));
  }

  function toggleStatus(id: string) {
    setJobs((prev) =>
      prev.map((job) =>
        job.id === id ? { ...job, status: job.status === 'active' ? 'paused' : 'active' } : job,
      ),
    );
  }

  function removeJob(id: string) {
    setJobs((prev) => prev.filter((job) => job.id !== id));
  }

  async function runJobNow(job: OmnichannelCronJob) {
    const campaign = campaigns.find((item) => item.id === job.campaign_id);
    if (!campaign) {
      setError('Linked campaign no longer exists.');
      return;
    }

    try {
      setRunningJobId(job.id);
      const recipients = loadRecipients();
      const result = await dispatchCampaign(campaign, recipients);
      const now = new Date().toISOString();
      const nextCampaigns = campaigns.map((item) =>
        item.id === campaign.id
          ? {
              ...item,
              status: result.submitted > 0 ? 'live' : item.status,
              last_run_at: now,
              last_run_count: result.submitted,
              updated_at: now,
            }
          : item,
      );

      setJobs((prev) =>
        prev.map((item) =>
          item.id === job.id
            ? {
                ...item,
                last_run_at: now,
                last_result: result.message,
                updated_at: now,
              }
            : item,
        ),
      );

      setCampaigns(nextCampaigns);
      saveCampaigns(nextCampaigns);

      setFeedback(result.message);
      setError(result.submitted > 0 ? null : result.message);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to run cron job');
      setFeedback(null);
    } finally {
      setRunningJobId(null);
    }
  }

  const campaignOptions = useMemo(
    () => campaigns.map((campaign) => ({ value: campaign.id, label: campaign.name })),
    [campaigns],
  );

  return (
    <div className="min-h-screen app-shell">
      <AppNav active="cronJobs" />

      <main className="mx-auto max-w-6xl px-4 py-6 sm:px-6">
        <div className="mb-4">
          <h1 className="text-2xl font-semibold tracking-tight">Cron Jobs</h1>
          <p className="mt-1 text-sm text-zinc-500">Schedule campaign runs and execute them on demand.</p>
        </div>

        {(feedback || error) && (
          <div className="mb-4 space-y-2">
            {feedback && <p className="rounded-md border border-emerald-500/30 bg-emerald-500/10 px-3 py-2 text-sm text-emerald-300">{feedback}</p>}
            {error && <p className="rounded-md border border-rose-500/30 bg-rose-500/10 px-3 py-2 text-sm text-rose-300">{error}</p>}
          </div>
        )}

        <section className="grid grid-cols-1 gap-4 lg:grid-cols-[360px_1fr]">
          <article className="rounded-xl border border-zinc-800 bg-zinc-950 p-4">
            <h2 className="inline-flex items-center gap-2 text-sm font-medium text-zinc-200">
              <Plus className="h-4 w-4" />
              New Cron Job
            </h2>
            <form onSubmit={addJob} className="mt-3 space-y-3">
              <input
                value={form.name}
                onChange={(event) => setForm((prev) => ({ ...prev, name: event.target.value }))}
                placeholder="job name"
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
              />
              <input
                value={form.schedule}
                onChange={(event) => setForm((prev) => ({ ...prev, schedule: event.target.value }))}
                placeholder="cron expression"
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm font-mono"
              />
              <select
                value={form.campaignId}
                onChange={(event) => setForm((prev) => ({ ...prev, campaignId: event.target.value }))}
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
              >
                {campaignOptions.map((option) => (
                  <option key={option.value} value={option.value}>{option.label}</option>
                ))}
              </select>
              <button
                type="submit"
                className="inline-flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm hover:bg-zinc-800"
              >
                <Plus className="h-3.5 w-3.5" />
                Create
              </button>
            </form>
            <p className="mt-3 text-xs text-zinc-500">Example: `0 9 * * *` means daily at 09:00.</p>
          </article>

          <article className="rounded-xl border border-zinc-800 bg-zinc-950">
            <div className="border-b border-zinc-800 px-4 py-3">
              <h2 className="inline-flex items-center gap-2 text-sm font-medium text-zinc-200">
                <Calendar className="h-4 w-4" />
                Scheduled Jobs
              </h2>
            </div>
            <div className="divide-y divide-zinc-800">
              {jobs.map((job) => {
                const campaign = campaigns.find((item) => item.id === job.campaign_id);
                return (
                  <article key={job.id} className="flex items-center justify-between gap-3 px-4 py-3">
                    <div className="min-w-0">
                      <p className="truncate text-sm text-zinc-200">{job.name}</p>
                      <p className="truncate text-xs text-zinc-500">
                        target: {campaign?.name || 'missing campaign'}
                      </p>
                      <p className="mt-1 inline-flex items-center gap-1 font-mono text-xs text-zinc-400">
                        <Clock3 className="h-3.5 w-3.5" />
                        {job.schedule}
                      </p>
                      {job.last_result && <p className="mt-1 text-xs text-zinc-500">{job.last_result}</p>}
                    </div>
                    <div className="flex items-center gap-2">
                      <button
                        type="button"
                        onClick={() => toggleStatus(job.id)}
                        className={`rounded-full border px-2 py-0.5 text-xs ${
                          job.status === 'active'
                            ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-300'
                            : 'border-amber-500/30 bg-amber-500/10 text-amber-300'
                        }`}
                      >
                        {job.status}
                      </button>
                      <button
                        type="button"
                        onClick={() => void runJobNow(job)}
                        disabled={runningJobId === job.id || job.status !== 'active'}
                        className="inline-flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-900 px-2 py-1 text-xs text-zinc-300 hover:bg-zinc-800 disabled:opacity-40"
                      >
                        <Play className="h-3.5 w-3.5" />
                        {runningJobId === job.id ? 'Running' : 'Run now'}
                      </button>
                      <button
                        type="button"
                        onClick={() => removeJob(job.id)}
                        className="inline-flex items-center gap-1 text-xs text-rose-300 hover:text-rose-200"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                        Delete
                      </button>
                    </div>
                  </article>
                );
              })}
              {jobs.length === 0 && (
                <p className="px-4 py-8 text-sm text-zinc-500">No cron jobs yet.</p>
              )}
            </div>
          </article>
        </section>
      </main>
    </div>
  );
}
