import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { AgentControlService, type AgentRun as AgentRunMessage } from "@x/sdk-agent-control";

import type { RunEnvelope, RunResource } from "./types.ts";

export class AgentControlClient {
  private readonly client;

  constructor(baseUrl: string) {
    const transport = createConnectTransport({
      baseUrl,
      useBinaryFormat: true,
    });
    this.client = createClient(AgentControlService, transport);
  }

  async getRun(id: string): Promise<RunEnvelope> {
    const res = await this.client.getRun({ id });
    if (!res.run) {
      throw new Error(`getRun(${id}): missing run in response`);
    }
    return fromAgentRun(res.run);
  }
}

function fromAgentRun(run: AgentRunMessage): RunEnvelope {
  return {
    id: run.id,
    message: run.message,
    status: String(run.status),
    resources: run.resources?.map(fromResource),
  };
}

function fromResource(resource: NonNullable<AgentRunMessage["resources"]>[number]): RunResource {
  return {
    uri: resource.uri,
    title: resource.title || undefined,
    mimeType: resource.mimeType || undefined,
    text: resource.text || undefined,
  };
}
