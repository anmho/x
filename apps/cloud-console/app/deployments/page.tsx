'use client';

import Link from 'next/link';
import { FormEvent, useEffect, useMemo, useState } from 'react';
import {
  AlertTriangle,
  BarChart3,
  Calendar,
  CheckCircle2,
  Clock3,
  Copy,
  CreditCard,
  ExternalLink,
  FolderKanban,
  KeyRound,
  Network,
  Plus,
  RefreshCw,
  Rocket,
  Send,
  Server,
  Shield,
  Workflow,
  XCircle,
} from 'lucide-react';
import { Notification, NotificationChannel, notificationApi } from '@/lib/api';
import { AccessKey, mintAccessKey, revokeAccessKey, listAccessKeys } from '@/lib/access-keys';
import { AppNav } from '@/app/_components/app-nav';
import { MetricCard } from '@/app/_components/blocks';
import { HeadlessSelect } from '@/app/_components/headless-select';
import { CATALOG_DEPLOYMENTS, CATALOG_PROJECTS, catalogProjectCounts } from '@/lib/project-catalog';
import { buildTemporalNamespacesUrl, buildTemporalWorkflowHistoryUrl } from '@/lib/temporal-links';

type Project = {
  id: string;
  name: string;
  role: 'main' | 'tooling';
  description: string;
};

type Deployment = {
  id: string;
  projectId: string;
  name: string;
  env: 'production' | 'staging' | 'development';
  region: string;
  status: 'healthy' | 'degraded';
};

type ApiKeyRecord = {
  id: string;
  projectId: string;
  deploymentId: string;
  name: string;
  prefix: string;
  scopes: string[];
  status: 'active' | 'rotated' | 'revoked';
  createdAt: string;
};

const PROJECTS_STORAGE_KEY = 'notifications:projects';
const DEPLOYMENTS_STORAGE_KEY = 'notifications:deployments';
const API_KEYS_STORAGE_KEY = 'notifications:keys';

const DEFAULT_PROJECTS: Project[] = CATALOG_PROJECTS.map((project) => ({
  id: project.id,
  name: project.name,
  role: project.role,
  description: project.description,
}));

const DEFAULT_DEPLOYMENTS: Deployment[] = CATALOG_DEPLOYMENTS.map((deployment) => ({
  id: deployment.id,
  projectId: deployment.projectId,
  name: deployment.name,
  env: deployment.env,
  region: deployment.region,
  status: deployment.status,
}));

const DEFAULT_KEYS: ApiKeyRecord[] = [
  {
    id: 'key_local',
    projectId: 'proj_notifications',
    deploymentId: 'dep_notifications_api_prod',
    name: 'local-dev-key',
    prefix: 'test-api-key',
    scopes: ['notifications:write'],
    status: 'active',
    createdAt: new Date().toISOString(),
  },
];

const SCOPES = ['notifications:read', 'notifications:write', 'notifications:admin'];

function formatDate(date: string) {
  return new Date(date).toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  });
}

function readStored<T>(key: string, fallback: T): T {
  try {
    const value = window.localStorage.getItem(key);
    if (!value) return fallback;
    return JSON.parse(value) as T;
  } catch {
    return fallback;
  }
}

function mapEnvironment(env: Deployment['env']): 'dev' | 'staging' | 'prod' {
  if (env === 'production') return 'prod';
  if (env === 'staging') return 'staging';
  return 'dev';
}

function mapAccessKeyToRecord(key: AccessKey, projects: Project[], deployments: Deployment[]): ApiKeyRecord {
  const matchedProject = projects.find((project) => project.name === key.application) ?? projects[0];
  const matchedDeployment = deployments.find((deployment) => deployment.projectId === matchedProject.id) ?? deployments[0];
  return {
    id: key.id,
    projectId: matchedProject?.id || 'proj_notifications',
    deploymentId: matchedDeployment?.id || 'dep_notifications_api_prod',
    name: key.owner,
    prefix: key.key_prefix,
    scopes: key.service_scopes.map((scope) => `${scope.service}:${scope.scope}`),
    status: key.status,
    createdAt: key.created_at,
  };
}

