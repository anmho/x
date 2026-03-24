import { query } from "@anthropic-ai/claude-agent-sdk";

export interface ClaudeRunInput {
  prompt: string;
  cwd: string;
  mcpConfigPath?: string;
}

export interface ClaudeRunResult {
  text: string;
}

export async function runWithClaude(input: ClaudeRunInput): Promise<ClaudeRunResult> {
	const parts: string[] = [];
	const mcpServers = await loadMcpServers(input.mcpConfigPath);

	for await (const message of query({
		prompt: input.prompt,
		options: {
			cwd: input.cwd,
			permissionMode: "dontAsk",
			settingSources: [],
			allowedTools: ["Read", "Write", "Edit", "Bash", "Glob", "Grep"],
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

async function loadMcpServers(configPath?: string): Promise<Record<string, unknown>> {
	if (!configPath) {
		return {};
	}
	const raw = await Bun.file(configPath).text();
	const parsed = JSON.parse(raw) as { mcpServers?: Record<string, unknown> };
	return parsed.mcpServers ?? {};
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
