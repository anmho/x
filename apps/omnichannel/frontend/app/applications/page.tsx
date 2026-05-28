import Link from 'next/link';
import { AppWindow, ArrowRight, KeyRound, Server } from 'lucide-react';
import { AppNav } from '@/app/_components/app-nav';
import { PageIntro } from '@/app/_components/blocks';
import { CATALOG_APPLICATIONS } from '@/lib/project-catalog';

function statusClass(status: 'healthy' | 'degraded') {
  return status === 'healthy'
    ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-300'
    : 'border-amber-500/30 bg-amber-500/10 text-amber-300';
}

export default function ApplicationsPage() {
  return (
    <AppNav active="applications">
      <main className="mx-auto max-w-6xl px-4 py-8 sm:px-6">
        <PageIntro
          title="Applications"
          description="All deployed apps across environments with fast access to key management."
        />

        <section className="mt-6 rounded-xl border border-zinc-800 bg-zinc-950">
          <div className="grid grid-cols-[2fr_1fr_1fr_1fr_1fr] gap-3 border-b border-zinc-800 px-4 py-3 text-xs uppercase tracking-wide text-zinc-500">
            <span>Name</span>
            <span>Project</span>
            <span>Environment</span>
            <span>Status</span>
            <span className="text-right">Actions</span>
          </div>

          <div className="divide-y divide-zinc-800">
            {CATALOG_APPLICATIONS.map((app) => (
              <article key={app.id} className="grid grid-cols-[2fr_1fr_1fr_1fr_1fr] items-center gap-3 px-4 py-3">
                <div>
                  <p className="inline-flex items-center gap-2 text-sm font-medium text-zinc-100">
                    <AppWindow className="h-4 w-4 text-zinc-400" />
                    {app.name}
                  </p>
                  <p className="mt-0.5 text-xs text-zinc-500">{app.apiKeys} API keys</p>
                </div>
                <p className="text-sm text-zinc-300">{app.projectLabel}</p>
                <p className="text-sm text-zinc-300">{app.environment}</p>
                <p>
                  <span className={`rounded-full border px-2 py-0.5 text-xs ${statusClass(app.status)}`}>
                    {app.status}
                  </span>
                </p>
                <div className="flex justify-end gap-2">
                  <Link
                    href="/api-keys"
                    className="inline-flex items-center gap-1 rounded-md border border-zinc-700 px-2.5 py-1.5 text-xs text-zinc-300 hover:bg-zinc-900"
                  >
                    <KeyRound className="h-3.5 w-3.5" />
                    Keys
                  </Link>
                  <Link
                    href="/deployments"
                    className="inline-flex items-center gap-1 rounded-md border border-zinc-700 px-2.5 py-1.5 text-xs text-zinc-300 hover:bg-zinc-900"
                  >
                    <Server className="h-3.5 w-3.5" />
                    Service
                  </Link>
                </div>
              </article>
            ))}
          </div>
        </section>

        <div className="mt-4 flex justify-end">
          <Link href="/api-keys" className="inline-flex items-center gap-1 text-sm text-zinc-400 hover:text-zinc-200">
            Open API key admin
            <ArrowRight className="h-4 w-4" />
          </Link>
        </div>
      </main>
    </AppNav>
  );
}
