import { createMcpServer } from "../common/jsonrpc-stdio";
import { getProject, loadPlatformProjects } from "../common/platform-projects";
import { mustRun } from "../common/shell";

function asString(value: unknown, field: string): string {
  if (typeof value !== "string" || !value.trim()) {
    throw new Error(`'${field}' must be a non-empty string.`);
  }
  return value;
}

createMcpServer({
  serverName: "vercel-observability",
  serverVersion: "0.1.0",
  tools: [
    {
      name: "configured_projects",
      description: "List Vercel-backed projects declared in platform.controlplane.json.",
      inputSchema: { type: "object", properties: {} },
    },
    {
      name: "recent_deployments",
      description: "Show recent deployments for a configured Vercel project using the local Vercel CLI session.",
      inputSchema: {
        type: "object",
        properties: {
          project: { type: "string", description: "Configured project name, for example cloud-console." }
        },
        required: ["project"],
        additionalProperties: false,
      },
    },
    {
      name: "inspect_url",
      description: "Inspect a Vercel deployment URL.",
      inputSchema: {
        type: "object",
        properties: {
          url: { type: "string", description: "Deployment URL to inspect." }
        },
        required: ["url"],
        additionalProperties: false,
      },
    },
    {
      name: "domains",
      description: "List Vercel domains from the current local CLI account.",
      inputSchema: { type: "object", properties: {} },
    },
  ],
  toolHandlers: {
    configured_projects: async () => {
      const projects = loadPlatformProjects().filter((project) => project.vercelProjectId);
      if (projects.length === 0) {
        return "No Vercel-backed projects are declared in platform.controlplane.json.";
      }
      return projects
        .map((project) => {
          const domains = project.vercelDomains.length > 0 ? project.vercelDomains.join(", ") : "none";
          return `${project.name}: vercelProjectId=${project.vercelProjectId}, domains=${domains}`;
        })
        .join("\n");
    },
    recent_deployments: async (args) => {
      const project = getProject(asString(args.project, "project"));
      if (!project.vercelProjectId) {
        throw new Error(`Project '${project.name}' is not configured for Vercel.`);
      }
      return mustRun("vercel", ["list", project.vercelProjectId]);
    },
    inspect_url: async (args) => {
      const url = asString(args.url, "url");
      return mustRun("vercel", ["inspect", url]);
    },
    domains: async () => mustRun("vercel", ["domains", "ls"]),
  },
});
