'use client';

import { FormEvent, useEffect, useMemo, useState } from 'react';
import { AppNav } from '@/app/_components/app-nav';
import { CATALOG_DEPLOYMENTS } from '@/lib/project-catalog';
import {
  ConsoleDomain,
  ConsoleDomainRecord,
  createDomainRecord,
  deleteDomainRecord,
  listDomainRecords,
  listDomains,
  reconcileDomains,
  updateDomainRecord,
} from '@/lib/domains-api';

function domainKey(domain: ConsoleDomain) {
  return `${domain.name}::${domain.provider}`;
}

function formatRecordName(record: ConsoleDomainRecord, zoneName: string) {
  if (record.name === zoneName) return '@';
  return record.name;
}

export default function DomainsPage() {
  const [domains, setDomains] = useState<ConsoleDomain[]>([]);
  const [selectedKey, setSelectedKey] = useState('');
  const [records, setRecords] = useState<ConsoleDomainRecord[]>([]);
  const [loadingDomains, setLoadingDomains] = useState(true);
  const [loadingRecords, setLoadingRecords] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [formType, setFormType] = useState('CNAME');
  const [formName, setFormName] = useState('');
  const [formContent, setFormContent] = useState('');
  const [formTTL, setFormTTL] = useState('300');
  const [formProxied, setFormProxied] = useState(false);
  const [editingRecordID, setEditingRecordID] = useState<string | null>(null);
  const [reconciling, setReconciling] = useState(false);

  const selectedDomain = useMemo(
    () => domains.find((entry) => domainKey(entry) === selectedKey) || null,
    [domains, selectedKey],
  );

  useEffect(() => {
    async function load() {
      try {
        setLoadingDomains(true);
        const data = await listDomains('cloud-console');
        setDomains(data);
        if (data.length > 0) {
          setSelectedKey(domainKey(data[0]));
        }
        setError(null);
      } catch (err) {
        const msg = err instanceof Error ? err.message : 'Failed to load domains';
        setError(msg);
      } finally {
        setLoadingDomains(false);
      }
    }
    void load();
  }, []);

  useEffect(() => {
    async function loadRecords() {
      if (!selectedDomain) {
        setRecords([]);
        return;
      }
      try {
        setLoadingRecords(true);
        const data = await listDomainRecords(selectedDomain.name, selectedDomain.provider);
        setRecords(data);
      } catch (err) {
        const msg = err instanceof Error ? err.message : 'Failed to load DNS records';
        setError(msg);
      } finally {
        setLoadingRecords(false);
      }
    }
    void loadRecords();
  }, [selectedDomain]);

  function resetForm() {
    setFormType('CNAME');
    setFormName('');
    setFormContent('');
    setFormTTL('300');
    setFormProxied(false);
    setEditingRecordID(null);
  }

  function applyDeploymentDefault(deploymentName: string) {
    setFormType('CNAME');
    setFormName(deploymentName);
    setFormContent('cname.vercel-dns.com');
    setFormTTL('300');
    setFormProxied(false);
  }

  async function refreshRecords() {
    if (!selectedDomain) return;
    const data = await listDomainRecords(selectedDomain.name, selectedDomain.provider);
    setRecords(data);
  }

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    if (!selectedDomain) return;
    try {
      const payload = {
        type: formType,
        name: formName.trim(),
        content: formContent.trim(),
        ttl: Number(formTTL),
        proxied: selectedDomain.provider === 'cloudflare' ? formProxied : undefined,
      };
      if (editingRecordID) {
        await updateDomainRecord(selectedDomain.name, selectedDomain.provider, editingRecordID, payload);
        setMessage('Record updated');
      } else {
        await createDomainRecord(selectedDomain.name, selectedDomain.provider, payload);
        setMessage('Record created');
      }
      setError(null);
      resetForm();
      await refreshRecords();
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to save record';
      setError(msg);
      setMessage(null);
    }
  }

  function startEdit(record: ConsoleDomainRecord) {
    setEditingRecordID(record.id);
    setFormType(record.type);
    setFormName(record.name);
    setFormContent(record.content);
    setFormTTL(String(record.ttl || 300));
    setFormProxied(Boolean(record.proxied));
  }

  async function removeRecord(record: ConsoleDomainRecord) {
    if (!selectedDomain || !record.id) return;
    if (!confirm(`Delete ${record.type} ${record.name}?`)) return;
    try {
      await deleteDomainRecord(selectedDomain.name, selectedDomain.provider, record.id);
      setMessage('Record deleted');
      setError(null);
      await refreshRecords();
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to delete record';
      setError(msg);
      setMessage(null);
    }
  }

  async function runReconcile(dryRun: boolean) {
    try {
      setReconciling(true);
      await reconcileDomains({ project: 'cloud-console', dry_run: dryRun, prune: false });
      setMessage(dryRun ? 'Reconcile preview complete' : 'Reconcile apply complete');
      setError(null);
      await refreshRecords();
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Reconcile failed';
      setError(msg);
      setMessage(null);
    } finally {
      setReconciling(false);
    }
  }

  return (
    <AppNav active="domains">
      <main className="mx-auto max-w-6xl px-4 py-8 sm:px-6">
        <div className="flex flex-wrap items-end justify-between gap-3">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Domains</h1>
            <p className="mt-1 text-sm text-zinc-400">
              Cloudflare + Vercel DNS management through the unified platform control plane.
            </p>
          </div>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => void runReconcile(true)}
              disabled={reconciling}
              className="rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-200 hover:bg-zinc-800 disabled:opacity-40"
            >
              Preview Reconcile
            </button>
            <button
              type="button"
              onClick={() => void runReconcile(false)}
              disabled={reconciling}
              className="rounded-md bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-500 disabled:opacity-40"
            >
              Apply Reconcile
            </button>
          </div>
        </div>

        {(message || error) && (
          <div className="mt-4 space-y-2">
            {message && <p className="rounded-md border border-emerald-500/40 bg-emerald-500/10 px-3 py-2 text-sm text-emerald-300">{message}</p>}
            {error && <p className="rounded-md border border-rose-500/40 bg-rose-500/10 px-3 py-2 text-sm text-rose-300">{error}</p>}
          </div>
        )}

        <section className="mt-6 rounded-xl border border-zinc-800 bg-zinc-950 p-4">
          <label className="text-xs uppercase tracking-wide text-zinc-500">Active zone</label>
          <select
            value={selectedKey}
            onChange={(event) => setSelectedKey(event.target.value)}
            className="mt-2 w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100"
            disabled={loadingDomains || domains.length === 0}
          >
            {domains.length === 0 && <option value="">No configured domains</option>}
            {domains.map((domain) => (
              <option key={domainKey(domain)} value={domainKey(domain)}>
                {domain.name} ({domain.provider})
              </option>
            ))}
          </select>
          {selectedDomain && (
            <p className="mt-2 text-xs text-zinc-500">
              project={selectedDomain.project || 'n/a'} • provider={selectedDomain.provider} • zone_id={selectedDomain.zone_id || selectedDomain.name}
            </p>
          )}
        </section>

        <section className="mt-6 grid grid-cols-1 gap-4 lg:grid-cols-[1.2fr_1fr]">
          <article className="rounded-xl border border-zinc-800 bg-zinc-950">
            <div className="border-b border-zinc-800 px-4 py-3">
              <p className="text-sm font-medium text-zinc-200">DNS Records</p>
            </div>
            <div className="p-4">
              {loadingRecords ? (
                <p className="text-sm text-zinc-500">Loading records...</p>
              ) : records.length === 0 ? (
                <p className="text-sm text-zinc-500">No records found for this zone.</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-zinc-800 text-sm">
                    <thead>
                      <tr className="text-left text-xs uppercase tracking-wide text-zinc-500">
                        <th className="px-3 py-2">Type</th>
                        <th className="px-3 py-2">Name</th>
                        <th className="px-3 py-2">Content</th>
                        <th className="px-3 py-2">TTL</th>
                        <th className="px-3 py-2">Actions</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-zinc-800">
                      {records.map((record) => (
                        <tr key={record.id || `${record.type}-${record.name}-${record.content}`}>
                          <td className="px-3 py-2">{record.type}</td>
                          <td className="px-3 py-2">{selectedDomain ? formatRecordName(record, selectedDomain.name) : record.name}</td>
                          <td className="max-w-[320px] truncate px-3 py-2 text-zinc-400">{record.content}</td>
                          <td className="px-3 py-2 text-zinc-400">{record.ttl || '-'}</td>
                          <td className="px-3 py-2">
                            <div className="flex gap-2">
                              <button
                                type="button"
                                onClick={() => startEdit(record)}
                                className="rounded border border-zinc-700 px-2 py-1 text-xs text-zinc-200 hover:bg-zinc-800"
                              >
                                Edit
                              </button>
                              <button
                                type="button"
                                onClick={() => void removeRecord(record)}
                                className="rounded border border-zinc-700 px-2 py-1 text-xs text-rose-300 hover:bg-zinc-800"
                              >
                                Delete
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </article>

          <article className="rounded-xl border border-zinc-800 bg-zinc-950 p-4">
            <p className="text-sm font-medium text-zinc-200">{editingRecordID ? 'Update record' : 'Create record'}</p>

            <div className="mt-3 rounded-md border border-zinc-800 bg-zinc-900/40 p-3">
              <p className="text-xs font-semibold uppercase tracking-wide text-zinc-500">Deployment-linked defaults</p>
              <div className="mt-2 flex flex-wrap gap-2">
                {CATALOG_DEPLOYMENTS.map((deployment) => (
                  <button
                    key={deployment.id}
                    type="button"
                    onClick={() => applyDeploymentDefault(deployment.name)}
                    className="rounded-md border border-zinc-700 px-2 py-1 text-xs text-zinc-300 hover:bg-zinc-800"
                  >
                    {deployment.name}
                  </button>
                ))}
              </div>
            </div>

            <form className="mt-4 space-y-3" onSubmit={onSubmit}>
              <label className="block">
                <span className="mb-1 block text-xs text-zinc-500">Type</span>
                <input
                  value={formType}
                  onChange={(event) => setFormType(event.target.value.toUpperCase())}
                  className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
                  required
                />
              </label>
              <label className="block">
                <span className="mb-1 block text-xs text-zinc-500">Name</span>
                <input
                  value={formName}
                  onChange={(event) => setFormName(event.target.value)}
                  className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
                  required
                />
              </label>
              <label className="block">
                <span className="mb-1 block text-xs text-zinc-500">Content</span>
                <input
                  value={formContent}
                  onChange={(event) => setFormContent(event.target.value)}
                  className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
                  required
                />
              </label>
              <label className="block">
                <span className="mb-1 block text-xs text-zinc-500">TTL</span>
                <input
                  value={formTTL}
                  onChange={(event) => setFormTTL(event.target.value)}
                  className="w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm"
                  type="number"
                  min={60}
                />
              </label>
              {selectedDomain?.provider === 'cloudflare' && (
                <label className="flex items-center gap-2 text-sm text-zinc-300">
                  <input
                    type="checkbox"
                    checked={formProxied}
                    onChange={(event) => setFormProxied(event.target.checked)}
                  />
                  Proxied through Cloudflare
                </label>
              )}
              <div className="flex gap-2">
                <button
                  type="submit"
                  disabled={!selectedDomain}
                  className="rounded-md bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-500 disabled:opacity-40"
                >
                  {editingRecordID ? 'Update' : 'Create'}
                </button>
                {editingRecordID && (
                  <button
                    type="button"
                    onClick={resetForm}
                    className="rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-300 hover:bg-zinc-800"
                  >
                    Cancel edit
                  </button>
                )}
              </div>
            </form>
          </article>
        </section>
      </main>
    </AppNav>
  );
}
