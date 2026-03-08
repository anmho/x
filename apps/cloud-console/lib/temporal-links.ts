const DEFAULT_BASE = 'http://localhost:8233';
const DEFAULT_NAMESPACE = 'default';

export function getTemporalBaseUrl(): string {
  const base = process.env.NEXT_PUBLIC_TEMPORAL_UI_URL || DEFAULT_BASE;
  return base.replace(/\/$/, '');
}

export function buildTemporalNamespacesUrl(): string {
  return `${getTemporalBaseUrl()}/namespaces`;
}

export function buildTemporalWorkflowHistoryUrl(
  namespace?: string | null,
  workflowId?: string | null,
  runId?: string | null,
): string | null {
  if (!workflowId) return null;
  const ns = namespace || process.env.NEXT_PUBLIC_TEMPORAL_NAMESPACE || DEFAULT_NAMESPACE;
  const base = getTemporalBaseUrl();
  if (runId) {
    return `${base}/namespaces/${encodeURIComponent(ns)}/workflows/${encodeURIComponent(workflowId)}/${encodeURIComponent(runId)}/history`;
  }
  return `${base}/namespaces/${encodeURIComponent(ns)}/workflows/${encodeURIComponent(workflowId)}`;
}
