'use client';

import { useEffect, useState } from 'react';
import { BarChart3, Check, CheckCircle2, Cloud, CreditCard, Key, Link2, Rocket, Settings, Shield, Unlink } from 'lucide-react';
import { AppNav } from '@/app/_components/app-nav';

type Tab = 'general' | 'integrations' | 'security';

type GCloudStatus = 'disconnected' | 'connecting' | 'connected';

const GCLOUD_SCOPES = [
  'https://www.googleapis.com/auth/cloud-platform',
  'https://www.googleapis.com/auth/cloudkms',
  'https://www.googleapis.com/auth/secretmanager',
  'https://www.googleapis.com/auth/run.admin',
];

const ADDITIONAL_INTEGRATIONS: {
  id: 'posthog' | 'stripe' | 'vercel';
  name: string;
  description: string;
  icon: React.ReactNode;
  actionLabel: string;
  note: string;
}[] = [
  {
    id: 'posthog',
    name: 'PostHog',
    description: 'Capture product analytics, session replay, and feature-flag telemetry in one place.',
    icon: <BarChart3 className="h-5 w-5 text-amber-300" />,
    actionLabel: 'PostHog coming soon',
    note: 'Connection flow will be exposed at /api/auth/posthog/connect when provider auth is enabled.',
  },
  {
    id: 'stripe',
    name: 'Stripe',
    description: 'Sync billing events, payment health, and subscription status with platform projects.',
    icon: <CreditCard className="h-5 w-5 text-indigo-300" />,
    actionLabel: 'Stripe coming soon',
    note: 'Billing provider setup will be available after Stripe account linking is released.',
  },
  {
    id: 'vercel',
    name: 'Vercel',
    description: 'Authorize deployments and environment sync with Vercel projects managed by this workspace.',
    icon: <Rocket className="h-5 w-5 text-cyan-300" />,
    actionLabel: 'Vercel coming soon',
    note: 'Vercel OAuth activation will unlock deployment/project linking from settings.',
  },
];

export default function SettingsPage() {
  const [tab, setTab] = useState<Tab>('general');
  const [gcloudStatus, setGcloudStatus] = useState<GCloudStatus>('disconnected');
  const [gcloudEmail, setGcloudEmail] = useState('');
  const [gcloudProject, setGcloudProject] = useState('');
  const [connectError, setConnectError] = useState<string | null>(null);

  // Read tab from URL param
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const t = params.get('tab');
    if (t === 'integrations' || t === 'security' || t === 'general') setTab(t);
  }, []);

  function handleConnectGCloud() {
    setGcloudStatus('connecting');
    setConnectError(null);
    // In production: redirect to /api/auth/google/connect
    // For now, simulate with a timeout
    setTimeout(() => {
      setGcloudStatus('connected');
      setGcloudEmail('andrew@example.com');
      setGcloudProject('my-gcp-project');
    }, 1800);
  }

  function handleDisconnect() {
    if (!confirm('Disconnect Google Cloud? The control plane will lose access to GCP resources.')) return;
    setGcloudStatus('disconnected');
    setGcloudEmail('');
    setGcloudProject('');
  }

  const tabs: { id: Tab; label: string; icon: React.ReactNode }[] = [
    { id: 'general', label: 'General', icon: <Settings className="h-4 w-4" /> },
    { id: 'integrations', label: 'Integrations', icon: <Link2 className="h-4 w-4" /> },
    { id: 'security', label: 'Security', icon: <Shield className="h-4 w-4" /> },
  ];

  return (
    <AppNav active="settings">
      <main className="mx-auto max-w-4xl px-4 py-8 sm:px-6">
        <div className="mb-8">
          <h1 className="text-2xl font-semibold tracking-tight">Settings</h1>
          <p className="mt-1 text-sm text-zinc-400">Manage your account, integrations, and security preferences.</p>
        </div>

        <div className="flex gap-8">
          {/* Sidebar tabs */}
          <nav className="w-48 shrink-0 space-y-0.5">
            {tabs.map((t) => (
              <button
                key={t.id}
                onClick={() => setTab(t.id)}
                className={`flex w-full items-center gap-2.5 rounded-lg px-3 py-2 text-sm transition-colors ${
                  tab === t.id
                    ? 'bg-zinc-800 text-zinc-100'
                    : 'text-zinc-400 hover:bg-zinc-900 hover:text-zinc-200'
                }`}
              >
                {t.icon}
                {t.label}
              </button>
            ))}
          </nav>

          {/* Content */}
          <div className="flex-1 min-w-0 space-y-6">
            {tab === 'general' && <GeneralTab />}
            {tab === 'integrations' && (
              <IntegrationsTab
                gcloudStatus={gcloudStatus}
                gcloudEmail={gcloudEmail}
                gcloudProject={gcloudProject}
                connectError={connectError}
                onConnect={handleConnectGCloud}
                onDisconnect={handleDisconnect}
              />
            )}
            {tab === 'security' && <SecurityTab />}
          </div>
        </div>
      </main>
    </AppNav>
  );
}

