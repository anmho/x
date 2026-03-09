#!/usr/bin/env node

import fs from 'node:fs/promises';
import path from 'node:path';
import process from 'node:process';

const DEFAULT_INPUT = 'docs/backlog/linear-repo-analysis-tickets.json';
const LINEAR_API_URL = process.env.LINEAR_API_URL || 'https://api.linear.app/graphql';

function parseArgs(argv) {
  const args = {
    input: DEFAULT_INPUT,
    dryRun: false,
    teamId: process.env.LINEAR_TEAM_ID,
    teamKey: process.env.LINEAR_TEAM_KEY,
  };

  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (arg === '--dry-run') {
      args.dryRun = true;
      continue;
    }
    if (arg === '--input') {
      args.input = argv[i + 1];
      i += 1;
      continue;
    }
    if (arg === '--team-id') {
      args.teamId = argv[i + 1];
      i += 1;
      continue;
    }
    if (arg === '--team-key') {
      args.teamKey = argv[i + 1];
      i += 1;
      continue;
    }
    if (arg === '--help' || arg === '-h') {
      printHelp();
      process.exit(0);
    }
    throw new Error(`unknown argument: ${arg}`);
  }

  return args;
}

function printHelp() {
  console.log(`Usage:
  node scripts/linear/create-issues.mjs [--input <path>] [--team-id <id> | --team-key <key>] [--dry-run]

Environment variables:
  LINEAR_API_KEY   required unless --dry-run
  LINEAR_TEAM_ID   optional if team provided in payload or via --team-key
  LINEAR_TEAM_KEY  optional fallback for team lookup by key
  LINEAR_API_URL   optional (default: https://api.linear.app/graphql)

Examples:
  node scripts/linear/create-issues.mjs --dry-run
  LINEAR_API_KEY=lin_api_xxx LINEAR_TEAM_KEY=ENG node scripts/linear/create-issues.mjs
`);
}

function validatePayload(payload) {
  if (!payload || typeof payload !== 'object') {
    throw new Error('payload must be a JSON object');
  }

  if (!Array.isArray(payload.tickets) || payload.tickets.length === 0) {
    throw new Error('payload.tickets must be a non-empty array');
  }

  for (const [index, ticket] of payload.tickets.entries()) {
    if (!ticket || typeof ticket !== 'object') {
      throw new Error(`ticket[${index}] must be an object`);
    }
    if (typeof ticket.title !== 'string' || ticket.title.trim() === '') {
      throw new Error(`ticket[${index}].title is required`);
    }
    if (typeof ticket.description !== 'string' || ticket.description.trim() === '') {
      throw new Error(`ticket[${index}].description is required`);
    }
    if (ticket.priority !== undefined) {
      const priority = Number(ticket.priority);
      if (!Number.isInteger(priority) || priority < 0 || priority > 4) {
        throw new Error(`ticket[${index}].priority must be an integer from 0..4`);
      }
    }
  }
}

async function linearRequest(apiKey, query, variables = {}) {
  const response = await fetch(LINEAR_API_URL, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: apiKey,
    },
    body: JSON.stringify({ query, variables }),
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(`linear request failed (${response.status}): ${text}`);
  }

  const body = await response.json();
  if (Array.isArray(body.errors) && body.errors.length > 0) {
    const message = body.errors.map((err) => err.message || JSON.stringify(err)).join('; ');
    throw new Error(`linear graphql error: ${message}`);
  }

  return body.data;
}

async function resolveTeamId(apiKey, teamKey) {
  const query = `
    query Teams {
      teams {
        nodes {
          id
          key
          name
        }
      }
    }
  `;

  const data = await linearRequest(apiKey, query);
  const teams = data?.teams?.nodes || [];
  const wanted = String(teamKey).toLowerCase();
  const match = teams.find((team) => String(team.key).toLowerCase() === wanted);

  if (!match) {
    const available = teams.map((team) => team.key).filter(Boolean).join(', ');
    throw new Error(`team key '${teamKey}' not found. Available team keys: ${available || '(none)'}`);
  }

  return match.id;
}

