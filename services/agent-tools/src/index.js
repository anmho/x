/**
 * agent-tools: MCP sidecar for agent-runner
 *
 * Exposes a minimal set of tools via MCP over HTTP on :3000.
 * Runs as a Cloud Run sidecar alongside the claude-code container.
 * Claude Code connects to it via --mcp-server http://localhost:3000/mcp
 *
 * Tools exposed:
 *   - run_shell:  execute a shell command and return stdout/stderr
 *   - read_file:  read a file from the working directory
 *   - write_file: write content to a file in the working directory
 */

import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StreamableHTTPServerTransport } from "@modelcontextprotocol/sdk/server/streamableHttp.js";
import { createServer } from "http";
import { execFile } from "child_process";
import { readFile, writeFile, mkdir } from "fs/promises";
import { dirname } from "path";
import { z } from "zod";

const PORT = parseInt(process.env.PORT ?? "3000", 10);
const WORK_DIR = process.env.WORK_DIR ?? "/tmp/agent-workspace";

await mkdir(WORK_DIR, { recursive: true });

const server = new McpServer({
  name: "agent-tools",
  version: "0.1.0",
});

server.tool(
  "run_shell",
  "Run a shell command in the agent workspace and return stdout/stderr.",
  { command: z.string().describe("The shell command to execute") },
  async ({ command }) => {
    return new Promise((resolve) => {
      execFile("sh", ["-c", command], { cwd: WORK_DIR, timeout: 60_000 }, (err, stdout, stderr) => {
        const out = [stdout, stderr].filter(Boolean).join("\n");
        resolve({ content: [{ type: "text", text: err ? `ERROR (${err.code}):\n${out}` : out }] });
      });
    });
  }
);

server.tool(
  "read_file",
  "Read a file relative to the agent workspace.",
  { path: z.string().describe("Relative path inside the workspace") },
  async ({ path }) => {
    const content = await readFile(`${WORK_DIR}/${path}`, "utf8").catch((e) => `ERROR: ${e.message}`);
    return { content: [{ type: "text", text: content }] };
  }
);

server.tool(
  "write_file",
  "Write content to a file in the agent workspace.",
  {
    path: z.string().describe("Relative path inside the workspace"),
    content: z.string().describe("File content"),
  },
  async ({ path, content }) => {
    const full = `${WORK_DIR}/${path}`;
    await mkdir(dirname(full), { recursive: true });
    await writeFile(full, content, "utf8");
    return { content: [{ type: "text", text: `written: ${path}` }] };
  }
);

const httpServer = createServer(async (req, res) => {
  if (req.method === "POST" && req.url === "/mcp") {
    const transport = new StreamableHTTPServerTransport({ sessionIdGenerator: undefined });
    await server.connect(transport);
    await transport.handleRequest(req, res, await readBody(req));
    return;
  }
  if (req.url === "/health") {
    res.writeHead(200);
    res.end("ok");
    return;
  }
  res.writeHead(404);
  res.end();
});

function readBody(req) {
  return new Promise((resolve) => {
    const chunks = [];
    req.on("data", (c) => chunks.push(c));
    req.on("end", () => resolve(JSON.parse(Buffer.concat(chunks).toString())));
  });
}

httpServer.listen(PORT, () => console.log(`agent-tools MCP server listening on :${PORT}`));
