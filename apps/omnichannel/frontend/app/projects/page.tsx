import Link from 'next/link';
import { BarChart3, CreditCard, FolderKanban, Network, Rocket } from 'lucide-react';
import { AppNav } from '@/app/_components/app-nav';
import { PageIntro } from '@/app/_components/blocks';
import { CATALOG_PROJECTS } from '@/lib/project-catalog';

export default function ProjectsPage() {
  return (
    <AppNav active="projects">
      <main className="mx-auto max-w-6xl px-4 py-8 sm:px-6">
        <PageIntro
          title="Projects"
          description="Project-level view with deployment counts, observability, product analytics, and billing."
        />

        <div className="mt-6 grid grid-cols-1 gap-4 lg:grid-cols-2">
          {CATALOG_PROJECTS.map((project) => (
            <article id={project.id} key={project.id} className="rounded-xl border border-zinc-800 bg-zinc-950 p-5">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="inline-flex items-center gap-2 text-base font-medium text-zinc-100">
                    <FolderKanban className="h-4 w-4 text-zinc-400" />
                    {project.label}
                  </p>
                  <p className="mt-1 text-sm text-zinc-400">{project.description}</p>
                </div>
                <span className="rounded-full border border-zinc-700 px-2 py-0.5 text-xs text-zinc-300">
                  {project.environment}
                </span>
              </div>

              <div className="mt-4 grid grid-cols-2 gap-3">
                <div className="rounded-md border border-zinc-800 bg-zinc-900/40 p-3">
                  <p className="text-xs uppercase tracking-wide text-zinc-500">Deployments</p>
                  <p className="mt-1 inline-flex items-center gap-1.5 text-lg font-semibold text-zinc-100">
                    <Rocket className="h-4 w-4 text-zinc-400" />
                    {project.deployments.length}
                  </p>
                </div>
                <Link
                  href="/applications"
                  className="rounded-md border border-zinc-800 bg-zinc-900/40 p-3 transition hover:bg-zinc-900"
                >
                  <p className="text-xs uppercase tracking-wide text-zinc-500">Applications</p>
                  <p className="mt-1 inline-flex items-center gap-1.5 text-sm font-medium text-zinc-200">
                    <Network className="h-4 w-4 text-zinc-400" />
                    {project.applications.length} app{project.applications.length === 1 ? '' : 's'}
                  </p>
                </Link>
              </div>

              <div className="mt-4 flex flex-wrap gap-2">
                <a
                  href={project.grafanaUrl}
                  target="_blank"
                  rel="noreferrer"
                  className="inline-flex items-center gap-1 rounded-md border border-zinc-700 px-2.5 py-1.5 text-xs text-zinc-300 hover:bg-zinc-900"
                >
                  <BarChart3 className="h-3.5 w-3.5" />
                  Grafana
                </a>
                <a
                  href={project.posthogUrl}
                  target="_blank"
                  rel="noreferrer"
                  className="inline-flex items-center gap-1 rounded-md border border-zinc-700 px-2.5 py-1.5 text-xs text-zinc-300 hover:bg-zinc-900"
                >
                  <BarChart3 className="h-3.5 w-3.5" />
                  PostHog
                </a>
                <a
                  href={project.stripeUrl}
                  target="_blank"
                  rel="noreferrer"
                  className="inline-flex items-center gap-1 rounded-md border border-zinc-700 px-2.5 py-1.5 text-xs text-zinc-300 hover:bg-zinc-900"
                >
                  <CreditCard className="h-3.5 w-3.5" />
                  Stripe
                </a>
              </div>
            </article>
          ))}
        </div>
      </main>
    </AppNav>
  );
}
