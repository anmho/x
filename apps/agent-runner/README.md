# agent-runner

`apps/agent-runner` is the Bun/TypeScript runtime worker for Project X agent execution.

Current scope:

- Bun-native worker process
- ConnectRPC client to `services/agent-control-api`
- Claude-first runtime adapter shape
- message-plus-resources materialization so any upstream source can provide a canonical agent message and attachments

Current invocation modes:

- `AGENT_RUN_ID=<uuid>`: fetch the run from the control plane with ConnectRPC, materialize its message/resources, and execute it
- `AGENT_MESSAGE=...`: local-only execution without a control-plane lookup

Environment:

- `AGENT_CONTROL_BASE_URL` default `http://localhost:8090`
- `AGENT_RUN_ID` optional control-plane run id
- `AGENT_MESSAGE` optional direct message for local-only execution
- `AGENT_CWD` optional working directory, defaults to current directory
- `AGENT_MCP_CONFIG_PATH` optional path to an MCP config JSON file

This is the worker foundation for the later leased/session-based data plane. It is not yet the full worker-assignment protocol.
