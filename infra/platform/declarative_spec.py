"""Declarative source-of-truth for Project X platform config materialization."""

from __future__ import annotations

import json
from copy import deepcopy
from pathlib import Path
from typing import Any

VERSION = "1"
DEFAULT_GCP_PROJECT_ID = "my-gcp-project"
DEFAULT_REGION = "us-central1"
INTEGRATION_OVERRIDES_PATH = Path(__file__).resolve().parent / "integrations.overrides.json"

PROJECT_SPECS: list[dict[str, Any]] = [
    {
        "name": "cloud-console",
        "description": "Web dashboard for service access and API key workflows",
        "path": "apps/cloud-console",
        "dependencies": [
            {"name": "access-api", "type": "stack"},
            {"name": "omnichannel-api", "type": "stack"},
        ],
        "deploy": {
            "cwd": "apps/cloud-console",
            "command": "npm run build",
            "health_url": "http://127.0.0.1:3000",
        },
        "targets": {
            "endpoints": [
                {
                    "name": "access-policy-check",
                    "url": "http://127.0.0.1:8090/v1/policy/check",
                    "service": "notifications",
                    "scope": "read",
                    "key_env": "ACCESS_API_KEY",
                },
                {
                    "name": "omnichannel-health",
                    "url": "http://127.0.0.1:8080/api/v1/health",
                },
            ],
        },
        "terraform": {
            "root": "infra/terraform/projects/cloud-console",
            "workspace": "dev",
        },
        "observability": {
            "vercel": {
                "project_id": "cloud-console",
                "environments": ["production", "preview"],
            },
            "gcp": {
                "project_id": DEFAULT_GCP_PROJECT_ID,
                "metric_prefixes": ["custom.googleapis.com/projectx/cloud_console"],
            },
        },
        "deployments": [
            {
                "name": "console-web",
                "provider": "gcp-cloud-run",
                "service": "console-web",
                "region": DEFAULT_REGION,
                "environment": "production",
                "desired_state": "present",
            }
        ],
        "control_plane": {
            "desired_state": "present",
            "gcp": {
                "project_id": DEFAULT_GCP_PROJECT_ID,
                "region": DEFAULT_REGION,
                "services": [
                    "run.googleapis.com",
                    "cloudresourcemanager.googleapis.com",
                    "iam.googleapis.com",
                ],
            },
            "domains": [
                {
                    "name": "anmhela.com",
                    "provider": "cloudflare",
                    "zone_id": "cf_zone_anmhela",
                    "desired_state": "present",
                    "project": "cloud-console",
                    "records": [
                        {
                            "type": "CNAME",
                            "name": "c",
                            "content": "cname.vercel-dns.com",
                            "ttl": 300,
                            "desired_state": "present",
                            "deployment_link": {
                                "project": "cloud-console",
                                "deployment_id": "console-web",
                                "host": "console-web",
                            },
                        }
                    ],
                },
                {
                    "name": "anmhela.com",
                    "provider": "vercel",
                    "desired_state": "present",
                    "project": "cloud-console",
                    "records": [
                        {
                            "type": "CNAME",
                            "name": "c",
                            "content": "cname.vercel-dns.com",
                            "ttl": 300,
                            "desired_state": "present",
                            "deployment_link": {
                                "project": "cloud-console",
                                "deployment_id": "console-web",
                                "host": "console-web",
                            },
                        }
                    ],
                },
            ],
        },
    },
    {
        "name": "access-api",
        "description": "Key issuance and policy enforcement service",
        "path": "services/access-api",
        "dependencies": [{"name": "supabase", "type": "stack"}],
        "deploy": {
            "cwd": "services/access-api",
            "command": "GOCACHE=/tmp/go-cache go test ./...",
            "health_url": "http://127.0.0.1:8090/health",
        },
        "targets": {
            "api_keys": [
                {
                    "name": "admin-key",
                    "provider": "custom",
                    "env_var": "ACCESS_API_ADMIN_KEY",
                }
            ]
        },
        "keys": {
            "managed": [
                {
                    "name": "notifications-write",
                    "service": "notifications",
                    "scope": "write",
                    "export_env": "OMNICHANNEL_API_KEY",
                }
            ]
        },
        "terraform": {
            "root": "infra/terraform/projects/access-api",
            "workspace": "dev",
        },
        "observability": {
            "gcp": {
                "project_id": DEFAULT_GCP_PROJECT_ID,
                "log_filters": [
                    'resource.type="cloud_run_revision"',
                    'labels.service_name="access-api"',
                ],
                "metric_prefixes": [
                    "run.googleapis.com",
                    "custom.googleapis.com/projectx/access_api",
                ],
            }
        },
        "deployments": [
            {
                "name": "access-api",
                "provider": "gcp-cloud-run",
                "service": "access-api",
                "region": DEFAULT_REGION,
                "environment": "production",
                "desired_state": "present",
            }
        ],
        "control_plane": {
            "desired_state": "present",
            "gcp": {
                "project_id": DEFAULT_GCP_PROJECT_ID,
                "region": DEFAULT_REGION,
                "services": [
                    "run.googleapis.com",
                    "secretmanager.googleapis.com",
                    "sqladmin.googleapis.com",
                ],
            },
        },
    },
    {
        "name": "omnichannel-api",
        "description": "Omnichannel notifications REST API",
        "path": "services/omnichannel/backend",
        "dependencies": [
            {"name": "supabase", "type": "stack"},
            {"name": "temporal", "type": "stack"},
        ],
        "deploy": {
            "cwd": "services/omnichannel/backend",
            "command": "GOCACHE=/tmp/go-cache go test ./...",
            "health_url": "http://127.0.0.1:8080/api/v1/health",
        },
        "terraform": {
            "root": "infra/terraform/projects/omnichannel-api",
            "workspace": "dev",
        },
        "observability": {
            "gcp": {
                "project_id": DEFAULT_GCP_PROJECT_ID,
                "log_filters": [
                    'resource.type="cloud_run_revision"',
                    'labels.service_name="omnichannel-api"',
                ],
                "metric_prefixes": [
                    "run.googleapis.com",
                    "custom.googleapis.com/projectx/omnichannel_api",
                ],
            }
        },
        "deployments": [
            {
                "name": "omnichannel-api",
                "provider": "gcp-cloud-run",
                "service": "omnichannel-api",
                "region": DEFAULT_REGION,
                "environment": "production",
                "desired_state": "present",
            }
        ],
        "control_plane": {
            "desired_state": "present",
            "gcp": {
                "project_id": DEFAULT_GCP_PROJECT_ID,
                "region": DEFAULT_REGION,
                "services": [
                    "run.googleapis.com",
                    "secretmanager.googleapis.com",
                ],
            },
        },
    },
    {
        "name": "omnichannel-worker",
        "description": "Temporal worker for omnichannel notification jobs",
        "path": "services/omnichannel/backend",
        "dependencies": [
            {"name": "supabase", "type": "stack"},
            {"name": "temporal", "type": "stack"},
        ],
        "deploy": {
            "cwd": "services/omnichannel/backend",
            "command": "GOCACHE=/tmp/go-cache go test ./...",
        },
        "terraform": {
            "root": "infra/terraform/projects/omnichannel-worker",
            "workspace": "dev",
        },
        "observability": {
            "gcp": {
                "project_id": DEFAULT_GCP_PROJECT_ID,
                "log_filters": [
                    'resource.type="cloud_run_revision"',
                    'labels.service_name="omnichannel-worker"',
                ],
                "metric_prefixes": [
                    "run.googleapis.com",
                    "custom.googleapis.com/projectx/omnichannel_worker",
                ],
            }
        },
        "deployments": [
            {
                "name": "omnichannel-worker",
                "provider": "gcp-cloud-run",
                "service": "omnichannel-worker",
                "region": DEFAULT_REGION,
                "environment": "production",
                "desired_state": "present",
            }
        ],
        "control_plane": {
            "desired_state": "present",
            "gcp": {
                "project_id": DEFAULT_GCP_PROJECT_ID,
                "region": DEFAULT_REGION,
                "services": [
                    "run.googleapis.com",
                    "secretmanager.googleapis.com",
                ],
            },
        },
    },
]

