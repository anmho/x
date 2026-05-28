'use client';

import { useState } from 'react';
import { Check, Eye, EyeOff, Key, Plus, Shield, Trash2, X } from 'lucide-react';
import { AppNav } from '@/app/_components/app-nav';
import { PageIntro } from '@/app/_components/blocks';

type SecretTarget = 'application' | 'agent' | 'workflow';

type SecretShare = {
  type: SecretTarget;
  id: string;
  name: string;
};

type Secret = {
  id: string;
  name: string;
  description: string;
  value: string;
  createdAt: string;
  shares: SecretShare[];
};

const INITIAL_SECRETS: Secret[] = [
  {
    id: 'sec_gcp_key',
    name: 'GCP_SERVICE_ACCOUNT_KEY',
    description: 'Google Cloud service account credentials for control plane operations',
    value: '{"type":"service_account","project_id":"my-gcp-project",...}',
    createdAt: '2026-03-01T12:00:00Z',
    shares: [
      { type: 'application', id: 'app_notif_api', name: 'notifications-api' },
      { type: 'workflow', id: 'wf_deploy', name: 'deploy-workflow' },
    ],
  },
  {
    id: 'sec_sendgrid',
    name: 'SENDGRID_API_KEY',
    description: 'SendGrid API key for email delivery',
    value: 'SG.xxxxxxxxxxxxxxxxxxxx',
    createdAt: '2026-03-02T12:00:00Z',
    shares: [
      { type: 'application', id: 'app_notif_worker', name: 'notifications-worker' },
    ],
  },
];

const AVAILABLE_TARGETS: SecretShare[] = [
  { type: 'application', id: 'app_notif_frontend', name: 'notifications-frontend' },
  { type: 'application', id: 'app_notif_api', name: 'notifications-api' },
  { type: 'application', id: 'app_notif_worker', name: 'notifications-worker' },
  { type: 'application', id: 'app_template_studio', name: 'template-studio' },
  { type: 'agent', id: 'agent_deploy', name: 'deploy-agent' },
  { type: 'agent', id: 'agent_monitor', name: 'monitor-agent' },
  { type: 'workflow', id: 'wf_deploy', name: 'deploy-workflow' },
  { type: 'workflow', id: 'wf_notify', name: 'notification-workflow' },
];

function targetBadgeClass(type: SecretTarget) {
  switch (type) {
    case 'application': return 'bg-blue-500/10 text-blue-300 border-blue-500/20';
    case 'agent': return 'bg-violet-500/10 text-violet-300 border-violet-500/20';
    case 'workflow': return 'bg-amber-500/10 text-amber-300 border-amber-500/20';
  }
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
}

