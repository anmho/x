import { NotificationChannel } from '@/lib/api';

export type OmnichannelRecipient = {
  id: string;
  destination: string;
  channel: NotificationChannel;
  segment: string;
  status: 'active' | 'paused';
  created_at: string;
};

export type OmnichannelCampaign = {
  id: string;
  name: string;
  segment: string;
  channel: NotificationChannel;
  status: 'draft' | 'scheduled' | 'live' | 'completed';
  subject: string;
  body: string;
  template_id?: string;
  last_run_at?: string;
  last_run_count?: number;
  updated_at: string;
};

export type OmnichannelCronJob = {
  id: string;
  name: string;
  schedule: string;
  campaign_id: string;
  status: 'active' | 'paused';
  last_run_at?: string;
  last_result?: string;
  updated_at: string;
};

const RECIPIENTS_KEY = 'omnichannel:recipients:v1';
const CAMPAIGNS_KEY = 'omnichannel:campaigns:v1';
const CRON_JOBS_KEY = 'omnichannel:cron-jobs:v1';

const DEFAULT_RECIPIENTS: OmnichannelRecipient[] = [
  {
    id: 'r-1',
    destination: 'core-team@example.com',
    channel: 'email',
    segment: 'internal',
    status: 'active',
    created_at: '2026-03-08T00:00:00.000Z',
  },
  {
    id: 'r-2',
    destination: 'beta-users@example.com',
    channel: 'email',
    segment: 'beta',
    status: 'active',
    created_at: '2026-03-08T00:00:00.000Z',
  },
  {
    id: 'r-3',
    destination: 'ios-sim-001',
    channel: 'app',
    segment: 'beta',
    status: 'active',
    created_at: '2026-03-08T00:00:00.000Z',
  },
];

const DEFAULT_CAMPAIGNS: OmnichannelCampaign[] = [
  {
    id: 'camp-1',
    name: 'Spring Launch',
    segment: 'beta',
    channel: 'email',
    status: 'scheduled',
    subject: 'Spring launch is live',
    body: '<p>Hi {{name}}, our spring launch is now live.</p>',
    updated_at: '2026-03-08T00:00:00.000Z',
  },
  {
    id: 'camp-2',
    name: 'Beta App Ping',
    segment: 'beta',
    channel: 'app',
    status: 'draft',
    subject: 'App test ping',
    body: '<p>Deliver this to the app emulator inbox.</p>',
    updated_at: '2026-03-08T00:00:00.000Z',
  },
];

const DEFAULT_CRON_JOBS: OmnichannelCronJob[] = [
  {
    id: 'job-1',
    name: 'Daily Spring Launch',
    schedule: '0 9 * * *',
    campaign_id: 'camp-1',
    status: 'active',
    updated_at: '2026-03-08T00:00:00.000Z',
  },
];

function readStore<T>(key: string, fallback: T): T {
  if (typeof window === 'undefined') return fallback;
  try {
    const raw = window.localStorage.getItem(key);
    if (!raw) return fallback;
    return JSON.parse(raw) as T;
  } catch {
    return fallback;
  }
}

function writeStore<T>(key: string, value: T) {
  if (typeof window === 'undefined') return;
  window.localStorage.setItem(key, JSON.stringify(value));
}

export function loadRecipients() {
  return readStore<OmnichannelRecipient[]>(RECIPIENTS_KEY, DEFAULT_RECIPIENTS);
}

export function saveRecipients(recipients: OmnichannelRecipient[]) {
  writeStore(RECIPIENTS_KEY, recipients);
}

export function loadCampaigns() {
  return readStore<OmnichannelCampaign[]>(CAMPAIGNS_KEY, DEFAULT_CAMPAIGNS);
}

export function saveCampaigns(campaigns: OmnichannelCampaign[]) {
  writeStore(CAMPAIGNS_KEY, campaigns);
}

export function loadCronJobs() {
  return readStore<OmnichannelCronJob[]>(CRON_JOBS_KEY, DEFAULT_CRON_JOBS);
}

export function saveCronJobs(jobs: OmnichannelCronJob[]) {
  writeStore(CRON_JOBS_KEY, jobs);
}
