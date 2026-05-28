import { BotRun, CreateBotRunRequest } from '@/lib/trading/types';

const now = new Date().toISOString();

const store: BotRun[] = [
  {
    id: 'run_kalshi_001',
    name: 'Kalshi CPI Momentum',
    botType: 'kalshi',
    strategy: 'event-breakout',
    market: 'INFLATION.CPI.MOM',
    environment: 'paper',
    integrations: ['kalshi', 'slack', 'webhook'],
    status: 'running',
    notionalUsd: 2500,
    maxRiskBps: 150,
    createdAt: now,
    updatedAt: now,
  },
  {
    id: 'run_crypto_001',
    name: 'BTC Basis Arb',
    botType: 'crypto',
    strategy: 'basis-arb',
    market: 'BTC-USD',
    environment: 'live',
    integrations: ['coinbase', 'binance', 'discord'],
    status: 'paused',
    notionalUsd: 5000,
    maxRiskBps: 120,
    createdAt: now,
    updatedAt: now,
  },
];

function makeId() {
  return `run_${Math.random().toString(36).slice(2, 10)}`;
}

export function listMockRuns(): BotRun[] {
  return [...store].sort((a, b) => b.updatedAt.localeCompare(a.updatedAt));
}

export function createMockRun(payload: CreateBotRunRequest): BotRun {
  const stamp = new Date().toISOString();
  const run: BotRun = {
    id: makeId(),
    name: payload.name,
    botType: payload.botType,
    strategy: payload.strategy,
    market: payload.market,
    environment: payload.environment,
    integrations: payload.integrations,
    status: payload.dryRun ? 'paused' : 'running',
    notionalUsd: payload.notionalUsd,
    maxRiskBps: payload.maxRiskBps,
    createdAt: stamp,
    updatedAt: stamp,
  };

  store.unshift(run);
  return run;
}

export function pauseMockRun(runId: string): BotRun | null {
  const run = store.find((item) => item.id === runId);
  if (!run) return null;
  run.status = 'paused';
  run.updatedAt = new Date().toISOString();
  return run;
}
