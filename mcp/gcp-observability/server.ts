import { createMcpServer } from "../common/jsonrpc-stdio";
import { getProject, loadPlatformProjects } from "../common/platform-projects";
import { mustRun } from "../common/shell";

function asString(value: unknown, field: string): string {
  if (typeof value !== "string" || !value.trim()) {
    throw new Error(`'${field}' must be a non-empty string.`);
  }
  return value;
}

function asOptionalString(value: unknown): string | undefined {
  return typeof value === "string" && value.trim() ? value : undefined;
}

createMcpServer({
  serverName: "gcp-observability",
  serverVersion: "0.1.0",
  tools: [
    {
      name: "configured_projects",
      description: "List GCP-backed services declared in platform.controlplane.json.",
      inputSchema: { type: "object", properties: {} },
    },
    {
      name: "recent_logs",
      description: "Read recent Cloud Logging entries for a configured service.",
      inputSchema: {
        type: "object",
        properties: {
          project: { type: "string", description: "Configured project name, for example omnichannel-api." },
          limit: { type: "number", description: "Maximum number of log lines to fetch. Default 20." },
          service: { type: "string", description: "Optional Cloud Run service override." }
        },
        required: ["project"],
        additionalProperties: false,
      },
    },
    {
      name: "run_services",
      description: "List Cloud Run services for a configured GCP project.",
      inputSchema: {
        type: "object",
        properties: {
          project: { type: "string", description: "Configured project name." }
        },
        required: ["project"],
        additionalProperties: false,
      },
    },
    {
      name: "active_account",
      description: "Show the active gcloud account.",
      inputSchema: { type: "object", properties: {} },
    },
  ],
  toolHandlers: {
    configured_projects: async () => {
      const projects = loadPlatformProjects().filter((project) => project.gcpProjectId);
      if (projects.length === 0) {
        return "No GCP-backed projects are declared in platform.controlplane.json.";
      }
      return projects
        .map((project) => {
          const services = project.gcpServices.length > 0 ? project.gcpServices.join(", ") : "none";
          return `${project.name}: gcpProjectId=${project.gcpProjectId}, region=${project.gcpRegion ?? "unknown"}, services=${services}`;
        })
        .join("\n");
    },
    recent_logs: async (args) => {
      const project = getProject(asString(args.project, "project"));
      const service = asOptionalString(args.service) ?? project.gcpServices[0];
      if (!project.gcpProjectId) {
        throw new Error(`Project '${project.name}' is not configured for GCP.`);
      }
      if (!service) {
        throw new Error(`Project '${project.name}' does not declare a Cloud Run service.`);
      }
      const limit = Number.isFinite(args.limit) ? String(args.limit) : "20";
      const filter = `resource.type=cloud_run_revision AND resource.labels.service_name=${service}`;
      return mustRun("gcloud", [
        "logging",
        "read",
        filter,
        "--project",
        project.gcpProjectId,
        "--limit",
        limit,
        "--format=value(timestamp,textPayload,jsonPayload.message,severity)",
      ]);
    },
    run_services: async (args) => {
      const project = getProject(asString(args.project, "project"));
      if (!project.gcpProjectId) {
        throw new Error(`Project '${project.name}' is not configured for GCP.`);
      }
      const region = project.gcpRegion ?? "us-central1";
      return mustRun("gcloud", [
        "run",
        "services",
        "list",
        "--project",
        project.gcpProjectId,
        "--region",
        region,
      ]);
    },
    active_account: async () =>
      mustRun("gcloud", ["auth", "list", "--filter=status:ACTIVE", "--format=value(account)"]),
  },
});
