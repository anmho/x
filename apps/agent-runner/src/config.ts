export interface WorkerConfig {
  controlPlaneBaseUrl: string;
  runId?: string;
  message?: string;
  cwd: string;
  mcpConfigPath?: string;
  provider: "claude";
}

export function loadConfig(env: Record<string, string | undefined> = process.env): WorkerConfig {
  return {
    controlPlaneBaseUrl: env.AGENT_CONTROL_BASE_URL ?? "http://localhost:8090",
    runId: emptyToUndefined(env.AGENT_RUN_ID),
    message: emptyToUndefined(env.AGENT_MESSAGE) ?? emptyToUndefined(env.AGENT_PROMPT),
    cwd: env.AGENT_CWD ?? process.cwd(),
    mcpConfigPath: emptyToUndefined(env.AGENT_MCP_CONFIG_PATH),
    provider: "claude",
  };
}

export function emptyToUndefined(value: string | undefined): string | undefined {
  if (value == null) {
    return undefined;
  }
  const trimmed = value.trim();
  return trimmed === "" ? undefined : trimmed;
}
