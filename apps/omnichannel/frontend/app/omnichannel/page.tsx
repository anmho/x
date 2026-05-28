'use client';

import { useEffect, useMemo, useState } from 'react';
import { KeyRound, Megaphone, Rocket, Send } from 'lucide-react';
import { AppNav } from '@/app/_components/app-nav';
import { PageIntro, ServiceTile } from '@/app/_components/blocks';
import { notificationApi } from '@/lib/api';
import { listAccessKeys } from '@/lib/access-keys';
import { loadCampaigns } from '@/lib/omnichannel-store';

const OMNICHANNEL_TILES = [
  {
    href: '/notifications/create',
    title: 'Send Notification',
    description: 'Manual send flow with channel previews.',
    icon: Send,
    iconClass: 'bg-lime-500/15 text-lime-300 border-lime-500/30',
  },
  {
    href: '/api-keys',
    title: 'Manage API Keys',
    description: 'Create, rotate, and revoke scoped service keys.',
    icon: KeyRound,
    iconClass: 'bg-amber-500/15 text-amber-300 border-amber-500/30',
  },
  {
    href: '/deployments',
    title: 'Notification Status',
    description: 'Inspect delivery health and open workflows for Temporal details.',
    icon: Rocket,
    iconClass: 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30',
  },
  {
    href: '/omnichannel/campaigns',
    title: 'Campaigns',
    description: 'Manage and launch campaigns against live recipient segments.',
    icon: Megaphone,
    iconClass: 'bg-orange-500/15 text-orange-300 border-orange-500/30',
  },
];

export default function OmnichannelPage() {
  const [notificationsCount, setNotificationsCount] = useState(0);
  const [apiKeysCount, setApiKeysCount] = useState(0);
  const [campaignsCount] = useState(() => loadCampaigns().length);
  const [workflowLinksCount, setWorkflowLinksCount] = useState(0);

  async function loadBackendStats() {
    try {
      const [notifications, keys] = await Promise.all([
        notificationApi.list(),
        listAccessKeys(),
      ]);
      const notificationRows = notifications.data ?? [];
      setNotificationsCount(notificationRows.length);
      setWorkflowLinksCount(notificationRows.filter((item) => Boolean(item.temporal_workflow_id)).length);
      setApiKeysCount(keys.length);
    } catch {
      setNotificationsCount(0);
      setWorkflowLinksCount(0);
      setApiKeysCount(0);
    }
  }

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void loadBackendStats();
    }, 0);
    return () => window.clearTimeout(timer);
  }, []);

  const metrics = useMemo(
    () => [
      { label: 'Notifications', value: notificationsCount },
      { label: 'API Keys', value: apiKeysCount },
      { label: 'Campaigns', value: campaignsCount },
      { label: 'Workflow Links', value: workflowLinksCount },
    ],
    [notificationsCount, apiKeysCount, campaignsCount, workflowLinksCount],
  );

  return (
    <AppNav active="omnichannel">
      <main className="mx-auto max-w-6xl px-4 py-8 sm:px-6">
        <PageIntro
          title="Omnichannel"
          description="Operate core messaging flows: send notifications, manage keys, monitor status, and run campaigns."
        />

        <section className="mt-5 grid grid-cols-2 gap-3 md:grid-cols-4">
          {metrics.map((item) => (
            <article key={item.label} className="rounded-lg border border-zinc-800 bg-zinc-950 p-3">
              <p className="text-xs uppercase tracking-wide text-zinc-500">{item.label}</p>
              <p className="mt-1 text-xl font-semibold text-zinc-100">{item.value}</p>
            </article>
          ))}
        </section>

        <div className="mt-6 grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
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
    </AppNav>
  );
}