function toIssueInput(ticket, teamId, parentId = null) {
  const input = {
    teamId,
    title: ticket.title,
    description: ticket.description,
  };

  if (parentId) {
    input.parentId = parentId;
  }

  if (ticket.priority !== undefined) {
    input.priority = Number(ticket.priority);
  }

  if (typeof ticket.estimate === 'number') {
    input.estimate = ticket.estimate;
  }

  if (typeof ticket.projectId === 'string' && ticket.projectId.trim() !== '') {
    input.projectId = ticket.projectId;
  }

  if (typeof ticket.stateId === 'string' && ticket.stateId.trim() !== '') {
    input.stateId = ticket.stateId;
  }

  return input;
}

function printDryRun(payload, teamId, teamKey) {
  const teamLabel = teamId ? `teamId=${teamId}` : teamKey ? `teamKey=${teamKey}` : 'team=(unset)';
  console.log(`Dry run: ${payload.tickets.length} tickets ready (${teamLabel})`);

  payload.tickets.forEach((ticket, index) => {
    const priority = ticket.priority === undefined ? 'unset' : String(ticket.priority);
    console.log(`\n[${index + 1}] ${ticket.title}`);
    console.log(`priority=${priority}`);
  });
}

async function createIssue(apiKey, input) {
  const mutation = `
    mutation IssueCreate($input: IssueCreateInput!) {
      issueCreate(input: $input) {
        success
        issue {
          id
          identifier
          title
          url
        }
      }
    }
  `;

  const data = await linearRequest(apiKey, mutation, { input });
  const result = data?.issueCreate;
  if (!result?.success || !result?.issue) {
    throw new Error(`issueCreate returned an unexpected response: ${JSON.stringify(data)}`);
  }
  return result.issue;
}

async function main() {
  const args = parseArgs(process.argv.slice(2));
  const inputPath = path.resolve(process.cwd(), args.input);

  const payloadText = await fs.readFile(inputPath, 'utf8');
  const payload = JSON.parse(payloadText);
  validatePayload(payload);

  const teamIdFromPayload = typeof payload.teamId === 'string' ? payload.teamId : undefined;
  const teamKeyFromPayload = typeof payload.teamKey === 'string' ? payload.teamKey : undefined;

  const teamId = args.teamId || teamIdFromPayload;
  const teamKey = args.teamKey || teamKeyFromPayload;

  if (args.dryRun) {
    printDryRun(payload, teamId, teamKey);
    return;
  }

  const apiKey = process.env.LINEAR_API_KEY;
  if (!apiKey) {
    throw new Error('LINEAR_API_KEY is required when not using --dry-run');
  }

  let resolvedTeamId = teamId;
  if (!resolvedTeamId) {
    if (!teamKey) {
      throw new Error('team is not set: provide --team-id, LINEAR_TEAM_ID, --team-key, LINEAR_TEAM_KEY, or payload teamKey/teamId');
    }
    resolvedTeamId = await resolveTeamId(apiKey, teamKey);
  }

  const parentRef = 'parent';
  let parentIssue = null;

  console.log(`Creating ${payload.tickets.length} Linear issues in team ${resolvedTeamId}...`);
  for (const [index, ticket] of payload.tickets.entries()) {
    const isChild = ticket.parentRef === parentRef;
    const parentId = isChild && parentIssue ? parentIssue.id : null;
    const input = toIssueInput(ticket, resolvedTeamId, parentId);
    const issue = await createIssue(apiKey, input);
    if (!parentIssue) {
      parentIssue = issue;
    }
    console.log(`[${index + 1}/${payload.tickets.length}] ${issue.identifier}: ${issue.title}${isChild ? ' (child)' : ' (parent)'}`);
    if (issue.url) {
      console.log(`  ${issue.url}`);
    }
  }

  if (parentIssue) {
    console.log(`\nParent ticket: ${parentIssue.identifier} — use in PR: Linear: [${parentIssue.identifier}](${parentIssue.url})`);
  }
  console.log('Done.');
}

main().catch((error) => {
  console.error(`Error: ${error.message}`);
  process.exit(1);
});
