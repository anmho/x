export type AccessScope = {
  service: string;
  scope: string;
};

export type AccessKey = {
  id: string;
  key?: string;
  key_prefix: string;
  application: string;
  environment: 'dev' | 'staging' | 'prod';
  owner: string;
  status: 'active' | 'rotated' | 'revoked';
  service_scopes: AccessScope[];
  created_at: string;
};

export type MintAccessKeyRequest = {
  application: string;
  environment: 'dev' | 'staging' | 'prod';
  owner: string;
  service_scopes: AccessScope[];
};

type KeysListResponse = {
  keys: AccessKey[];
};

type MintResponse = {
  key: AccessKey;
};

type RevokeResponse = {
  key: AccessKey;
};

async function parseJSON<T>(response: Response): Promise<T> {
  const text = await response.text();
  const payload = text ? JSON.parse(text) : {};
  if (!response.ok) {
    const message = payload?.error || `Request failed (${response.status})`;
    throw new Error(message);
  }
  return payload as T;
}

export async function listAccessKeys(): Promise<AccessKey[]> {
  const response = await fetch('/api/keys', {
    method: 'GET',
    cache: 'no-store',
  });
  const payload = await parseJSON<KeysListResponse>(response);
  return payload.keys || [];
}

export async function mintAccessKey(input: MintAccessKeyRequest): Promise<AccessKey> {
  const response = await fetch('/api/keys', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  });
  const payload = await parseJSON<MintResponse>(response);
  return payload.key;
}

export async function revokeAccessKey(id: string): Promise<AccessKey> {
  const response = await fetch(`/api/keys/${encodeURIComponent(id)}/revoke`, {
    method: 'POST',
  });
  const payload = await parseJSON<RevokeResponse>(response);
  return payload.key;
}

