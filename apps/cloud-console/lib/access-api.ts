/**
 * Server-side helper for proxying requests to the access API.
 * Used by /api/keys and /api/keys/[id]/revoke routes.
 *
 * ACCESS_API_ADMIN_KEY has no fallback. If unset, accessAPI throws at first use.
 */

const ACCESS_API_BASE_URL = (process.env.ACCESS_API_URL || 'http://127.0.0.1:8090').replace(/\/$/, '');

function getAdminKey(): string {
  const key = process.env.ACCESS_API_ADMIN_KEY;
  if (!key || key.trim() === '') {
    throw new Error(
      'ACCESS_API_ADMIN_KEY environment variable is required. Set it in .env or your environment.'
    );
  }
  return key;
}

/**
 * Fetches the access API with the admin key. Throws if ACCESS_API_ADMIN_KEY is unset.
 */
export async function accessAPI(path: string, init?: RequestInit): Promise<Response> {
  const adminKey = getAdminKey();
  return fetch(`${ACCESS_API_BASE_URL}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': adminKey,
      ...(init?.headers || {}),
    },
    cache: 'no-store',
  });
}

export { ACCESS_API_BASE_URL };
