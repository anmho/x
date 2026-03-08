import { dispatchCampaign } from '@/lib/omnichannel-dispatch';
import { isCronDue } from '@/lib/cron';
import {
  loadCampaigns,
  loadCronJobs,
  loadRecipients,
  saveCampaigns,
  saveCronJobs,
} from '@/lib/omnichannel-store';

const LOCK_KEY = 'omnichannel:cron:lock:v1';
const LOCK_TIMEOUT_MS = 25_000;

function acquireLock(): boolean {
  if (typeof window === 'undefined') return false;
  const now = Date.now();
  const current = window.localStorage.getItem(LOCK_KEY);
  if (current) {
    const ts = Number.parseInt(current, 10);
    if (Number.isFinite(ts) && now-ts < LOCK_TIMEOUT_MS) return false;
  }
  window.localStorage.setItem(LOCK_KEY, String(now));
  return true;
}

function releaseLock() {
  if (typeof window === 'undefined') return;
  window.localStorage.removeItem(LOCK_KEY);
}

export async function runDueCronJobsTick(now = new Date()) {
  if (!acquireLock()) return;
  try {
    const jobs = loadCronJobs();
    const campaigns = loadCampaigns();
    const recipients = loadRecipients();

    let changed = false;
    const nextJobs = [...jobs];
    const nextCampaigns = [...campaigns];

    for (let i = 0; i < nextJobs.length; i += 1) {
      const job = nextJobs[i];
      if (job.status !== 'active') continue;
      if (!isCronDue(job.schedule, now, job.last_run_at)) continue;

      const campaignIndex = nextCampaigns.findIndex((item) => item.id === job.campaign_id);
      if (campaignIndex < 0) {
        nextJobs[i] = {
          ...job,
          last_run_at: now.toISOString(),
          last_result: 'Skipped: linked campaign not found.',
          updated_at: now.toISOString(),
        };
        changed = true;
        continue;
      }

      const campaign = nextCampaigns[campaignIndex];
      const result = await dispatchCampaign(campaign, recipients);
      const runAt = now.toISOString();

      nextJobs[i] = {
        ...job,
        last_run_at: runAt,
        last_result: result.message,
        updated_at: runAt,
      };

      nextCampaigns[campaignIndex] = {
        ...campaign,
        status: result.submitted > 0 ? 'live' : campaign.status,
        last_run_at: runAt,
        last_run_count: result.submitted,
        updated_at: runAt,
      };
      changed = true;
    }

    if (changed) {
      saveCronJobs(nextJobs);
      saveCampaigns(nextCampaigns);
    }
  } finally {
    releaseLock();
  }
}
