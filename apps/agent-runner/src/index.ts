import { AgentControlClient } from "./control-plane.ts";
import { runWithClaude } from "./claude.ts";
import { loadConfig } from "./config.ts";
import { materializePrompt } from "./prompt.ts";
import type { RunEnvelope } from "./types.ts";

async function main(): Promise<void> {
  const config = loadConfig();
  const run = await resolveRun(config.controlPlaneBaseUrl, config.runId, config.message);
  const prompt = materializePrompt(run);

  console.log(JSON.stringify({ event: "run.started", runId: run.id, status: run.status }));
  const result = await runWithClaude({
    prompt,
    cwd: config.cwd,
    mcpConfigPath: config.mcpConfigPath,
  });
  console.log(JSON.stringify({ event: "run.completed", runId: run.id, outputChars: result.text.length }));
}

async function resolveRun(baseUrl: string, runId?: string, message?: string): Promise<RunEnvelope> {
  if (runId) {
    const client = new AgentControlClient(baseUrl);
    return client.getRun(runId);
  }
  if (message) {
    return {
      id: "local",
      message,
      status: "LOCAL_ONLY",
    };
  }
  throw new Error("AGENT_RUN_ID or AGENT_MESSAGE is required");
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