export default function SecretsPage() {
  const [secrets, setSecrets] = useState<Secret[]>(INITIAL_SECRETS);
  const [revealedIds, setRevealedIds] = useState<Set<string>>(new Set());
  const [showCreate, setShowCreate] = useState(false);
  const [shareModalId, setShareModalId] = useState<string | null>(null);

  const [newName, setNewName] = useState('');
  const [newDesc, setNewDesc] = useState('');
  const [newValue, setNewValue] = useState('');
  const [selectedTargets, setSelectedTargets] = useState<string[]>([]);

  function toggleReveal(id: string) {
    setRevealedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }

  function deleteSecret(id: string) {
    if (!confirm('Delete this secret? This cannot be undone.')) return;
    setSecrets((prev) => prev.filter((s) => s.id !== id));
  }

  function createSecret(e: React.FormEvent) {
    e.preventDefault();
    const id = `sec_${Date.now()}`;
    const shares = AVAILABLE_TARGETS.filter((t) => selectedTargets.includes(t.id));
    setSecrets((prev) => [
      {
        id,
        name: newName.trim(),
        description: newDesc.trim(),
        value: newValue.trim(),
        createdAt: new Date().toISOString(),
        shares,
      },
      ...prev,
    ]);
    setNewName('');
    setNewDesc('');
    setNewValue('');
    setSelectedTargets([]);
    setShowCreate(false);
  }

  function updateShares(secretId: string, shares: SecretShare[]) {
    setSecrets((prev) => prev.map((s) => (s.id === secretId ? { ...s, shares } : s)));
    setShareModalId(null);
  }

  const shareModalSecret = secrets.find((s) => s.id === shareModalId);

  return (
    <AppNav active="secrets">
      <main className="mx-auto max-w-4xl px-4 py-8 sm:px-6">
        <div className="flex items-end justify-between gap-4">
          <PageIntro
            title="Secrets"
            description="Account-level secrets. Share with specific applications, agents, or workflows."
          />
          <button
            onClick={() => setShowCreate(true)}
            className="inline-flex shrink-0 items-center gap-2 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm font-medium hover:bg-zinc-800 transition-colors"
          >
            <Plus className="h-4 w-4" />
            New secret
          </button>
        </div>

        {/* Info banner */}
        <div className="mt-4 flex items-start gap-3 rounded-xl border border-zinc-800 bg-zinc-950 px-4 py-3">
          <Shield className="mt-0.5 h-4 w-4 shrink-0 text-zinc-500" />
          <p className="text-sm text-zinc-400">
            Secrets are stored at the <span className="font-medium text-zinc-200">account level</span> and are never exposed in logs or build output.
            You control which applications, agents, and workflows have access.
          </p>
        </div>

        {/* Secrets list */}
        <div className="mt-6 space-y-3">
          {secrets.length === 0 && (
            <div className="rounded-xl border border-zinc-800 bg-zinc-950 py-14 text-center">
              <Key className="mx-auto h-10 w-10 text-zinc-700" />
              <p className="mt-3 text-sm text-zinc-400">No secrets yet. Create one to get started.</p>
            </div>
          )}

          {secrets.map((secret) => (
            <article key={secret.id} className="rounded-xl border border-zinc-800 bg-zinc-950 p-4">
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <Key className="h-4 w-4 shrink-0 text-zinc-500" />
                    <p className="font-mono text-sm font-semibold text-zinc-100">{secret.name}</p>
                  </div>
                  {secret.description && (
                    <p className="mt-1 text-sm text-zinc-400">{secret.description}</p>
                  )}

                  {/* Value */}
                  <div className="mt-3 flex items-center gap-2">
                    <code className="flex-1 rounded-md border border-zinc-800 bg-zinc-900 px-3 py-1.5 font-mono text-xs text-zinc-300">
                      {revealedIds.has(secret.id) ? secret.value : '•'.repeat(Math.min(secret.value.length, 40))}
                    </code>
                    <button
                      onClick={() => toggleReveal(secret.id)}
                      className="rounded-md p-1.5 text-zinc-500 hover:bg-zinc-800 hover:text-zinc-300 transition-colors"
                      title={revealedIds.has(secret.id) ? 'Hide' : 'Reveal'}
                    >
                      {revealedIds.has(secret.id) ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                    </button>
                  </div>

                  {/* Shares */}
                  <div className="mt-3 flex flex-wrap items-center gap-1.5">
                    <span className="text-xs text-zinc-600">Shared with:</span>
                    {secret.shares.length === 0 ? (
                      <span className="text-xs text-zinc-600">nobody</span>
                    ) : (
                      secret.shares.map((share) => (
                        <span
                          key={`${share.type}:${share.id}`}
                          className={`rounded-full border px-2 py-0.5 text-[11px] font-medium ${targetBadgeClass(share.type)}`}
                        >
                          {share.name}
                        </span>
                      ))
                    )}
                    <button
                      onClick={() => setShareModalId(secret.id)}
                      className="text-xs text-blue-400 hover:text-blue-300 transition-colors"
                    >
                      Edit access →
                    </button>
                  </div>
                </div>

                <div className="flex shrink-0 items-center gap-1">
                  <p className="mr-2 text-xs text-zinc-600">{formatDate(secret.createdAt)}</p>
                  <button
                    onClick={() => deleteSecret(secret.id)}
                    className="rounded-md p-1.5 text-zinc-600 hover:bg-zinc-800 hover:text-rose-400 transition-colors"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              </div>
            </article>
          ))}
        </div>
      </main>

      {/* Create secret modal */}
      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={() => setShowCreate(false)} />
          <div className="relative w-full max-w-lg overflow-hidden rounded-xl border border-zinc-800 bg-zinc-950 shadow-2xl">
            <div className="flex items-center justify-between border-b border-zinc-800 px-5 py-4">
              <p className="font-semibold text-zinc-100">New secret</p>
              <button onClick={() => setShowCreate(false)} className="text-zinc-500 hover:text-zinc-300 transition-colors">
                <X className="h-4 w-4" />
              </button>
            </div>
            <form onSubmit={createSecret} className="space-y-4 p-5 text-sm">
              <label className="block">
                <span className="mb-1.5 block text-xs font-medium text-zinc-400">Name</span>
                <input
                  required
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  placeholder="MY_SECRET_KEY"
                  className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 font-mono text-zinc-100 placeholder-zinc-600 focus:border-zinc-500 focus:outline-none"
                />
              </label>
              <label className="block">
                <span className="mb-1.5 block text-xs font-medium text-zinc-400">Description</span>
                <input
                  value={newDesc}
                  onChange={(e) => setNewDesc(e.target.value)}
                  placeholder="What is this secret used for?"
                  className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-zinc-100 placeholder-zinc-600 focus:border-zinc-500 focus:outline-none"
                />
              </label>
              <label className="block">
                <span className="mb-1.5 block text-xs font-medium text-zinc-400">Value</span>
                <textarea
                  required
                  value={newValue}
                  onChange={(e) => setNewValue(e.target.value)}
                  rows={3}
                  className="w-full resize-none rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 font-mono text-zinc-100 placeholder-zinc-600 focus:border-zinc-500 focus:outline-none"
                />
              </label>

              <div>
                <p className="mb-2 text-xs font-medium text-zinc-400">Share with</p>
                <div className="grid grid-cols-2 gap-1.5">
                  {AVAILABLE_TARGETS.map((target) => {
                    const selected = selectedTargets.includes(target.id);
                    return (
                      <button
                        key={target.id}
                        type="button"
                        onClick={() =>
                          setSelectedTargets((prev) =>
                            selected ? prev.filter((id) => id !== target.id) : [...prev, target.id],
                          )
                        }
                        className={`flex items-center gap-2 rounded-lg border px-3 py-2 text-left text-xs transition-colors ${
                          selected
                            ? 'border-blue-500/40 bg-blue-500/10 text-blue-300'
                            : 'border-zinc-800 text-zinc-400 hover:bg-zinc-900'
                        }`}
                      >
                        {selected && <Check className="h-3 w-3 shrink-0" />}
                        <span className="truncate">{target.name}</span>
                        <span className={`ml-auto shrink-0 rounded px-1 py-0.5 text-[10px] ${targetBadgeClass(target.type)}`}>
                          {target.type[0].toUpperCase()}
                        </span>
                      </button>
                    );
                  })}
                </div>
              </div>

              <div className="flex justify-end gap-2 pt-1">
                <button
                  type="button"
                  onClick={() => setShowCreate(false)}
                  className="rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm hover:bg-zinc-800 transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="rounded-md bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-500 transition-colors"
                >
                  Create secret
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Share/access modal */}
      {shareModalSecret && (
        <ShareModal
          secret={shareModalSecret}
          onSave={(shares) => updateShares(shareModalSecret.id, shares)}
          onClose={() => setShareModalId(null)}
        />
      )}
    </AppNav>
  );
}

