export type ConsoleDomain = {
  name: string;
  provider: string;
  zone_id?: string;
  project?: string;
};

export type ConsoleDomainRecord = {
  id: string;
  type: string;
  name: string;
  content: string;
  ttl?: number;
  proxied?: boolean;
  provider?: string;
};

type DomainsResponse = {
  domains: ConsoleDomain[];
};

type RecordsResponse = {
  records: ConsoleDomainRecord[];
};

type RecordResponse = {
  record: ConsoleDomainRecord;
};

type ReconcileResponse = {
  ok: boolean;
  dry_run: boolean;
  prune: boolean;
  project: string;
};

async function parseJSON<T>(response: Response): Promise<T> {
  const text = await response.text();
  const payload = text ? JSON.parse(text) : {};
  if (!response.ok) {
    throw new Error(payload?.error || `Request failed (${response.status})`);
  }
  return payload as T;
}

export async function listDomains(project?: string): Promise<ConsoleDomain[]> {
  const query = project ? `?project=${encodeURIComponent(project)}` : '';
  const response = await fetch(`/api/domains${query}`, { method: 'GET', cache: 'no-store' });
  const payload = await parseJSON<DomainsResponse>(response);
  return payload.domains || [];
}

export async function listDomainRecords(domain: string, provider: string): Promise<ConsoleDomainRecord[]> {
  const response = await fetch(`/api/domains/${encodeURIComponent(domain)}/records?provider=${encodeURIComponent(provider)}`, {
    method: 'GET',
    cache: 'no-store',
  });
  const payload = await parseJSON<RecordsResponse>(response);
  return payload.records || [];
}

export async function createDomainRecord(domain: string, provider: string, input: Partial<ConsoleDomainRecord>): Promise<ConsoleDomainRecord> {
  const response = await fetch(`/api/domains/${encodeURIComponent(domain)}/records?provider=${encodeURIComponent(provider)}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  });
  const payload = await parseJSON<RecordResponse>(response);
  return payload.record;
}

export async function updateDomainRecord(
  domain: string,
  provider: string,
  recordId: string,
  input: Partial<ConsoleDomainRecord>,
): Promise<ConsoleDomainRecord> {
  const response = await fetch(
    `/api/domains/${encodeURIComponent(domain)}/records/${encodeURIComponent(recordId)}?provider=${encodeURIComponent(provider)}`,
    {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(input),
    },
  );
  const payload = await parseJSON<RecordResponse>(response);
  return payload.record;
}

export async function deleteDomainRecord(domain: string, provider: string, recordId: string): Promise<void> {
  const response = await fetch(
    `/api/domains/${encodeURIComponent(domain)}/records/${encodeURIComponent(recordId)}?provider=${encodeURIComponent(provider)}`,
    { method: 'DELETE' },
  );
  await parseJSON<{ deleted: boolean }>(response);
}

export async function reconcileDomains(input: { project?: string; dry_run: boolean; prune?: boolean }): Promise<ReconcileResponse> {
  const response = await fetch('/api/domains/reconcile', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  });
  return parseJSON<ReconcileResponse>(response);
}

