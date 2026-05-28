'use client';

import { useEffect, useMemo, useState } from 'react';
import { CheckCircle2, Clock3, ExternalLink, RefreshCw, Search, Workflow, XCircle } from 'lucide-react';
import { AppNav } from '@/app/_components/app-nav';
import { HeadlessSelect } from '@/app/_components/headless-select';
import { PageIntro } from '@/app/_components/blocks';
import { CATALOG_PROJECTS } from '@/lib/project-catalog';
import { buildTemporalNamespacesUrl, buildTemporalWorkflowHistoryUrl } from '@/lib/temporal-links';
import { WorkflowListItem, listTemporalWorkflows } from '@/lib/temporal-workflows';

const STATUS_OPTIONS = [
  { value: 'all', label: 'All statuses' },
  { value: 'running', label: 'Running' },
  { value: 'completed', label: 'Completed' },
  { value: 'failed', label: 'Failed' },
  { value: 'timed_out', label: 'Timed out' },
  { value: 'canceled', label: 'Canceled' },
  { value: 'terminated', label: 'Terminated' },
];

const ENV_OPTIONS = [
  { value: 'production', label: 'Production' },
  { value: 'staging', label: 'Staging' },
  { value: 'development', label: 'Development' },
];

function statusBadge(status: string) {
  switch (status) {
    case 'completed':
      return <span className="inline-flex items-center gap-1 rounded-full border border-emerald-500/30 bg-emerald-500/10 px-2 py-0.5 text-xs text-emerald-300"><CheckCircle2 className="h-3 w-3" />completed</span>;
    case 'failed':
      return <span className="inline-flex items-center gap-1 rounded-full border border-rose-500/30 bg-rose-500/10 px-2 py-0.5 text-xs text-rose-300"><XCircle className="h-3 w-3" />failed</span>;
    case 'running':
      return <span className="inline-flex items-center gap-1 rounded-full border border-blue-500/30 bg-blue-500/10 px-2 py-0.5 text-xs text-blue-300"><Clock3 className="h-3 w-3 animate-spin" />running</span>;
    default:
      return <span className="inline-flex items-center gap-1 rounded-full border border-amber-500/30 bg-amber-500/10 px-2 py-0.5 text-xs text-amber-300"><Clock3 className="h-3 w-3" />{status}</span>;
  }
}

function formatDate(iso?: string) {
  if (!iso) return '—';
  return new Date(iso).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' });
}

