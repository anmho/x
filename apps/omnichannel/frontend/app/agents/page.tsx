import { Bot, Play } from 'lucide-react';
import { AppNav } from '@/app/_components/app-nav';
import { PageIntro } from '@/app/_components/blocks';

type AgentStatus = 'running' | 'idle' | 'error';

type Agent = {
  id: string;
  name: string;
  description: string;
  status: AgentStatus;
  lastRun: string;
  runsToday: number;
  secrets: string[];
  workflow?: string;
};

const AGENTS: Agent[] = [
  {
    id: 'agent_deploy',
    name: 'deploy-agent',
    description: 'Orchestrates Cloud Run deployments via Temporal workflows on push to main.',
    status: 'idle',
    lastRun: '2026-03-05T14:22:00Z',
    runsToday: 3,
    secrets: ['GCP_SERVICE_ACCOUNT_KEY'],
    workflow: 'deploy-workflow',
  },
  {
    id: 'agent_monitor',
    name: 'monitor-agent',
    description: 'Polls health endpoints and dispatches failure alerts via the omnichannel API.',
    status: 'running',
    lastRun: '2026-03-06T09:01:00Z',
    runsToday: 12,
    secrets: ['SENDGRID_API_KEY'],
    workflow: 'notification-workflow',
  },
];

function statusLabel(status: AgentStatus) {
  switch (status) {
    case 'running': return <span className="rounded-full border border-emerald-500/30 bg-emerald-500/10 px-2 py-0.5 text-xs text-emerald-300">Running</span>;
    case 'error': return <span className="rounded-full border border-rose-500/30 bg-rose-500/10 px-2 py-0.5 text-xs text-rose-300">Error</span>;
    default: return <span className="rounded-full border border-zinc-700 bg-zinc-900 px-2 py-0.5 text-xs text-zinc-400">Idle</span>;
  }
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' });
}

export default function AgentsPage() {
  return (
    <AppNav active="agents">
      <main className="mx-auto max-w-4xl px-4 py-8 sm:px-6">
        <PageIntro
          title="Agents"
          description="Autonomous agents running workflows and integrations on your behalf."
        />

        <div className="mt-6 space-y-4">
          {AGENTS.map((agent) => (
            <article key={agent.id} className="rounded-xl border border-zinc-800 bg-zinc-950 p-5">
              <div className="flex items-start justify-between gap-4">
                <div className="flex items-start gap-3">
                  <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border border-zinc-800 bg-zinc-900">
                    <Bot className="h-4 w-4 text-zinc-400" />
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <p className="font-mono text-sm font-semibold text-zinc-100">{agent.name}</p>
                      {statusLabel(agent.status)}
                    </div>
                    <p className="mt-1 text-sm text-zinc-400">{agent.description}</p>
                  </div>
                </div>

                <div className="flex shrink-0 items-center gap-2">
                  <button className="inline-flex items-center gap-1.5 rounded-md border border-zinc-700 bg-zinc-900 px-2.5 py-1.5 text-xs text-zinc-300 hover:bg-zinc-800 transition-colors">
                    <Play className="h-3.5 w-3.5" />
                    Run now
                  </button>
                </div>
              </div>

              <div className="mt-4 grid grid-cols-3 gap-3 text-xs">
                <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-3">
                  <p className="text-zinc-500 uppercase tracking-wide">Last run</p>
                  <p className="mt-1 text-zinc-300">{formatDate(agent.lastRun)}</p>
                </div>
                <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-3">
                  <p className="text-zinc-500 uppercase tracking-wide">Runs today</p>
                  <p className="mt-1 text-zinc-300">{agent.runsToday}</p>
                </div>
                <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-3">
                  <p className="text-zinc-500 uppercase tracking-wide">Workflow</p>
                  <p className="mt-1 text-zinc-300">{agent.workflow ?? 'none'}</p>
                </div>
              </div>

              {agent.secrets.length > 0 && (
                <div className="mt-3 flex flex-wrap items-center gap-1.5">
                  <p className="text-xs text-zinc-600">Secrets:</p>
                  {agent.secrets.map((s) => (
                    <span key={s} className="rounded border border-zinc-800 bg-zinc-900 px-2 py-0.5 font-mono text-[11px] text-zinc-400">
                      {s}
                    </span>
                  ))}
                </div>
              )}
            </article>
          ))}
        </div>
      </main>
    </AppNav>
  );
}
