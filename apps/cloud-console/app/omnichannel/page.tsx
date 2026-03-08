'use client';

import { useEffect, useMemo, useState } from 'react';
import {
  BellRing,
  Calendar,
  FileText,
  Megaphone,
  Network,
  Rocket,
  Send,
  Smartphone,
  Workflow,
} from 'lucide-react';
import { AppNav } from '@/app/_components/app-nav';
import { PageIntro, ServiceTile } from '@/app/_components/blocks';
import { notificationApi, templateApi } from '@/lib/api';
import { loadCampaigns, loadCronJobs, loadRecipients } from '@/lib/omnichannel-store';

const OMNICHANNEL_TILES = [
  {
    href: '/omnichannel/test',
    title: 'Test',
    description: 'Run channel tests before launching customer-facing sends.',
    icon: BellRing,
    iconClass: 'bg-blue-500/15 text-blue-300 border-blue-500/30',
  },
  {
    href: '/templates',
    title: 'Templates',
    description: 'Create and manage reusable notification templates.',
    icon: FileText,
    iconClass: 'bg-amber-500/15 text-amber-300 border-amber-500/30',
  },
  {
    href: '/deployments',
    title: 'Status',
    description: 'Inspect delivery pipelines, workflow activity, and health.',
    icon: Rocket,
    iconClass: 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30',
  },
  {
    href: '/omnichannel/recipients',
    title: 'Recipients',
    description: 'Maintain recipient lists and group-level targeting.',
    icon: Network,
    iconClass: 'bg-cyan-500/15 text-cyan-300 border-cyan-500/30',
  },
  {
    href: '/omnichannel/cron-jobs',
    title: 'Cron Jobs',
    description: 'Configure recurring schedules and run jobs on demand.',
    icon: Calendar,
    iconClass: 'bg-fuchsia-500/15 text-fuchsia-300 border-fuchsia-500/30',
  },
  {
    href: '/omnichannel/campaigns',
    title: 'Campaigns',
    description: 'Manage and launch campaigns against live recipient segments.',
    icon: Megaphone,
    iconClass: 'bg-orange-500/15 text-orange-300 border-orange-500/30',
  },
  {
    href: '/omnichannel/emulator',
    title: 'App Emulator',
    description: 'Inspect app/push payloads delivered through the local relay.',
    icon: Smartphone,
    iconClass: 'bg-sky-500/15 text-sky-300 border-sky-500/30',
  },
  {
    href: '/workflows',
    title: 'Workflow Timeline',
    description: 'Review Temporal workflow runs and execution states.',
    icon: Workflow,
    iconClass: 'bg-violet-500/15 text-violet-300 border-violet-500/30',
  },
  {
    href: '/notifications/create',
    title: 'Send Notification',
    description: 'Manual send/test form with live channel previews.',
    icon: Send,
    iconClass: 'bg-lime-500/15 text-lime-300 border-lime-500/30',
  },
];

export default function OmnichannelPage() {
  const [notificationsCount, setNotificationsCount] = useState(0);
  const [templatesCount, setTemplatesCount] = useState(0);
  const [recipientsCount, setRecipientsCount] = useState(0);
  const [campaignsCount, setCampaignsCount] = useState(0);
  const [cronJobsCount, setCronJobsCount] = useState(0);

  useEffect(() => {
    const recipients = loadRecipients();
    const campaigns = loadCampaigns();
    const jobs = loadCronJobs();
    setRecipientsCount(recipients.length);
    setCampaignsCount(campaigns.length);
    setCronJobsCount(jobs.length);

    void loadBackendStats();
  }, []);

  async function loadBackendStats() {
    try {
      const [notifications, templates] = await Promise.all([
        notificationApi.list(),
        templateApi.list(),
      ]);
      setNotificationsCount(notifications.data?.length || 0);
      setTemplatesCount(templates.data?.length || 0);
    } catch {
      setNotificationsCount(0);
      setTemplatesCount(0);
    }
  }

  const metrics = useMemo(
    () => [
      { label: 'Notifications', value: notificationsCount },
      { label: 'Templates', value: templatesCount },
      { label: 'Recipients', value: recipientsCount },
      { label: 'Campaigns', value: campaignsCount },
      { label: 'Cron Jobs', value: cronJobsCount },
    ],
    [notificationsCount, templatesCount, recipientsCount, campaignsCount, cronJobsCount],
  );

  return (
    <div className="min-h-screen app-shell">
      <AppNav active="omnichannel" />

      <main className="mx-auto max-w-6xl px-4 py-8 sm:px-6">
        <PageIntro
          title="Omnichannel"
          description="Operate testing, templates, status, recipients, schedules, and campaigns from one control surface."
        />

        <section className="mt-5 grid grid-cols-2 gap-3 md:grid-cols-5">
          {metrics.map((item) => (
            <article key={item.label} className="rounded-lg border border-zinc-800 bg-zinc-950 p-3">
              <p className="text-xs uppercase tracking-wide text-zinc-500">{item.label}</p>
              <p className="mt-1 text-xl font-semibold text-zinc-100">{item.value}</p>
            </article>
          ))}
        </section>

        <div className="mt-6 grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {OMNICHANNEL_TILES.map((service) => (
            <ServiceTile
              key={service.href}
              href={service.href}
              title={service.title}
              description={service.description}
              icon={service.icon}
              iconClass={service.iconClass}
            />
          ))}
        </div>
      </main>
    </div>
  );
}