ACCOUNT_SPECS: list[dict[str, Any]] = [
    {
        "name": "default",
        "secrets": [
            {
                "name": "access-api-admin-key",
                "desired_state": "present",
                "source_env": "ACCESS_API_ADMIN_KEY",
                "shares": [
                    {
                        "platform": "gcp",
                        "project_id": DEFAULT_GCP_PROJECT_ID,
                        "name": "access-api-admin-key",
                        "target_type": "project",
                        "target_id": "access-api",
                        "project": "access-api",
                    },
                    {
                        "platform": "gcp",
                        "project_id": DEFAULT_GCP_PROJECT_ID,
                        "name": "console-access-api-key",
                        "target_type": "project",
                        "target_id": "cloud-console",
                        "project": "cloud-console",
                    },
                    {
                        "platform": "vercel",
                        "project_id": "cloud-console",
                        "name": "ACCESS_API_ADMIN_KEY",
                        "target_type": "application",
                        "target_id": "console-web",
                        "project": "cloud-console",
                        "environments": ["development", "preview", "production"],
                    },
                    {
                        "platform": "gcp",
                        "project_id": DEFAULT_GCP_PROJECT_ID,
                        "name": "template-studio-access-api-key",
                        "target_type": "deployment",
                        "target_id": "template-studio",
                        "project": "omnichannel-api",
                    },
                ],
            }
        ],
    }
]