function GeneralTab() {
  return (
    <div className="space-y-6">
      <section className="rounded-xl border border-zinc-800 bg-zinc-950 p-5">
        <h2 className="mb-4 text-sm font-semibold text-zinc-200">Profile</h2>
        <div className="flex items-center gap-4">
          <div className="flex h-14 w-14 items-center justify-center rounded-full bg-blue-600 text-lg font-semibold text-white">
            AH
          </div>
          <div>
            <p className="font-medium text-zinc-100">Andrew Ho</p>
            <p className="text-sm text-zinc-500">andrew@example.com</p>
          </div>
        </div>
        <div className="mt-5 space-y-3 text-sm">
          <label className="block">
            <span className="mb-1.5 block text-xs font-medium text-zinc-400">Display name</span>
            <input
              defaultValue="Andrew Ho"
              className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-zinc-100 focus:border-zinc-500 focus:outline-none"
            />
          </label>
          <label className="block">
            <span className="mb-1.5 block text-xs font-medium text-zinc-400">Email</span>
            <input
              defaultValue="andrew@example.com"
              type="email"
              className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-zinc-400 focus:border-zinc-500 focus:outline-none"
              disabled
            />
          </label>
        </div>
        <div className="mt-4 flex justify-end">
          <button className="rounded-md bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-500 transition-colors">
            Save changes
          </button>
        </div>
      </section>

      <section className="rounded-xl border border-zinc-800 bg-zinc-950 p-5">
        <h2 className="mb-1 text-sm font-semibold text-zinc-200">Preferences</h2>
        <p className="mb-4 text-xs text-zinc-500">UI and notification preferences.</p>
        <div className="space-y-3 text-sm">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-zinc-200">Theme</p>
              <p className="text-xs text-zinc-500">Choose Light, Dark, or System from the avatar quick settings menu.</p>
            </div>
            <span className="rounded-md border border-zinc-700 bg-zinc-900 px-2 py-1 text-xs text-zinc-400">Quick settings</span>
          </div>
        </div>
      </section>
    </div>
  );
}

