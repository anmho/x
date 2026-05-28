'use client';

import { useEffect, useRef, useState } from 'react';
import { Bot, ChevronRight, Clock, Loader2, Play, TerminalSquare, XCircle, CheckCircle2 } from 'lucide-react';
import { AppNav } from '@/app/_components/app-nav';
import { AgentRun, createRun, listRuns, watchRun, RunStatus, RuntimeProvider } from '@/lib/agent-runs';

function StatusBadge({ status }: { status: RunStatus }) {
  const map: Record<RunStatus, { label: string; className: string; icon: React.ReactNode }> = {
    PENDING: { label: 'Pending', className: 'bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300', icon: <Clock className="w-3 h-3" /> },
    RUNNING: { label: 'Running', className: 'bg-sky-100 text-sky-700 dark:bg-sky-950 dark:text-sky-300', icon: <Loader2 className="w-3 h-3 animate-spin" /> },
    SUCCEEDED: { label: 'Succeeded', className: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300', icon: <CheckCircle2 className="w-3 h-3" /> },
    FAILED: { label: 'Failed', className: 'bg-rose-100 text-rose-700 dark:bg-rose-950 dark:text-rose-300', icon: <XCircle className="w-3 h-3" /> },
  };
  const { label, className, icon } = map[status] ?? map.PENDING;
  return (
    <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${className}`}>
      {icon}{label}
    </span>
  );
}

function RuntimeBadge({ provider }: { provider: RuntimeProvider }) {
  return (
    <span className="inline-flex items-center rounded-full bg-amber-100 px-2 py-0.5 text-xs font-medium text-amber-700 dark:bg-amber-950 dark:text-amber-300">
      {provider}
    </span>
  );
}

export default function AgentsPage() {
  const [runs, setRuns] = useState<AgentRun[]>([]);
  const [prompt, setPrompt] = useState('');
  const [provider, setProvider] = useState<RuntimeProvider>('claude');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selected, setSelected] = useState<string | null>(null);
  const [output, setOutput] = useState<string>('');
  const [streaming, setStreaming] = useState(false);
  const outputRef = useRef<HTMLPreElement>(null);
  const abortRef = useRef<AbortController | null>(null);

  const refresh = async () => {
    try {
      const data = await listRuns(50);
      setRuns(data);
    } catch {
      // ignore transient fetch errors
    }
  };

  useEffect(() => {
    refresh();
    const id = setInterval(refresh, 5000);
    return () => clearInterval(id);
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!prompt.trim()) return;
    setSubmitting(true);
    setError(null);
    try {
      const run = await createRun(prompt.trim(), provider);
      setPrompt('');
      setRuns((prev) => [run, ...prev]);
      openRun(run.id);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create run');
    } finally {
      setSubmitting(false);
    }
  };

  const openRun = (id: string) => {
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;
    setSelected(id);
    setOutput('');
    setStreaming(true);

    void (async () => {
      try {
        for await (const event of watchRun(id, controller.signal)) {
          if (event.type === 'snapshot') {
            setRuns((prev) => prev.map((run) => (run.id === event.run.id ? event.run : run)));
            continue;
          }
          if (event.type === 'output') {
            setOutput((prev) => prev + event.chunk);
            setTimeout(() => outputRef.current?.scrollTo(0, outputRef.current.scrollHeight), 0);
            continue;
          }
          if (event.type === 'completed') {
            setStreaming(false);
            await refresh();
            return;
          }
        }
      } catch {
        if (!controller.signal.aborted) {
          setStreaming(false);
        }
      }
    })();
  };

  useEffect(() => () => abortRef.current?.abort(), []);

  const selectedRun = runs.find((r) => r.id === selected);

  return (
    <div className="min-h-screen bg-zinc-50 text-zinc-950 dark:bg-zinc-950 dark:text-zinc-50">
      <AppNav section="agents" />
      <main className="max-w-7xl mx-auto px-4 py-8">
        <div className="mb-6">
          <h1 className="text-2xl font-bold flex items-center gap-2">
            <Bot className="w-6 h-6" /> Agent Runs
          </h1>
          <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">Submit a prompt and run it with Claude or Codex using the same declarative runtime contract.</p>
        </div>

        <form onSubmit={handleSubmit} className="mb-6 grid gap-3 lg:grid-cols-[auto_1fr_auto]">
          <div className="inline-flex rounded-lg border border-zinc-200 bg-white p-1 dark:border-zinc-800 dark:bg-zinc-900">
            {(['claude', 'codex'] as RuntimeProvider[]).map((value) => (
              <button
                key={value}
                type="button"
                onClick={() => setProvider(value)}
                className={`rounded-md px-3 py-2 text-sm font-medium capitalize transition ${provider === value ? 'bg-zinc-950 text-white dark:bg-zinc-100 dark:text-zinc-950' : 'text-zinc-600 hover:text-zinc-950 dark:text-zinc-400 dark:hover:text-zinc-100'}`}
              >
                {value}
              </button>
            ))}
          </div>
          <input
            className="w-full rounded-lg border border-zinc-300 bg-white px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-sky-500 dark:border-zinc-700 dark:bg-zinc-900"
            placeholder="Enter a prompt for the agent…"
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            disabled={submitting}
          />
          <button
            type="submit"
            disabled={submitting || !prompt.trim()}
            className="inline-flex items-center justify-center gap-2 rounded-lg bg-sky-600 px-4 py-2 text-sm font-medium text-white hover:bg-sky-700 disabled:opacity-50"
          >
            {submitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4" />}
            Run
          </button>
        </form>
        {error && <p className="mb-4 text-sm text-rose-600 dark:text-rose-400">{error}</p>}

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          <div className="overflow-hidden rounded-xl border border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-900">
            <div className="flex items-center justify-between border-b border-zinc-100 px-4 py-3 dark:border-zinc-800">
              <span className="text-sm font-semibold text-zinc-700 dark:text-zinc-200">Recent Runs</span>
              <span className="text-xs text-zinc-400">{runs.length} total</span>
            </div>
            <ul className="max-h-[60vh] divide-y divide-zinc-100 overflow-y-auto dark:divide-zinc-800">
              {runs.length === 0 && (
                <li className="px-4 py-8 text-center text-sm text-zinc-400">No runs yet. Submit a prompt above.</li>
              )}
              {runs.map((run) => (
                <li
                  key={run.id}
                  onClick={() => openRun(run.id)}
                  className={`flex cursor-pointer items-start gap-3 px-4 py-3 hover:bg-zinc-50 dark:hover:bg-zinc-800/60 ${selected === run.id ? 'bg-sky-50 dark:bg-sky-950/30' : ''}`}
                >
                  <TerminalSquare className="mt-0.5 h-4 w-4 shrink-0 text-zinc-400" />
                  <div className="flex-1 min-w-0">
                    <p className="truncate text-sm text-zinc-800 dark:text-zinc-100">{run.prompt}</p>
                    <div className="mt-1 flex items-center gap-2 text-xs text-zinc-400">
                      <span>{new Date(run.created_at).toLocaleString()}</span>
                      <RuntimeBadge provider={run.runtime.provider} />
                    </div>
                  </div>
                  <div className="flex items-center gap-1 shrink-0">
                    <StatusBadge status={run.status} />
                    <ChevronRight className="h-3 w-3 text-zinc-300 dark:text-zinc-600" />
                  </div>
                </li>
              ))}
            </ul>
          </div>

          <div className="flex flex-col overflow-hidden rounded-xl border border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-900">
            <div className="flex items-center justify-between border-b border-zinc-100 px-4 py-3 dark:border-zinc-800">
              <span className="text-sm font-semibold text-zinc-700 dark:text-zinc-200">
                {selectedRun ? `Output — ${selectedRun.runtime.provider} / ${selectedRun.id.slice(0, 8)}` : 'Output'}
              </span>
              {streaming && (
                <span className="flex items-center gap-1 text-xs text-sky-600 dark:text-sky-400">
                  <Loader2 className="w-3 h-3 animate-spin" /> streaming
                </span>
              )}
              {selectedRun && !streaming && <StatusBadge status={selectedRun.status} />}
            </div>
            <pre
              ref={outputRef}
              className="max-h-[60vh] flex-1 overflow-auto bg-zinc-950 p-4 font-mono text-xs whitespace-pre-wrap text-emerald-300"
            >
              {output || (selected ? 'Waiting for output…' : 'Select a run to view output.')}
            </pre>
          </div>
        </div>
      </main>
    </div>
  );
}
