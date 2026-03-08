import { BotRun, BotRunsResponse, CreateBotRunRequest, CreateBotRunResponse } from '@/lib/trading/types';

const XAPI_BASE_URL = process.env.XAPI_BASE_URL;
const XAPI_API_KEY = process.env.XAPI_API_KEY;

function hasLiveXapiConfig() {
  return Boolean(XAPI_BASE_URL && XAPI_API_KEY);
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  if (!XAPI_BASE_URL) {
    throw new Error('XAPI_BASE_URL is not configured');
  }

  const response = await fetch(`${XAPI_BASE_URL}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${XAPI_API_KEY}`,
      ...init?.headers,
    },
    cache: 'no-store',
  });

  if (!response.ok) {
    const message = await response.text();
    throw new Error(`XAPI request failed (${response.status}): ${message}`);
  }

  return (await response.json()) as T;
}

export function isXapiLiveEnabled() {
  return hasLiveXapiConfig();
}

export async function listRunsFromXapi(): Promise<BotRunsResponse> {
  const payload = await request<{ runs: BotRun[] }>('/v1/trading/bot-runs');
  return { runs: payload.runs, source: 'xapi-live' };
}

export async function createRunInXapi(body: CreateBotRunRequest): Promise<CreateBotRunResponse> {
  const payload = await request<{ run: BotRun }>('/v1/trading/bot-runs', {
    method: 'POST',
    body: JSON.stringify(body),
  });
  return { run: payload.run, source: 'xapi-live' };
}

export async function pauseRunInXapi(runId: string): Promise<CreateBotRunResponse> {
  const payload = await request<{ run: BotRun }>(`/v1/trading/bot-runs/${runId}/pause`, {
    method: 'POST',
  });
  return { run: payload.run, source: 'xapi-live' };
}
