import { createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import {
  AgentControlService,
  type AgentRun as AgentRunMessage,
  ApprovalPolicy,
  RunStatus as RunStatusEnum,
  RuntimeProvider as RuntimeProviderEnum,
  SandboxMode,
} from '@x/sdk-agent-control';

const BASE = process.env.NEXT_PUBLIC_AGENT_CONTROL_URL ?? 'http://localhost:8090';

const transport = createConnectTransport({
  baseUrl: BASE,
  useBinaryFormat: true,
});

const client = createClient(AgentControlService, transport);

export type RunStatus = 'PENDING' | 'RUNNING' | 'SUCCEEDED' | 'FAILED';
export type RuntimeProvider = 'claude' | 'codex';

export interface AgentRun {
  id: string;
  message: string;
  runtime: {
    provider: RuntimeProvider;
  };
  status: RunStatus;
  job_name?: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
}

export type RunEvent =
  | { type: 'snapshot'; run: AgentRun }
  | { type: 'output'; chunk: string }
  | { type: 'completed'; status: RunStatus };

export async function listRuns(limit = 20): Promise<AgentRun[]> {
  const res = await client.listRuns({ pageSize: limit });
  return (res.runs ?? []).map((run) => fromMessage(run));
}

export async function createRun(message: string, provider: RuntimeProvider): Promise<AgentRun> {
  const res = await client.createRun({
    message,
    runtime: {
      provider: toProvider(provider),
      sandboxMode: SandboxMode.SANDBOX_MODE_WORKSPACE_WRITE,
      approvalPolicy: ApprovalPolicy.APPROVAL_POLICY_NEVER,
    },
  });
  if (!res.run) {
    throw new Error('create run: missing run in response');
  }
  return fromMessage(res.run);
}

export async function* watchRun(id: string, signal?: AbortSignal): AsyncIterable<RunEvent> {
  const stream = client.watchRun({ id, replayOutput: true }, { signal });
  for await (const event of stream) {
    switch (event.event.case) {
      case 'snapshot':
        if (event.event.value.run) {
          yield { type: 'snapshot', run: fromMessage(event.event.value.run) };
        }
        break;
      case 'output':
        yield { type: 'output', chunk: event.event.value.chunk };
        break;
      case 'completed':
        yield { type: 'completed', status: fromStatus(event.event.value.status) };
        break;
      default:
        break;
    }
  }
}

function fromMessage(run: AgentRunMessage): AgentRun {
  return {
    id: run.id,
    message: run.message,
    runtime: {
      provider: fromProvider(run.runtime?.provider ?? RuntimeProviderEnum.RUNTIME_PROVIDER_CLAUDE),
    },
    status: fromStatus(run.status),
    job_name: run.jobName,
    started_at: run.startedAt,
    completed_at: run.completedAt,
    created_at: run.createdAt,
  };
}

function fromStatus(status: RunStatusEnum): RunStatus {
  switch (status) {
    case RunStatusEnum.RUN_STATUS_RUNNING:
      return 'RUNNING';
    case RunStatusEnum.RUN_STATUS_SUCCEEDED:
      return 'SUCCEEDED';
    case RunStatusEnum.RUN_STATUS_FAILED:
      return 'FAILED';
    default:
      return 'PENDING';
  }
}

function toProvider(provider: RuntimeProvider): RuntimeProviderEnum {
  switch (provider) {
    case 'codex':
      return RuntimeProviderEnum.RUNTIME_PROVIDER_CODEX;
    default:
      return RuntimeProviderEnum.RUNTIME_PROVIDER_CLAUDE;
  }
}

function fromProvider(provider: RuntimeProviderEnum): RuntimeProvider {
  switch (provider) {
    case RuntimeProviderEnum.RUNTIME_PROVIDER_CODEX:
      return 'codex';
    default:
      return 'claude';
  }
}