export default function WorkflowsPage() {
  const defaultProjectId = CATALOG_PROJECTS[0]?.id || 'proj_notifications';
  const [projectId, setProjectId] = useState(defaultProjectId);
  const [environment, setEnvironment] = useState('production');
  const [status, setStatus] = useState('all');
  const [query, setQuery] = useState('');

  const [items, setItems] = useState<WorkflowListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const projectOptions = useMemo(
    () => CATALOG_PROJECTS.map((project) => ({ value: project.id, label: project.label, description: project.name })),
    [],
  );

  async function load() {
    try {
      setLoading(true);
      const response = await listTemporalWorkflows({
        projectId,
        environment,
        statuses: status === 'all' ? [] : [status],
        query: query.trim(),
        pageSize: 50,
      });
      setItems(response.items || []);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load workflows');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [projectId, environment, status]);

  const stats = useMemo(() => {
    return {
      total: items.length,
      running: items.filter((item) => item.status === 'running').length,
      completed: items.filter((item) => item.status === 'completed').length,
      failed: items.filter((item) => item.status === 'failed').length,
    };
  }, [items]);

  return (
    <AppNav active="workflows">
      <main className="mx-auto max-w-6xl px-4 py-8 sm:px-6">
        <div className="flex items-end justify-between gap-4">
          <PageIntro
            title="Workflows"
            description="Temporal workflow executions grouped by project and environment."
          />
          <div className="flex items-center gap-2">
            <a
              href={buildTemporalNamespacesUrl()}
              target="_blank"
              rel="noreferrer"
              className="inline-flex shrink-0 items-center gap-2 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm hover:bg-zinc-800 transition-colors"
            >
              <ExternalLink className="h-3.5 w-3.5" />
              Open Temporal Namespaces
            </a>
            <button
              onClick={() => void load()}
              className="inline-flex shrink-0 items-center gap-2 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm hover:bg-zinc-800 transition-colors"
            >
              <RefreshCw className="h-3.5 w-3.5" />
              Refresh
            </button>
          </div>
        </div>

        <div className="mt-4 grid grid-cols-1 gap-3 md:grid-cols-4">
          <HeadlessSelect value={projectId} onValueChange={setProjectId} options={projectOptions} ariaLabel="Project" />
          <HeadlessSelect value={environment} onValueChange={setEnvironment} options={ENV_OPTIONS} ariaLabel="Environment" />
          <HeadlessSelect value={status} onValueChange={setStatus} options={STATUS_OPTIONS} ariaLabel="Status" />
          <form
            onSubmit={(event) => {
              event.preventDefault();
              void load();
            }}
            className="flex items-center gap-2 rounded-md border border-zinc-700 bg-zinc-900 px-3"
          >
            <Search className="h-4 w-4 text-zinc-500" />
            <input
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder="Workflow ID"
              className="h-10 w-full bg-transparent text-sm text-zinc-100 outline-none placeholder:text-zinc-500"
            />
          </form>
        </div>

        <div className="mt-6 grid grid-cols-4 gap-3">
          {[
            { label: 'Total', value: stats.total },
            { label: 'Running', value: stats.running },
            { label: 'Completed', value: stats.completed },
            { label: 'Failed', value: stats.failed },
          ].map((s) => (
            <div key={s.label} className="rounded-xl border border-zinc-800 bg-zinc-950 p-4">
              <p className="text-xs uppercase tracking-wide text-zinc-500">{s.label}</p>
              <p className="mt-2 text-2xl font-semibold">{s.value}</p>
            </div>
          ))}
        </div>

        {error && (
          <p className="mt-4 rounded-md border border-rose-500/30 bg-rose-500/10 px-3 py-2 text-sm text-rose-300">{error}</p>
        )}

        <div className="mt-6 rounded-xl border border-zinc-800 bg-zinc-950">
          <div className="grid grid-cols-[2fr_1fr_1fr_1fr_1fr] gap-4 border-b border-zinc-800 px-4 py-3 text-xs font-medium uppercase tracking-wide text-zinc-500">
            <span>Workflow ID</span>
            <span>Type</span>
            <span>Status</span>
            <span>Namespace</span>
            <span>Started</span>
          </div>

          {loading ? (
            <div className="space-y-2 p-4">
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="h-10 animate-pulse rounded-md bg-zinc-900" />
              ))}
            </div>
          ) : items.length === 0 ? (
            <div className="px-4 py-14 text-center">
              <Workflow className="mx-auto h-10 w-10 text-zinc-700" />
              <p className="mt-3 text-sm text-zinc-500">No workflow executions found for this filter.</p>
            </div>
          ) : (
            <div className="divide-y divide-zinc-800">
              {items.map((item) => {
                const url = item.temporalUiUrl || buildTemporalWorkflowHistoryUrl(item.namespace, item.workflowId, item.runId);
                return (
                  <article key={`${item.workflowId}:${item.runId}`} className="grid grid-cols-[2fr_1fr_1fr_1fr_1fr] items-center gap-4 px-4 py-3 hover:bg-zinc-900/40 transition-colors">
                    <div className="min-w-0">
                      {url ? (
                        <a
                          href={url}
                          target="_blank"
                          rel="noreferrer"
                          className="flex items-center gap-1.5 font-mono text-xs text-blue-400 hover:text-blue-300 truncate"
                        >
                          <ExternalLink className="h-3 w-3 shrink-0" />
                          <span className="truncate">{item.workflowId}</span>
                        </a>
                      ) : (
                        <span className="font-mono text-xs text-zinc-400 truncate block">{item.workflowId}</span>
                      )}
                      {item.runId && (
                        <p className="mt-0.5 font-mono text-[11px] text-zinc-600 truncate">run: {item.runId}</p>
                      )}
                    </div>
                    <p className="truncate text-sm text-zinc-300">{item.workflowType || 'n/a'}</p>
                    <div>{statusBadge(item.status)}</div>
                    <p className="truncate text-sm text-zinc-400">{item.namespace}</p>
                    <p className="text-xs text-zinc-500">{formatDate(item.startTime)}</p>
                  </article>
                );
              })}
            </div>
          )}
        </div>
      </main>
    </AppNav>
  );
}