export default function NotificationsServicePage() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [projects, setProjects] = useState<Project[]>(DEFAULT_PROJECTS);
  const [deployments, setDeployments] = useState<Deployment[]>(DEFAULT_DEPLOYMENTS);
  const [keys, setKeys] = useState<ApiKeyRecord[]>(DEFAULT_KEYS);

  const [selectedProjectId, setSelectedProjectId] = useState('proj_notifications');
  const [newKeyName, setNewKeyName] = useState('');
  const [newKeyDeploymentId, setNewKeyDeploymentId] = useState('dep_notifications_api_prod');
  const [newScopes, setNewScopes] = useState<string[]>(['notifications:write']);

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [createdKey, setCreatedKey] = useState<string | null>(null);

  const [testChannel, setTestChannel] = useState<NotificationChannel>('email');
  const [testRecipient, setTestRecipient] = useState('user@example.com');
  const [testSubject, setTestSubject] = useState('Notifications channel test');
  const [testBody, setTestBody] = useState('<p>Test from cloud console</p>');
  const [testing, setTesting] = useState(false);

  useEffect(() => {
    if (typeof window !== 'undefined') {
      setProjects(readStored<Project[]>(PROJECTS_STORAGE_KEY, DEFAULT_PROJECTS));
      setDeployments(readStored<Deployment[]>(DEPLOYMENTS_STORAGE_KEY, DEFAULT_DEPLOYMENTS));
      setKeys(readStored<ApiKeyRecord[]>(API_KEYS_STORAGE_KEY, DEFAULT_KEYS));
    }

    void refreshNotifications();
    void refreshKeys();
  }, []);

  useEffect(() => {
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(PROJECTS_STORAGE_KEY, JSON.stringify(projects));
      window.localStorage.setItem(DEPLOYMENTS_STORAGE_KEY, JSON.stringify(deployments));
      window.localStorage.setItem(API_KEYS_STORAGE_KEY, JSON.stringify(keys));
    }
  }, [projects, deployments, keys]);

  const selectedProject = useMemo(
    () => projects.find((project) => project.id === selectedProjectId) ?? projects[0],
    [projects, selectedProjectId],
  );

  const projectDeployments = useMemo(
    () => deployments.filter((deployment) => deployment.projectId === selectedProjectId),
    [deployments, selectedProjectId],
  );

  const projectKeys = useMemo(
    () => keys.filter((key) => key.projectId === selectedProjectId),
    [keys, selectedProjectId],
  );

  const deploymentCountsByProject = useMemo(() => {
    const counts: Record<string, number> = {};
    for (const deployment of deployments) {
      counts[deployment.projectId] = (counts[deployment.projectId] || 0) + 1;
    }
    return counts;
  }, [deployments]);

  const selectedProjectCatalog = useMemo(
    () => CATALOG_PROJECTS.find((project) => project.id === selectedProjectId),
    [selectedProjectId],
  );

  useEffect(() => {
    if (projectDeployments.length === 0) {
      setNewKeyDeploymentId('');
      return;
    }
    if (!projectDeployments.some((deployment) => deployment.id === newKeyDeploymentId)) {
      setNewKeyDeploymentId(projectDeployments[0].id);
    }
  }, [projectDeployments, newKeyDeploymentId]);

  async function refreshNotifications() {
    try {
      setLoading(true);
      const response = await notificationApi.list();
      setNotifications(response.data ?? []);
      setError(null);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to load notification activity';
      setError(message);
    } finally {
      setLoading(false);
    }
  }

  async function refreshKeys() {
    try {
      const data = await listAccessKeys();
      setKeys(data.map((key) => mapAccessKeyToRecord(key, projects, deployments)));
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to load keys';
      setError(message);
      setSuccess(null);
    }
  }

  const stats = useMemo(() => {
    const total = notifications.length;
    const pending = notifications.filter((item) => item.status === 'pending' || item.status === 'processing').length;
    const failed = notifications.filter((item) => item.status === 'failed').length;
    return { total, pending, failed };
  }, [notifications]);

  function statusIcon(status: Notification['status']) {
    switch (status) {
      case 'sent':
        return <CheckCircle2 className="h-3.5 w-3.5 text-emerald-300" />;
      case 'failed':
        return <XCircle className="h-3.5 w-3.5 text-rose-300" />;
      case 'processing':
        return <Clock3 className="h-3.5 w-3.5 animate-spin text-blue-300" />;
      default:
        return <Clock3 className="h-3.5 w-3.5 text-amber-300" />;
    }
  }

  function statusClass(status: Notification['status']) {
    switch (status) {
      case 'sent':
        return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-300';
      case 'failed':
        return 'border-rose-500/30 bg-rose-500/10 text-rose-300';
      case 'processing':
        return 'border-blue-500/30 bg-blue-500/10 text-blue-300';
      case 'pending':
        return 'border-amber-500/30 bg-amber-500/10 text-amber-300';
      default:
        return 'border-zinc-700 bg-zinc-900 text-zinc-300';
    }
  }

  function toggleScope(scope: string) {
    setNewScopes((prev) => (prev.includes(scope) ? prev.filter((item) => item !== scope) : [...prev, scope]));
  }

  async function createKey(event: FormEvent) {
    event.preventDefault();
    if (!newKeyName.trim() || newScopes.length === 0 || !newKeyDeploymentId) return;

    const deployment = deployments.find((item) => item.id === newKeyDeploymentId);
    if (!deployment || !selectedProject) {
      setError('Select a valid deployment');
      return;
    }

    try {
      const minted = await mintAccessKey({
        application: selectedProject.name,
        environment: mapEnvironment(deployment.env),
        owner: newKeyName.trim(),
        service_scopes: newScopes.map((entry) => {
          const [service, scope] = entry.split(':');
          return { service, scope };
        }),
      });
      const mapped = mapAccessKeyToRecord(minted, projects, deployments);
      setKeys((prev) => [mapped, ...prev]);
      setNewKeyName('');
      setNewScopes(['notifications:write']);
      setCreatedKey(minted.key || null);
      setSuccess(`Created API key ${mapped.name}`);
      setError(null);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to mint key';
      setError(message);
      setSuccess(null);
    }
  }

  async function revokeKey(id: string) {
    try {
      const revoked = await revokeAccessKey(id);
      const mapped = mapAccessKeyToRecord(revoked, projects, deployments);
      setKeys((prev) => prev.map((entry) => (entry.id === id ? mapped : entry)));
      setSuccess('Key revoked');
      setError(null);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to revoke key';
      setError(message);
      setSuccess(null);
    }
  }

  async function copyText(value: string) {
    try {
      await navigator.clipboard.writeText(value);
      setSuccess('Copied');
      setError(null);
    } catch {
      setError('Clipboard write failed');
      setSuccess(null);
    }
  }

  async function submitChannelTest(event: FormEvent) {
    event.preventDefault();

    try {
      setTesting(true);
      await notificationApi.create({
        channel: testChannel,
        recipient: testRecipient,
        recipient_email: testRecipient,
        subject: testSubject,
        body: testBody,
        metadata: {
          source: 'services.notifications',
          tool_project: 'console',
          target_project: selectedProject?.name || 'notifications',
        },
      });
      setSuccess(`Submitted ${testChannel} test`);
      setError(null);
      await refreshNotifications();
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to submit channel test';
      setError(message);
      setSuccess(null);
    } finally {
      setTesting(false);
    }
  }

  return (
    <div className="min-h-screen app-shell">
      <AppNav active="omnichannelStatus" />

      <main className="mx-auto max-w-6xl px-4 py-6 sm:px-6">
        <div className="mb-4 flex items-center justify-between gap-3">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Notifications</h1>
            <p className="mt-1 text-xs text-zinc-500">
              Main project: <span className="font-medium text-zinc-300">notifications</span> • Tooling project: <span className="font-medium text-zinc-300">console</span>
            </p>
          </div>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => void refreshNotifications()}
              className="inline-flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-200 hover:bg-zinc-800"
            >
              <RefreshCw className="h-3.5 w-3.5" />
              Refresh
            </button>
            <Link
              href="/notifications/create"
              className="inline-flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-200 hover:bg-zinc-800"
            >
              <Send className="h-3.5 w-3.5" />
              Send
            </Link>
            <Link
              href="/api-keys"
              className="inline-flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-200 hover:bg-zinc-800"
            >
              <KeyRound className="h-3.5 w-3.5" />
              API Keys
            </Link>
          </div>
        </div>

        <div className="mb-4 grid grid-cols-1 gap-3 sm:grid-cols-3">
          <MetricCard title="Total" value={stats.total.toString()} />
          <MetricCard title="Pending" value={stats.pending.toString()} />
          <MetricCard title="Failed" value={stats.failed.toString()} />
        </div>

        {(error || success) && (
          <div className="mb-4 space-y-2">
            {error && (
              <div className="rounded-md border border-rose-500/30 bg-rose-500/10 px-3 py-2 text-sm text-rose-300">
                <span className="inline-flex items-center gap-1">
                  <AlertTriangle className="h-4 w-4" />
                  {error}
                </span>
              </div>
            )}
            {success && <div className="rounded-md border border-emerald-500/30 bg-emerald-500/10 px-3 py-2 text-sm text-emerald-300">{success}</div>}
          </div>
        )}

        <section className="mb-4 rounded-xl border border-zinc-800 bg-zinc-950 p-4">
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-[180px_1fr] sm:items-center">
            <label className="text-xs uppercase tracking-wide text-zinc-500">Active project</label>
            <HeadlessSelect
              value={selectedProjectId}
              onValueChange={setSelectedProjectId}
              options={projects.map((project) => ({
                value: project.id,
                label: project.name,
                description: `${catalogProjectCounts(project.id).applications} apps / ${deploymentCountsByProject[project.id] || 0} deployments`,
              }))}
              ariaLabel="Select project"
            />
          </div>
          {selectedProject && <p className="mt-2 text-sm text-zinc-400">{selectedProject.description}</p>}
          {selectedProjectCatalog && (
            <p className="mt-1 text-xs text-zinc-500">
              Associated apps: {selectedProjectCatalog.applications.map((app) => app.name).join(', ') || 'none'}
            </p>
          )}
        </section>

        {selectedProjectCatalog && (
          <section className="mb-4">
            <Link
              href={`/projects#${selectedProjectCatalog.id}`}
              className="block rounded-xl border border-zinc-800 bg-zinc-950 p-5 transition hover:border-zinc-700 hover:bg-zinc-900/60"
            >
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="inline-flex items-center gap-2 text-base font-medium text-zinc-100">
                    <FolderKanban className="h-4 w-4 text-zinc-400" />
                    {selectedProjectCatalog.label}
                  </p>
                  <p className="mt-1 text-sm text-zinc-400">{selectedProjectCatalog.description}</p>
                </div>
                <span className="rounded-full border border-zinc-700 px-2 py-0.5 text-xs text-zinc-300">
                  {selectedProjectCatalog.environment}
                </span>
              </div>

              <div className="mt-4 grid grid-cols-2 gap-3">
                <div className="rounded-md border border-zinc-800 bg-zinc-900/40 p-3">
                  <p className="text-xs uppercase tracking-wide text-zinc-500">Deployments</p>
                  <p className="mt-1 inline-flex items-center gap-1.5 text-lg font-semibold text-zinc-100">
                    <Rocket className="h-4 w-4 text-zinc-400" />
                    {projectDeployments.length}
                  </p>
                </div>
                <div className="rounded-md border border-zinc-800 bg-zinc-900/40 p-3">
                  <p className="text-xs uppercase tracking-wide text-zinc-500">Applications</p>
                  <p className="mt-1 inline-flex items-center gap-1.5 text-sm font-medium text-zinc-200">
                    <Network className="h-4 w-4 text-zinc-400" />
                    {selectedProjectCatalog.applications.length} app
                    {selectedProjectCatalog.applications.length === 1 ? '' : 's'}
                  </p>
                </div>
              </div>

              <div className="mt-4 flex flex-wrap gap-2">
                <span className="inline-flex items-center gap-1 rounded-md border border-zinc-700 px-2.5 py-1.5 text-xs text-zinc-300">
                  <BarChart3 className="h-3.5 w-3.5" />
                  Grafana
                </span>
                <span className="inline-flex items-center gap-1 rounded-md border border-zinc-700 px-2.5 py-1.5 text-xs text-zinc-300">
                  <BarChart3 className="h-3.5 w-3.5" />
                  PostHog
                </span>
                <span className="inline-flex items-center gap-1 rounded-md border border-zinc-700 px-2.5 py-1.5 text-xs text-zinc-300">
                  <CreditCard className="h-3.5 w-3.5" />
                  Stripe
                </span>
              </div>

              <p className="mt-3 text-xs text-zinc-500">Click to open the full project overview.</p>
            </Link>
          </section>
        )}

        <section className="mb-4 rounded-xl border border-zinc-800 bg-zinc-950">
          <div className="flex items-center justify-between border-b border-zinc-800 px-4 py-3">
            <h2 className="inline-flex items-center gap-2 text-sm font-medium text-zinc-200">
              <Server className="h-4 w-4" />
              Deployments
            </h2>
            <span className="text-xs text-zinc-500">{selectedProject?.name} / {projectDeployments.length} services</span>
          </div>
          <div className="divide-y divide-zinc-800">
            {projectDeployments.map((deployment) => (
              <article key={deployment.id} className="flex items-center justify-between px-4 py-3">
                <div>
                  <p className="text-sm font-medium text-zinc-100">{deployment.name}</p>
                  <p className="text-xs text-zinc-500">
                    {deployment.env} • {deployment.region}
                  </p>
                </div>
                <span
                  className={`rounded-full border px-2 py-0.5 text-xs ${
                    deployment.status === 'healthy'
                      ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-300'
                      : 'border-amber-500/30 bg-amber-500/10 text-amber-300'
                  }`}
                >
                  {deployment.status}
                </span>
                {deployment.name.includes('temporal') && (
                  <a
                    href={buildTemporalNamespacesUrl()}
                    target="_blank"
                    rel="noreferrer"
                    className="inline-flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-900 px-2 py-1 text-xs text-zinc-300 hover:bg-zinc-800"
                  >
                    <ExternalLink className="h-3.5 w-3.5" />
                    Workflow Deployment
                  </a>
                )}
              </article>
            ))}
          </div>
        </section>

        <section className="mb-4 grid grid-cols-1 gap-4 lg:grid-cols-2">
          <article className="rounded-xl border border-zinc-800 bg-zinc-950">
            <div className="border-b border-zinc-800 px-4 py-3">
              <h2 className="inline-flex items-center gap-2 text-sm font-medium text-zinc-200">
                <Shield className="h-4 w-4" /> Project API Keys
              </h2>
            </div>
            <div className="space-y-3 p-4">
              <form onSubmit={createKey} className="space-y-2">
                <input
                  value={newKeyName}
                  onChange={(event) => setNewKeyName(event.target.value)}
                  placeholder="key name"
                  className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
                />
                <div className="rounded-md border border-zinc-800 bg-zinc-900 px-3 py-2 text-xs text-zinc-400">
                  Project scope: <span className="text-zinc-200">{selectedProject?.name}</span>
                </div>
                <HeadlessSelect
                  value={newKeyDeploymentId}
                  onValueChange={setNewKeyDeploymentId}
                  options={projectDeployments.map((deployment) => ({
                    value: deployment.id,
                    label: `Deployment: ${deployment.name}`,
                    description: `${deployment.env} • ${deployment.region}`,
                  }))}
                  ariaLabel="Select deployment"
                  placeholder="Select deployment"
                  disabled={projectDeployments.length === 0}
                />
                <div className="flex flex-wrap gap-1.5">
                  {SCOPES.map((scope) => (
                    <label
                      key={scope}
                      className={`cursor-pointer rounded-full border px-2 py-1 text-xs ${
                        newScopes.includes(scope)
                          ? 'border-zinc-600 bg-zinc-800 text-zinc-100'
                          : 'border-zinc-700 bg-zinc-900 text-zinc-400'
                      }`}
                    >
                      <input
                        type="checkbox"
                        className="hidden"
                        checked={newScopes.includes(scope)}
                        onChange={() => toggleScope(scope)}
                      />
                      {scope}
                    </label>
                  ))}
                </div>
                <button
                  type="submit"
                  className="inline-flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm hover:bg-zinc-800"
                  disabled={!newKeyDeploymentId}
                >
                  <Plus className="h-3.5 w-3.5" /> Create key
                </button>
              </form>

              {createdKey && (
                <div className="rounded-md border border-emerald-500/30 bg-emerald-500/10 p-3 text-xs text-emerald-300">
                  <p>Copy this key now:</p>
                  <code className="mt-1 block text-zinc-100">{createdKey}</code>
                </div>
              )}

              <div className="space-y-2">
                {projectKeys.map((entry) => (
                  <article key={entry.id} className="rounded-md border border-zinc-800 bg-zinc-900/50 p-3">
                    <div className="flex items-start justify-between gap-2">
                      <div>
                        <p className="text-sm font-medium">{entry.name}</p>
                        <p className="text-xs text-zinc-500">
                          {deployments.find((deployment) => deployment.id === entry.deploymentId)?.name ?? entry.deploymentId}
                        </p>
                        <p className="mt-1 font-mono text-xs text-zinc-300">{entry.prefix}••••••••</p>
                      </div>
                      <span className={`rounded-full px-2 py-0.5 text-xs ${entry.status === 'active' ? 'bg-emerald-500/20 text-emerald-300' : 'bg-rose-500/20 text-rose-300'}`}>
                        {entry.status}
                      </span>
                    </div>
                    <div className="mt-2 flex flex-wrap gap-1">
                      {entry.scopes.map((scope) => (
                        <span key={scope} className="rounded-full border border-zinc-700 bg-zinc-900 px-2 py-0.5 text-xs text-zinc-300">
                          {scope}
                        </span>
                      ))}
                    </div>
                    <div className="mt-2 flex gap-3 text-xs">
                      <button type="button" onClick={() => void copyText(`${entry.prefix}••••••••`)} className="inline-flex items-center gap-1 text-zinc-300 hover:text-zinc-100">
                        <Copy className="h-3 w-3" /> Copy
                      </button>
                      {entry.status === 'active' && (
                        <button type="button" onClick={() => revokeKey(entry.id)} className="inline-flex items-center gap-1 text-rose-300 hover:text-rose-200">
                          <KeyRound className="h-3 w-3" /> Revoke
                        </button>
                      )}
                    </div>
                  </article>
                ))}
              </div>
            </div>
          </article>

          <article className="rounded-xl border border-zinc-800 bg-zinc-950">
            <div className="border-b border-zinc-800 px-4 py-3">
              <h2 className="inline-flex items-center gap-2 text-sm font-medium text-zinc-200">
                <Plus className="h-4 w-4" /> Omnichannel Test Tool
              </h2>
              <p className="mt-1 text-xs text-zinc-500">Hosted by project: <span className="text-zinc-300">console</span></p>
            </div>
            <form onSubmit={submitChannelTest} className="space-y-2 p-4">
              <HeadlessSelect
                value={testChannel}
                onValueChange={(value) => setTestChannel(value as NotificationChannel)}
                options={[
                  { value: 'email', label: 'email' },
                  { value: 'sms', label: 'sms' },
                  { value: 'push', label: 'push' },
                  { value: 'webhook', label: 'webhook' },
                  { value: 'app', label: 'app (emulator)' },
                  { value: 'imessage', label: 'imessage (emulator)' },
                ]}
                ariaLabel="Select channel"
              />
              <input
                value={testRecipient}
                onChange={(event) => setTestRecipient(event.target.value)}
                placeholder="recipient"
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
              />
              <input
                value={testSubject}
                onChange={(event) => setTestSubject(event.target.value)}
                placeholder="subject"
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
              />
              <textarea
                value={testBody}
                onChange={(event) => setTestBody(event.target.value)}
                rows={4}
                className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
              />
              <button
                type="submit"
                disabled={testing}
                className="inline-flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm hover:bg-zinc-800 disabled:opacity-40"
              >
                <Send className="h-3.5 w-3.5" /> {testing ? 'Submitting...' : 'Run test'}
              </button>
            </form>
          </article>
        </section>

        <section className="mb-4 rounded-xl border border-zinc-800 bg-zinc-950">
          <div className="border-b border-zinc-800 px-4 py-3">
            <h2 className="inline-flex items-center gap-2 text-sm font-medium text-zinc-200">
              <Workflow className="h-4 w-4" /> Temporal Workflows
            </h2>
          </div>
          {loading ? (
            <div className="space-y-2 p-4">
              {Array.from({ length: 4 }).map((_, i) => (
                <div key={i} className="h-10 animate-pulse rounded-md bg-zinc-900" />
              ))}
            </div>
          ) : notifications.length === 0 ? (
            <div className="px-4 py-10 text-sm text-zinc-500">No workflow executions yet.</div>
          ) : (
            <div className="divide-y divide-zinc-800">
              {notifications.slice(0, 8).map((notification) => (
                <article key={notification.id} className="px-4 py-3">
                  <div className="flex items-center justify-between gap-2">
                    <p className="truncate text-sm text-zinc-200">{notification.subject}</p>
                    <span className={`inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs ${statusClass(notification.status)}`}>
                      {statusIcon(notification.status)}
                      {notification.status}
                    </span>
                  </div>
                  {buildTemporalWorkflowHistoryUrl(notification.temporal_workflow_id, notification.temporal_run_id) ? (
                    <>
                      <p className="mt-1 truncate font-mono text-xs text-zinc-500">
                        wf:{' '}
                        <a
                          href={buildTemporalWorkflowHistoryUrl(notification.temporal_workflow_id, notification.temporal_run_id) || '#'}
                          target="_blank"
                          rel="noreferrer"
                          className="underline decoration-zinc-700 underline-offset-2 hover:text-zinc-300"
                        >
                          {notification.temporal_workflow_id}
                        </a>
                      </p>
                      <p className="truncate font-mono text-xs text-zinc-600">
                        run:{' '}
                        <a
                          href={buildTemporalWorkflowHistoryUrl(notification.temporal_workflow_id, notification.temporal_run_id) || '#'}
                          target="_blank"
                          rel="noreferrer"
                          className="underline decoration-zinc-800 underline-offset-2 hover:text-zinc-400"
                        >
                          {notification.temporal_run_id || 'n/a'}
                        </a>
                      </p>
                    </>
                  ) : (
                    <>
                      <p className="mt-1 truncate font-mono text-xs text-zinc-500">wf: {notification.temporal_workflow_id || 'n/a'}</p>
                      <p className="truncate font-mono text-xs text-zinc-600">run: {notification.temporal_run_id || 'n/a'}</p>
                    </>
                  )}
                </article>
              ))}
            </div>
          )}
        </section>

        <section className="rounded-xl border border-zinc-800 bg-zinc-950">
          <div className="border-b border-zinc-800 px-4 py-3">
            <h2 className="text-sm font-medium text-zinc-200">Recent Notifications</h2>
          </div>

          {loading ? (
            <div className="space-y-2 p-4">
              {Array.from({ length: 6 }).map((_, i) => (
                <div key={i} className="h-10 animate-pulse rounded-md bg-zinc-900" />
              ))}
            </div>
          ) : notifications.length === 0 ? (
            <div className="px-4 py-12 text-center text-sm text-zinc-400">No notifications yet.</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-zinc-800">
                <thead>
                  <tr>
                    <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-zinc-500">Recipient</th>
                    <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-zinc-500">Subject</th>
                    <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-zinc-500">Status</th>
                    <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-zinc-500">Workflow</th>
                    <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-zinc-500">Created</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800">
                  {notifications.slice(0, 25).map((notification) => (
                    <tr key={notification.id} className="hover:bg-zinc-900/40">
                      <td className="px-4 py-3 text-sm text-zinc-200">{notification.recipient_email}</td>
                      <td className="max-w-[420px] px-4 py-3 text-sm text-zinc-300">
                        <div className="truncate">{notification.subject}</div>
                      </td>
                      <td className="px-4 py-3 text-sm">
                        <span className={`inline-flex items-center gap-1 rounded-full border px-2.5 py-1 text-xs ${statusClass(notification.status)}`}>
                          {statusIcon(notification.status)}
                          {notification.status}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-xs text-zinc-400">
                        {buildTemporalWorkflowHistoryUrl(notification.temporal_workflow_id, notification.temporal_run_id) ? (
                          <a
                            href={buildTemporalWorkflowHistoryUrl(notification.temporal_workflow_id, notification.temporal_run_id) || '#'}
                            target="_blank"
                            rel="noreferrer"
                            className="inline-flex items-center gap-1 text-zinc-300 underline decoration-zinc-700 underline-offset-2 hover:text-zinc-100"
                          >
                            <ExternalLink className="h-3.5 w-3.5" />
                            Open
                          </a>
                        ) : (
                          'n/a'
                        )}
                      </td>
                      <td className="px-4 py-3 text-sm text-zinc-500">
                        <span className="inline-flex items-center gap-1.5">
                          <Calendar className="h-3.5 w-3.5" />
                          {formatDate(notification.created_at)}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      </main>
    </div>
  );
}
