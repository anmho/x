# Optional Skills

When additional tooling is needed for MCP or CLI workflows, add the relevant skill here so future agents know what to launch or install.

- `schpet-linear-cli` — wraps `schpet/linear-cli` so you can run commands such as `linear issue list` or `linear issue create` directly from the repo shell. Install it with the skill installer (see `.cursor/skills/.system/skill-installer/SKILL.md`), then mention it in `AGENTS.md` and reference it via the `skill` command when making MCP calls.

If you need a new skill but the repo-level installer is missing, document the source path, installation steps, and any required environment variables before calling `gh` or `linear` directly.
