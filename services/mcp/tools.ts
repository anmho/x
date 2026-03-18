/**
 * All MCP tool definitions and handlers.
 * Currently exposes: gcp-observability, vercel-observability.
 */

import { getProject, loadPlatformProjects } from "../../mcp/common/platform-projects.ts";
import { mustRun } from "../../mcp/common/shell.ts";
import type { Tool } from "./transport.ts";

function asString(value: unknown, field: string): string {
  if (typeof value !== "string" || !value.trim()) {
    throw new Error(`'${field}' must be a non-empty string.`);
  }
  return value;
}

function asOptionalString(value: unknown): string | undefined {
  return typeof value === "string" && value.trim() ? value : undefined;
}

// ---------------------------------------------------------------------------
// Tool definitions
// ---------------------------------------------------------------------------

export const tools: Tool[] = [
  // GCP observability
  {
    name: "gcp_configured_projects",
    description: "List GCP-backed services declared in platform.controlplane.json.",
    inputSchema: { type: "object", properties: {} },
  },
  {
    name: "gcp_recent_logs",
    description: "Read recent Cloud Logging entries for a configured GCP service.",
    inputSchema: {
      type: "object",
      properties: {
        project: { type: "string", description: "Configured project name, e.g. omnichannel-api." },
        limit: { type: "number", description: "Maximum log lines to fetch. Default 20." },
        service: { type: "string", description: "Optional Cloud Run service override." },
      },
      required: ["project"],
      additionalProperties: false,
    },
  },
  {
    name: "gcp_run_services",
    description: "List Cloud Run services for a configured GCP project.",
    inputSchema: {
      type: "object",
      properties: {
        project: { type: "string", description: "Configured project name." },
      },
      required: ["project"],
      additionalProperties: false,
    },
  },
  {
    name: "gcp_active_account",
    description: "Show the active gcloud account.",
    inputSchema: { type: "object", properties: {} },
  },

  // Vercel observability
  {
    name: "vercel_configured_projects",
    description: "List Vercel-backed projects declared in platform.controlplane.json.",
    inputSchema: { type: "object", properties: {} },
  },
  {
    name: "vercel_recent_deployments",
    description: "Show recent deployments for a configured Vercel project.",
    inputSchema: {
      type: "object",
      properties: {
        project: { type: "string", description: "Configured project name, e.g. cloud-console." },
      },
      required: ["project"],
      additionalProperties: false,
    },
  },
  {
    name: "vercel_inspect_url",
    description: "Inspect a Vercel deployment URL.",
    inputSchema: {
      type: "object",
      properties: {
        url: { type: "string", description: "Deployment URL to inspect." },
      },
      required: ["url"],
      additionalProperties: false,
    },
  },
  {
    name: "vercel_domains",
    description: "List Vercel domains from the current local CLI account.",
    inputSchema: { type: "object", properties: {} },
  },
];

// ---------------------------------------------------------------------------
// Tool handlers
// ---------------------------------------------------------------------------

export const toolHandlers: Record<string, (args: Record<string, unknown>) => Promise<string> | string> = {
  gcp_configured_projects: () => {
    const projects = loadPlatformProjects().filter((p) => p.gcpProjectId);
    if (projects.length === 0) return "No GCP-backed projects declared in platform.controlplane.json.";
    return projects
      .map((p) => {
        const services = p.gcpServices.length > 0 ? p.gcpServices.join(", ") : "none";
        return `${p.name}: gcpProjectId=${p.gcpProjectId}, region=${p.gcpRegion ?? "unknown"}, services=${services}`;
      })
      .join("\n");
  },

  gcp_recent_logs: async (args) => {
    const project = getProject(asString(args.project, "project"));
    if (!project.gcpProjectId) throw new Error(`Project '${project.name}' is not configured for GCP.`);
    const service = asOptionalString(args.service) ?? project.gcpServices[0];
    if (!service) throw new Error(`Project '${project.name}' declares no Cloud Run service.`);
    const limit = Number.isFinite(args.limit) ? String(args.limit) : "20";
    return mustRun("gcloud", [
      "logging", "read",
      `resource.type=cloud_run_revision AND resource.labels.service_name=${service}`,
      "--project", project.gcpProjectId,
      "--limit", limit,
      "--format=value(timestamp,textPayload,jsonPayload.message,severity)",
    ]);
  },

  gcp_run_services: async (args) => {
    const project = getProject(asString(args.project, "project"));
    if (!project.gcpProjectId) throw new Error(`Project '${project.name}' is not configured for GCP.`);
    return mustRun("gcloud", [
      "run", "services", "list",
      "--project", project.gcpProjectId,
      "--region", project.gcpRegion ?? "us-central1",
    ]);
  },

  gcp_active_account: () =>
    mustRun("gcloud", ["auth", "list", "--filter=status:ACTIVE", "--format=value(account)"]),

  vercel_configured_projects: () => {
    const projects = loadPlatformProjects().filter((p) => p.vercelProjectId);
    if (projects.length === 0) return "No Vercel-backed projects declared in platform.controlplane.json.";
    return projects
      .map((p) => {
        const domains = p.vercelDomains.length > 0 ? p.vercelDomains.join(", ") : "none";
        return `${p.name}: vercelProjectId=${p.vercelProjectId}, domains=${domains}`;
      })
      .join("\n");
  },

  vercel_recent_deployments: async (args) => {
    const project = getProject(asString(args.project, "project"));
    if (!project.vercelProjectId) throw new Error(`Project '${project.name}' is not configured for Vercel.`);
    return mustRun("vercel", ["list", project.vercelProjectId]);
  },

  vercel_inspect_url: async (args) => mustRun("vercel", ["inspect", asString(args.url, "url")]),

  vercel_domains: () => mustRun("vercel", ["domains", "ls"]),
};
