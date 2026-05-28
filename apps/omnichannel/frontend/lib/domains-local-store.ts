import 'server-only';
import fs from 'node:fs';
import path from 'node:path';

type DomainConfigRecord = {
  id?: string;
  type: string;
  name: string;
  content: string;
  ttl?: number;
  proxied?: boolean;
  desired_state?: string;
};

type DomainConfigZone = {
  name: string;
  provider: string;
  zone_id?: string;
  project?: string;
  desired_state?: string;
  records?: DomainConfigRecord[];
};

type ControlPlaneConfigProject = {
  name: string;
  domains?: DomainConfigZone[];
};

type ControlPlaneConfig = {
  projects?: ControlPlaneConfigProject[];
};

type LocalDomain = {
  name: string;
  provider: string;
  zone_id?: string;
  project?: string;
};

type LocalDomainRecord = {
  id: string;
  type: string;
  name: string;
  content: string;
  ttl?: number;
  proxied?: boolean;
  provider?: string;
};

type LocalDomainRecordInput = {
  type?: string;
  name?: string;
  content?: string;
  ttl?: number;
  proxied?: boolean;
};

const LOCAL_RECORD_PREFIX = 'local-rec-';
const state = {
  seeded: false,
  nextID: 1,
  domains: [] as LocalDomain[],
  recordsByDomainKey: new Map<string, LocalDomainRecord[]>(),
};

function toDomainKey(domain: string, provider: string) {
  return `${domain.trim().toLowerCase()}::${provider.trim().toLowerCase()}`;
}

function cloneRecord(record: LocalDomainRecord): LocalDomainRecord {
  return {
    ...record,
  };
}

function cloneDomain(domain: LocalDomain): LocalDomain {
  return {
    ...domain,
  };
}

function nextRecordID() {
  const id = `${LOCAL_RECORD_PREFIX}${state.nextID}`;
  state.nextID += 1;
  return id;
}

function resolveControlPlaneConfigPath(): string | null {
  const cwd = process.cwd();
  const candidates = [
    'platform.controlplane.json',
    '../platform.controlplane.json',
    '../../platform.controlplane.json',
    '../../../platform.controlplane.json',
    '../../../../platform.controlplane.json',
  ].map((entry) => path.resolve(cwd, entry));
  for (const candidate of candidates) {
    if (fs.existsSync(candidate)) {
      return candidate;
    }
  }
  return null;
}

function coerceJSONConfig(raw: string): ControlPlaneConfig {
  try {
    return JSON.parse(raw) as ControlPlaneConfig;
  } catch {
    return {};
  }
}

function seedLocalStore() {
  if (state.seeded) {
    return;
  }
  state.seeded = true;

  const configPath = resolveControlPlaneConfigPath();
  if (!configPath) {
    return;
  }

  let parsed: ControlPlaneConfig = {};
  try {
    const raw = fs.readFileSync(configPath, 'utf8');
    parsed = coerceJSONConfig(raw);
  } catch {
    parsed = {};
  }
  const projects = parsed.projects || [];

  for (const project of projects) {
    const domains = project.domains || [];
    for (const domain of domains) {
      if ((domain.desired_state || '').trim().toLowerCase() === 'absent') {
        continue;
      }

      const normalizedDomain: LocalDomain = {
        name: domain.name,
        provider: domain.provider,
        zone_id: domain.zone_id,
        project: domain.project || project.name,
      };
      state.domains.push(normalizedDomain);

      const key = toDomainKey(normalizedDomain.name, normalizedDomain.provider);
      const existing = state.recordsByDomainKey.get(key) || [];
      const records = domain.records || [];
      for (const record of records) {
        if ((record.desired_state || '').trim().toLowerCase() === 'absent') {
          continue;
        }
        const normalizedRecord: LocalDomainRecord = {
          id: record.id || nextRecordID(),
          type: record.type,
          name: record.name,
          content: record.content,
          ttl: record.ttl,
          proxied: record.proxied,
          provider: normalizedDomain.provider,
        };
        existing.push(normalizedRecord);
      }
      state.recordsByDomainKey.set(key, existing);
    }
  }
}

function requireDomain(domain: string, provider: string): LocalDomain {
  seedLocalStore();
  const key = toDomainKey(domain, provider);
  const found = state.domains.find((entry) => toDomainKey(entry.name, entry.provider) === key);
  if (!found) {
    throw new Error(`domain zone "${domain}" (${provider}) not found in local fallback store`);
  }
  return found;
}

function requireValidRecordInput(input: LocalDomainRecordInput) {
  if (!input.type || !input.name || !input.content) {
    throw new Error('type, name, and content are required');
  }
}

export function listLocalDomains(project?: string): LocalDomain[] {
  seedLocalStore();
  const projectName = (project || '').trim();
  const source = projectName
    ? state.domains.filter((entry) => (entry.project || '').trim() === projectName)
    : state.domains;
  return source.map(cloneDomain);
}

export function listLocalDomainRecords(domain: string, provider: string): LocalDomainRecord[] {
  requireDomain(domain, provider);
  const key = toDomainKey(domain, provider);
  const records = state.recordsByDomainKey.get(key) || [];
  return records.map(cloneRecord);
}

export function createLocalDomainRecord(domain: string, provider: string, input: LocalDomainRecordInput): LocalDomainRecord {
  requireDomain(domain, provider);
  requireValidRecordInput(input);
  const key = toDomainKey(domain, provider);
  const records = state.recordsByDomainKey.get(key) || [];
  const created: LocalDomainRecord = {
    id: nextRecordID(),
    type: String(input.type).toUpperCase(),
    name: String(input.name).trim(),
    content: String(input.content).trim(),
    ttl: Number.isFinite(input.ttl) ? Number(input.ttl) : 300,
    proxied: input.proxied,
    provider,
  };
  records.push(created);
  state.recordsByDomainKey.set(key, records);
  return cloneRecord(created);
}

export function updateLocalDomainRecord(
  domain: string,
  provider: string,
  recordID: string,
  input: LocalDomainRecordInput,
): LocalDomainRecord {
  requireDomain(domain, provider);
  requireValidRecordInput(input);
  const key = toDomainKey(domain, provider);
  const records = state.recordsByDomainKey.get(key) || [];
  const target = records.find((entry) => entry.id === recordID);
  if (!target) {
    throw new Error(`record "${recordID}" not found`);
  }
  target.type = String(input.type).toUpperCase();
  target.name = String(input.name).trim();
  target.content = String(input.content).trim();
  target.ttl = Number.isFinite(input.ttl) ? Number(input.ttl) : target.ttl;
  target.proxied = input.proxied;
  return cloneRecord(target);
}

export function deleteLocalDomainRecord(domain: string, provider: string, recordID: string): boolean {
  requireDomain(domain, provider);
  const key = toDomainKey(domain, provider);
  const records = state.recordsByDomainKey.get(key) || [];
  const index = records.findIndex((entry) => entry.id === recordID);
  if (index < 0) {
    return false;
  }
  records.splice(index, 1);
  state.recordsByDomainKey.set(key, records);
  return true;
}
