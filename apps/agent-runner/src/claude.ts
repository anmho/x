import { query } from "@anthropic-ai/claude-agent-sdk";

export interface ClaudeRunInput {
  prompt: string;
  cwd: string;
  mcpConfigPath: string;
}

export interface ClaudeRunResult {
  text: string;
}

export const MCP_ALLOWED_TOOLS = [
  "mcp__tools__workspace_read_file",
  "mcp__tools__workspace_list_files",
  "mcp__tools__workspace_search",
  "mcp__tools__workspace_apply_patch",
  "mcp__tools__git_status",
  "mcp__tools__git_diff",
  "mcp__tools__git_checkout",
] as const;

export async function runWithClaude(input: ClaudeRunInput): Promise<ClaudeRunResult> {
	const parts: string[] = [];
	const mcpServers = await loadMcpServers(input.mcpConfigPath);
	if (Object.keys(mcpServers).length === 0) {
		throw new Error("Project X MCP server configuration is required; native Claude workspace tools are disabled.");
	}

	for await (const message of query({
		prompt: input.prompt,
		options: {
			cwd: input.cwd,
			permissionMode: "dontAsk",
			settingSources: [],
			allowedTools: [...MCP_ALLOWED_TOOLS],
			mcpServers,
		},
	})) {
    const text = extractText(message);
    if (!text) {
      continue;
    }
    parts.push(text);
    process.stdout.write(text.endsWith("\n") ? text : `${text}\n`);
  }

	return { text: parts.join("\n").trim() };
}

async function loadMcpServers(configPath: string): Promise<Record<string, unknown>> {
	const raw = await Bun.file(configPath).text();
	const parsed = JSON.parse(raw) as { mcpServers?: Record<string, unknown> };
	return expandEnvStrings(parsed.mcpServers ?? {}) as Record<string, unknown>;
}

function expandEnvStrings(value: unknown): unknown {
	if (typeof value === "string") {
		return value.replace(/\$\{([A-Z0-9_]+)\}/g, (_, key: string) => process.env[key] ?? "");
	}
	if (Array.isArray(value)) {
		return value.map((item) => expandEnvStrings(item));
	}
	if (value && typeof value === "object") {
		return Object.fromEntries(
			Object.entries(value as Record<string, unknown>).map(([key, entry]) => [key, expandEnvStrings(entry)]),
		);
	}
	return value;
}

function extractText(message: unknown): string {
  if (!message || typeof message !== "object") {
    return "";
  }
  const record = message as Record<string, unknown>;
  if (typeof record.content === "string") {
    return record.content;
  }
  if (Array.isArray(record.content)) {
    const parts = record.content
      .map((part) => {
        if (!part || typeof part !== "object") {
          return "";
        }
        const partRecord = part as Record<string, unknown>;
        return typeof partRecord.text === "string" ? partRecord.text : "";
      })
      .filter(Boolean);
    return parts.join("\n");
  }
  return "";
}