def _load_integration_overrides() -> dict[str, list[dict[str, Any]]]:
    if not INTEGRATION_OVERRIDES_PATH.exists():
        return {}

    try:
        payload = json.loads(INTEGRATION_OVERRIDES_PATH.read_text(encoding="utf-8"))
    except json.JSONDecodeError:
        return {}

    project_entries = payload.get("projects", [])
    if not isinstance(project_entries, list):
        return {}

    overrides: dict[str, list[dict[str, Any]]] = {}
    for project_entry in project_entries:
        if not isinstance(project_entry, dict):
            continue
        name = str(project_entry.get("name", "")).strip()
        integrations = project_entry.get("integrations", [])
        if not name or not isinstance(integrations, list):
            continue

        sanitized: list[dict[str, Any]] = []
        for item in integrations:
            if not isinstance(item, dict):
                continue
            slug = str(item.get("slug", "")).strip()
            if not slug:
                continue
            cleaned: dict[str, Any] = {"slug": slug}
            for key in ("name", "provider", "created_at", "updated_at"):
                value = item.get(key)
                if isinstance(value, str) and value.strip():
                    cleaned[key] = value.strip()
            if isinstance(item.get("enabled"), bool):
                cleaned["enabled"] = item["enabled"]
            config = item.get("config")
            if isinstance(config, dict):
                normalized_config: dict[str, str] = {}
                for cfg_key, cfg_value in config.items():
                    key_str = str(cfg_key).strip()
                    if not key_str:
                        continue
                    normalized_config[key_str] = str(cfg_value)
                if normalized_config:
                    cleaned["config"] = normalized_config
            sanitized.append(cleaned)

        overrides[name] = sanitized
    return overrides


def _specs_with_overrides() -> list[dict[str, Any]]:
    overrides = _load_integration_overrides()
    specs: list[dict[str, Any]] = []
    for spec in PROJECT_SPECS:
        enriched = deepcopy(spec)
        if spec["name"] in overrides:
            enriched["integrations"] = overrides[spec["name"]]
        specs.append(enriched)
    return specs


def _materialize_platform_projects() -> dict[str, Any]:
    projects: list[dict[str, Any]] = []
    for spec in _specs_with_overrides():
        entry: dict[str, Any] = {
            "name": spec["name"],
            "description": spec["description"],
            "path": spec["path"],
            "dependencies": deepcopy(spec.get("dependencies", [])),
            "deploy": deepcopy(spec["deploy"]),
            "deployments": deepcopy(spec.get("deployments", [])),
        }
        for optional_key in ("integrations", "targets", "keys", "terraform", "observability"):
            if optional_key in spec:
                entry[optional_key] = deepcopy(spec[optional_key])
        projects.append(entry)
    return {"version": VERSION, "projects": projects}


def _materialize_control_plane() -> dict[str, Any]:
    accounts: list[dict[str, Any]] = []
    for account in ACCOUNT_SPECS:
        account_entry = {"name": account["name"]}
        if account.get("secrets"):
            account_entry["secrets"] = deepcopy(account["secrets"])
        accounts.append(account_entry)

    projects: list[dict[str, Any]] = []
    for spec in _specs_with_overrides():
        control = spec.get("control_plane", {})
        project_entry: dict[str, Any] = {
            "name": spec["name"],
            "desired_state": control.get("desired_state", "present"),
            "gcp": deepcopy(control.get("gcp", {})),
            "deployments": deepcopy(spec.get("deployments", [])),
        }
        if control.get("domains"):
            project_entry["domains"] = deepcopy(control["domains"])
        if control.get("secrets"):
            project_entry["secrets"] = deepcopy(control["secrets"])
        projects.append(project_entry)
    return {"version": VERSION, "accounts": accounts, "projects": projects}


def _materialize_deployment_catalog() -> dict[str, Any]:
    deployments: list[dict[str, Any]] = []
    for spec in _specs_with_overrides():
        for deployment in spec.get("deployments", []):
            entry = deepcopy(deployment)
            entry["project"] = spec["name"]
            entry["project_path"] = spec["path"]
            deployments.append(entry)
    return {"version": VERSION, "deployments": deployments}


def materialized_documents() -> dict[str, dict[str, Any]]:
    platform_projects = _materialize_platform_projects()
    control_plane = _materialize_control_plane()
    deployment_catalog = _materialize_deployment_catalog()
    return {
        "platform.projects.json": platform_projects,
        "platform.projects.example.json": platform_projects,
        "platform.controlplane.json": control_plane,
        "platform.controlplane.example.json": control_plane,
        "infra/platform/deployments.generated.json": deployment_catalog,
    }
