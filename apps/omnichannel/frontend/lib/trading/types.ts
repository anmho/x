export type BotType = 'kalshi' | 'crypto';

export type BotRunStatus = 'running' | 'paused' | 'stopped' | 'error';

export type TradingIntegration =
  | 'kalshi'
  | 'coinbase'
  | 'binance'
  | 'hyperliquid'
  | 'discord'
  | 'telegram'
  | 'slack'
  | 'webhook';

export type CreateBotRunRequest = {
  name: string;
  botType: BotType;
  strategy: string;
  market: string;
  environment: 'paper' | 'live';
  notionalUsd: number;
  maxRiskBps: number;
  integrations: TradingIntegration[];
  dryRun: boolean;
};

export type BotRun = {
  id: string;
  name: string;
  botType: BotType;
  strategy: string;
  market: string;
  environment: 'paper' | 'live';
  integrations: TradingIntegration[];
  status: BotRunStatus;
  notionalUsd: number;
  maxRiskBps: number;
  createdAt: string;
  updatedAt: string;
};

export type BotRunsResponse = {
  runs: BotRun[];
  source: 'xapi-live' | 'mock';
};

export type CreateBotRunResponse = {
  run: BotRun;
  source: 'xapi-live' | 'mock';
};
