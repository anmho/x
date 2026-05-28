import { Blocks, BellRing, Calendar, FolderKanban } from 'lucide-react';
import { AppNav } from './_components/app-nav';
import { PageIntro, ServiceTile } from './_components/blocks';

const SERVICES = [
  {
    href: '/omnichannel',
    title: 'Omnichannel',
    description: 'Test notifications, manage templates, monitor status, and run campaigns.',
    icon: Blocks,
    iconClass: 'bg-violet-500/15 text-violet-300 border-violet-500/30',
  },
  {
    href: '/cron-jobs',
    title: 'Cron Jobs',
    description: 'Manage recurring schedules and run jobs on demand.',
    icon: Calendar,
    iconClass: 'bg-fuchsia-500/15 text-fuchsia-300 border-fuchsia-500/30',
  },
  {
    href: '/applications',
    title: 'Applications',
    description: 'View deployed apps, environments, and quick links for API key operations.',
    icon: BellRing,
    iconClass: 'bg-blue-500/15 text-blue-300 border-blue-500/30',
  },
  {
    href: '/projects',
    title: 'Projects',
    description: 'Track project-level ownership, observability, billing, and deployment health.',
    icon: FolderKanban,
    iconClass: 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30',
  },
];

export default function HomePage() {
  return (
    <AppNav active="home">
      <main className="mx-auto max-w-6xl px-4 py-8 sm:px-6">
        <PageIntro
          title="Home"
          description="Cloud console entry point for services, applications, and projects."
        />

        <div className="mt-6 grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
          {SERVICES.map((service) => (
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