function ShareModal({
  secret,
  onSave,
  onClose,
}: {
  secret: Secret;
  onSave: (shares: SecretShare[]) => void;
  onClose: () => void;
}) {
  const [selected, setSelected] = useState<string[]>(secret.shares.map((s) => s.id));

  function toggle(id: string) {
    setSelected((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]));
  }

  function save() {
    const shares = AVAILABLE_TARGETS.filter((t) => selected.includes(t.id));
    onSave(shares);
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={onClose} />
      <div className="relative w-full max-w-md overflow-hidden rounded-xl border border-zinc-800 bg-zinc-950 shadow-2xl">
        <div className="flex items-center justify-between border-b border-zinc-800 px-5 py-4">
          <div>
            <p className="font-semibold text-zinc-100">Manage access</p>
            <p className="mt-0.5 font-mono text-xs text-zinc-500">{secret.name}</p>
          </div>
          <button onClick={onClose} className="text-zinc-500 hover:text-zinc-300 transition-colors">
            <X className="h-4 w-4" />
          </button>
        </div>
        <div className="p-4">
          {(['application', 'agent', 'workflow'] as SecretTarget[]).map((type) => {
            const targets = AVAILABLE_TARGETS.filter((t) => t.type === type);
            return (
              <div key={type} className="mb-4">
                <p className="mb-2 text-xs font-semibold uppercase tracking-wider text-zinc-500 capitalize">{type}s</p>
                <div className="space-y-1">
                  {targets.map((target) => {
                    const isSelected = selected.includes(target.id);
                    return (
                      <button
                        key={target.id}
                        onClick={() => toggle(target.id)}
                        className={`flex w-full items-center gap-3 rounded-lg px-3 py-2 text-left text-sm transition-colors ${
                          isSelected ? 'bg-blue-500/10 text-blue-300' : 'text-zinc-400 hover:bg-zinc-900 hover:text-zinc-200'
                        }`}
                      >
                        <div className={`flex h-4 w-4 shrink-0 items-center justify-center rounded border transition-colors ${isSelected ? 'border-blue-400 bg-blue-500' : 'border-zinc-700'}`}>
                          {isSelected && <Check className="h-2.5 w-2.5 text-white" />}
                        </div>
                        {target.name}
                      </button>
                    );
                  })}
                </div>
              </div>
            );
          })}
        </div>
        <div className="flex justify-end gap-2 border-t border-zinc-800 px-5 py-4">
          <button onClick={onClose} className="rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm hover:bg-zinc-800 transition-colors">
            Cancel
          </button>
          <button onClick={save} className="rounded-md bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-500 transition-colors">
            Save access
          </button>
        </div>
      </div>
    </div>
  );
}