function IntegrationsTab({
  gcloudStatus,
  gcloudEmail,
  gcloudProject,
  connectError,
  onConnect,
  onDisconnect,
}: {
  gcloudStatus: GCloudStatus;
  gcloudEmail: string;
  gcloudProject: string;
  connectError: string | null;
  onConnect: () => void;
  onDisconnect: () => void;
}) {
  return (
    <div className="space-y-4">
      {/* Google Cloud */}
      <section className="rounded-xl border border-zinc-800 bg-zinc-950 p-5">
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-start gap-3">
            {/* Google Cloud icon */}
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border border-zinc-800 bg-zinc-900">
              <svg viewBox="0 0 24 24" className="h-5 w-5" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M12.34 9.27L14.8 6.81C14.17 6.29 13.38 6 12.5 6C10.57 6 9 7.57 9 9.5C9 9.67 9.01 9.83 9.04 10H6.04C6.01 9.84 6 9.67 6 9.5C6 5.91 8.91 3 12.5 3C14.32 3 15.96 3.74 17.14 4.93L15.55 6.52C15.16 6.2 14.73 5.96 14.27 5.81" fill="#4285F4"/>
                <path d="M17.96 9.5C17.96 9.83 17.93 10.16 17.87 10.47H15.04C15.1 10.16 15.13 9.83 15.13 9.5C15.13 8.63 14.83 7.83 14.33 7.19L15.92 5.6C17.18 6.76 17.96 8.05 17.96 9.5Z" fill="#EA4335"/>
                <path d="M12.5 13C14.43 13 16 11.43 16 9.5H13C13 10.33 12.33 11 11.5 11C11.17 11 10.86 10.89 10.62 10.69L8.17 13.14C8.83 13.69 9.63 14 10.5 14C11.2 14 11.87 13.79 12.42 13.43L12.5 13Z" fill="#34A853"/>
                <path d="M9 9.5C9 8.67 9.31 7.91 9.82 7.34L7.87 5.39C6.72 6.55 6 8.15 6 9.5H9Z" fill="#FBBC05"/>
              </svg>
            </div>
            <div>
              <p className="font-medium text-zinc-100">Google Cloud</p>
              <p className="mt-0.5 text-sm text-zinc-400">
                Allow the control plane to manage GCP resources on your behalf — Cloud Run deployments, Secret Manager, IAM, and more.
              </p>
            </div>
          </div>

          {gcloudStatus === 'connected' ? (
            <span className="inline-flex shrink-0 items-center gap-1 rounded-full border border-emerald-500/30 bg-emerald-500/10 px-2.5 py-1 text-xs text-emerald-300">
              <CheckCircle2 className="h-3 w-3" /> Connected
            </span>
          ) : (
            <span className="inline-flex shrink-0 items-center gap-1 rounded-full border border-zinc-700 bg-zinc-900 px-2.5 py-1 text-xs text-zinc-500">
              Not connected
            </span>
          )}
        </div>

        {gcloudStatus === 'connected' && (
          <div className="mt-4 rounded-lg border border-zinc-800 bg-zinc-900/50 p-4">
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <p className="text-xs text-zinc-500">Account</p>
                <p className="mt-1 text-zinc-200">{gcloudEmail}</p>
              </div>
              <div>
                <p className="text-xs text-zinc-500">GCP Project</p>
                <p className="mt-1 text-zinc-200">{gcloudProject}</p>
              </div>
            </div>

            <div className="mt-4">
              <p className="mb-2 text-xs font-medium text-zinc-500">Granted scopes</p>
              <div className="space-y-1">
                {GCLOUD_SCOPES.map((scope) => (
                  <div key={scope} className="flex items-center gap-2 text-xs text-zinc-400">
                    <Check className="h-3 w-3 shrink-0 text-emerald-400" />
                    <span className="font-mono">{scope}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {connectError && (
          <p className="mt-3 rounded-md border border-rose-500/30 bg-rose-500/10 px-3 py-2 text-sm text-rose-300">{connectError}</p>
        )}

        <div className="mt-4 flex gap-2">
          {gcloudStatus !== 'connected' ? (
            <button
              onClick={onConnect}
              disabled={gcloudStatus === 'connecting'}
              className="inline-flex items-center gap-2 rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-500 disabled:opacity-60 transition-colors"
            >
              {gcloudStatus === 'connecting' ? (
                <>
                  <div className="h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white" />
                  Connecting…
                </>
              ) : (
                <>
                  <Cloud className="h-4 w-4" />
                  Connect with Google Cloud
                </>
              )}
            </button>
          ) : (
            <button
              onClick={onDisconnect}
              className="inline-flex items-center gap-2 rounded-md border border-zinc-700 bg-zinc-900 px-4 py-2 text-sm text-zinc-300 hover:bg-zinc-800 hover:text-rose-400 transition-colors"
            >
              <Unlink className="h-4 w-4" />
              Disconnect
            </button>
          )}
        </div>

        <div className="mt-4 border-t border-zinc-800 pt-4">
          <p className="text-xs text-zinc-600">
            The OAuth flow redirects to{' '}
            <code className="rounded bg-zinc-900 px-1 py-0.5 text-zinc-500">/api/auth/google/connect</code>{' '}
            and stores a refresh token server-side. Your credentials are never exposed to the browser after initial authorization.
          </p>
        </div>
      </section>

      {ADDITIONAL_INTEGRATIONS.map((integration) => (
        <IntegrationCard
          key={integration.id}
          name={integration.name}
          description={integration.description}
          icon={integration.icon}
          actionLabel={integration.actionLabel}
          note={integration.note}
        />
      ))}
    </div>
  );
}

function IntegrationCard({
  name,
  description,
  icon,
  actionLabel,
  note,
}: {
  name: string;
  description: string;
  icon: React.ReactNode;
  actionLabel: string;
  note: string;
}) {
  return (
    <section className="rounded-xl border border-zinc-800 bg-zinc-950 p-5">
      <div className="flex items-start justify-between gap-4">
        <div className="flex items-start gap-3">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border border-zinc-800 bg-zinc-900">
            {icon}
          </div>
          <div>
            <p className="font-medium text-zinc-100">{name}</p>
            <p className="mt-0.5 text-sm text-zinc-400">{description}</p>
          </div>
        </div>
        <span className="inline-flex shrink-0 items-center gap-1 rounded-full border border-zinc-700 bg-zinc-900 px-2.5 py-1 text-xs text-zinc-500">
          Not connected
        </span>
      </div>

      <div className="mt-4 flex gap-2">
        <button
          disabled
          className="inline-flex cursor-not-allowed items-center gap-2 rounded-md border border-zinc-700 bg-zinc-900 px-4 py-2 text-sm text-zinc-500 opacity-80"
        >
          {actionLabel}
        </button>
      </div>

      <div className="mt-4 border-t border-zinc-800 pt-4">
        <p className="text-xs text-zinc-600">{note}</p>
      </div>
    </section>
  );
}

function SecurityTab() {
  return (
    <div className="space-y-4">
      <section className="rounded-xl border border-zinc-800 bg-zinc-950 p-5">
        <h2 className="mb-4 text-sm font-semibold text-zinc-200">API Keys</h2>
        <p className="text-sm text-zinc-400">
          Personal API keys grant programmatic access to the platform API. Prefer using project-scoped secrets for service-to-service calls.
        </p>
        <div className="mt-4 flex gap-2">
          <a
            href="/api-keys"
            className="inline-flex items-center gap-2 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-200 hover:bg-zinc-800 transition-colors"
          >
            <Key className="h-4 w-4" />
            Manage API keys
          </a>
          <a
            href="/secrets"
            className="inline-flex items-center gap-2 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-200 hover:bg-zinc-800 transition-colors"
          >
            <Shield className="h-4 w-4" />
            Manage secrets
          </a>
        </div>
      </section>

      <section className="rounded-xl border border-zinc-800 bg-zinc-950 p-5">
        <h2 className="mb-1 text-sm font-semibold text-zinc-200">Session</h2>
        <p className="mb-4 text-sm text-zinc-400">You are currently signed in.</p>
        <a
          href="/api/auth/logout"
          className="inline-flex items-center gap-2 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-300 hover:bg-zinc-800 hover:text-rose-400 transition-colors"
        >
          Sign out
        </a>
      </section>
    </div>
  );
}
