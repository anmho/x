# Agent Mistake Log

This file is the permanent mistake memory for repository agents.

## Rules

1. Add a new entry immediately when a mistake is discovered.
2. Do not rewrite history; append corrections as new entries.
3. Every entry must include a concrete prevention step and verification.
4. Use UTC timestamps in ISO-8601 format.

## Entry Template

```text
## <YYYY-MM-DDTHH:MM:SSZ> - <short title>
- What happened:
- Root cause:
- How to avoid it:
- What to do instead:
- Verification:
```

## Entries

## 2026-03-18T07:48:00Z - Sent a string instead of a boolean to `gh api`
- What happened: I retried the branch-protection update with correctly quoted array flags but used `-f strict=false`, and GitHub rejected the request because `strict` arrived as the string `"false"` instead of a boolean.
- Root cause: I used the generic form field flag without checking whether this endpoint required typed boolean serialization.
- How to avoid it: For `gh api` requests that need booleans or numbers, prefer typed flags such as `-F strict=false` or pass a JSON body with explicit types.
- What to do instead: Use `-F strict=false` alongside the quoted `contexts[]` fields when updating required status checks.
- Verification: After adding this entry, I reran the branch-protection request with a boolean-typed `strict=false` parameter.

## 2026-03-18T07:47:00Z - Forgot to quote `gh api` array parameters in zsh
- What happened: I tried to update branch protection with `gh api ... -f contexts[]=...` in zsh without quoting the array-style flags, and the shell rejected the command with `no matches found`.
- Root cause: I treated the GitHub CLI flag syntax as shell-safe even though unquoted brackets trigger zsh glob parsing.
- How to avoid it: Quote every `-f 'contexts[]=...'` style argument in zsh when calling `gh api`, especially for GitHub array parameters.
- What to do instead: Use `-f 'contexts[]=Affected Docs'` style quoted flags or disable globbing for the command.
- Verification: After adding this entry, I reran the branch-protection update with quoted array arguments instead of bare bracket syntax.

## 2026-03-18T07:44:00Z - Repeated parallel git write after documenting the rule
- What happened: While staging the Mintlify docs follow-up, I again launched `git add` and `git commit` in parallel, which recreated the `.git/index.lock` commit failure I had just documented.
- Root cause: I fell back to the multi-tool wrapper out of habit instead of switching to strictly sequential Git write steps after logging the first mistake.
- How to avoid it: Once a Git-write sequencing mistake is logged, stop using the parallel wrapper for any subsequent `git add`, `git commit`, or `git push` calls in the same task.
- What to do instead: Run repository-mutating Git commands one at a time and wait for each to complete before starting the next.
- Verification: After adding this entry, I resumed the remaining Git write flow only with sequential single-command executions.

## 2026-03-18T07:39:30Z - Created overlapping collaboration tickets before finishing a backlog search
- What happened: I created new collaboration/mailbox tickets for the MCP mailbox work and only afterward discovered overlapping existing collaboration issues during the required reconciliation scan.
- Root cause: I moved from architecture discussion into execution too quickly and skipped the focused Linear search that should happen before ticket creation.
- How to avoid it: Before creating any ticket for a new workstream, search Linear using the exact domain nouns from the task and reconcile open issues first.
- What to do instead: Reuse the existing issue when it already covers the scope, and create only the missing delta.
- Verification: Ran a focused collaboration/mailbox issue search after implementation, documented the overlap in the active ExecPlan, and constrained the current slice to the mailbox MCP core only.

## 2026-03-18T08:02:00Z - Ran `git add` and `git commit` in parallel
- What happened: I used the parallel tool to launch `git add` and `git commit` at the same time while trying to record the final ExecPlan update, and `git commit` failed on `.git/index.lock`.
- Root cause: I treated dependent git write operations as parallel-safe even though both mutate the repository index.
- How to avoid it: Never parallelize git commands that write repository state; only parallelize independent read-only inspection commands.
- What to do instead: Run `git add`, wait for it to finish, then run `git commit` sequentially.
- Verification: After this entry, I re-ran the plan-only commit flow sequentially instead of through the parallel wrapper.

## 2026-03-18T07:23:50Z - Computed extension workspace root one directory too high
- What happened: I wrote the new Chrome extension build and verify scripts with `workspaceRoot` resolving three directories above `apps/linear-ticket-sidepanel`, which made the build try to create `/Users/andrewho/repos/projects/dist` outside the repository and caused Nx build/verify failures.
- Root cause: I translated the folder depth from memory instead of checking the actual `apps/<app>/scripts` path relationship back to the repo root.
- How to avoid it: For every new app-local script that computes repository-relative paths, explicitly derive the expected absolute target once and verify it against the current repo root before running Nx targets.
- What to do instead: Use `path.resolve(appRoot, "..", "..")` for app-local scripts under `apps/<name>/scripts`, and make verify targets depend on build so missing dist output cannot masquerade as a separate issue.
- Verification: Updated both extension scripts to resolve the repo root two levels above `appRoot`, added `dependsOn: [\"build\"]` for `verify` and `preflight`, and re-ran the Nx targets.

## 2026-03-18T07:16:12Z - Started agent-control with ad hoc REST and fetch helpers instead of the repo RPC contract
- What happened: I initially added a hand-written HTTP/JSON client/server surface for `services/agent-control-api` and `apps/cloud-console` before reconfirming that this repository expects control-plane APIs to be protobuf-first ConnectRPC services with generated SDKs.
- Root cause: I optimized for speed on the MVP path and failed to stop at the architecture boundary to verify whether the control plane should follow the existing ConnectRPC/Buf SDK pattern.
- How to avoid it: For any new control-plane or service-to-UI contract in this repository, inspect existing protobuf/Connect patterns and confirm SDK generation/consumption before creating request handlers or frontend clients by hand.
- What to do instead: Define the `.proto` contract first, generate Go/TS stubs, implement the server against generated interfaces, and make the UI consume the generated Connect client or a thin wrapper over it.
- Verification: Added `services/agent-control-api/proto/agentcontrol/v1/agent_control.proto`, generated Go/TS Connect stubs, switched the server toward `agentcontrol.v1.AgentControlService`, and updated Mintlify docs to describe the ConnectRPC control-plane model.

## 2026-03-18T06:24:00Z - Docs verifier missed tabbed Mintlify navigation structure
- What happened: I added `scripts/ci/verify_docs_config.mjs`, but the first implementation only handled array-style `navigation` and reported `0 nav page entries` for the tabbed `docs.json` shape.
- Root cause: I implemented the page collector against an older Mintlify config shape and did not verify against the actual nested `navigation.tabs[].groups[].pages[]` structure before relying on output.
- How to avoid it: For schema-dependent validation scripts, run a real sample parse from current repository config and assert non-zero expected values in the same edit cycle.
- What to do instead: Implement recursive traversal that handles `navigation`, `tabs`, `groups`, and `pages` nodes and verify expected page count before finalizing.
- Verification: Updated collector recursion to traverse tabbed/grouped structures and re-ran `npm run verify:docs`, which now reports `docs config verified: docs.json (4 nav page entries)`.

## 2026-03-10T09:31:51Z - Touched unrelated dirty doc content while patching Greptile guidance
- What happened: I patched `docs/runbooks/cursor-cloud-parity.md` for Greptile guidance and accidentally removed a pre-existing duplicate `LINEAR_API_KEY` row from the user's dirty worktree.
- Root cause: I applied the patch against the file without first checking whether adjacent hunks already had uncommitted user edits that needed to be preserved verbatim.
- How to avoid it: When editing a dirty file, inspect the local diff around the target hunk before patching and constrain changes to the smallest possible context.
- What to do instead: Re-read the touched hunk after patching, restore any unrelated local edits immediately, and keep the task diff scoped to the requested topic only.
- Verification: Restored the `LINEAR_API_KEY` rows, rechecked the focused diff for `docs/runbooks/cursor-cloud-parity.md`, and kept only the Greptile-specific additions.

## 2026-03-10T09:28:00Z - Ran retired docs verify mode during validation
- What happened: I ran the retired docs-mode verify command during validation even though that command path had already been removed from expected workflow usage.
- Root cause: I followed stale validation muscle memory from earlier plan steps instead of reconfirming current approved command surfaces before execution.
- How to avoid it: Before running validation commands, re-check the active command surface in `package.json`, `scripts/verify`, and current policy docs for deprecated paths.
- What to do instead: Use only current workflow/verification commands and remove deprecated command references immediately when discovered.
- Verification: Created `ANM-128`, removed `docs` mode from `scripts/verify`, removed the retired docs-verify alias from `package.json`, and scrubbed stale command references.

## 2026-03-10T08:44:00Z - Introduced duplicate constant while patching docs verifier
- What happened: I initially edited `scripts/verify` and declared `const base` twice inside the Node docs-check snippet, which would have caused a runtime syntax error.
- Root cause: I inserted compatibility logic quickly and missed a redundant pre-existing declaration during the same patch.
- How to avoid it: After each inline script patch, re-read the full modified block for duplicate declarations before moving to the next file edit.
- What to do instead: Apply the compatibility insertion in one pass and run an immediate focused diff/scan on the edited function.
- Verification: Removed the duplicate declaration and confirmed the `verify_docs` block now parses with a single `const base` declaration.

## 2026-03-10T08:41:53Z - Added non-existent agent schema path during docs refresh
- What happened: I updated `docs/repository-architecture-deep-dive.md` to reference `agents/blueprints/agent.schema.json`, then discovered that path does not exist in this repository snapshot.
- Root cause: I carried over an outdated repository layout assumption without verifying the actual `agents/` tree first.
- How to avoid it: Before adding any path reference in docs, verify existence with `ls`/`find` in the target directory.
- What to do instead: Resolve concrete paths from current filesystem state, then write doc references.
- Verification: Rechecked `agents/` with `find`, removed the invalid path reference, and replaced it with a valid statement about current contract anchors.

## 2026-03-10T06:17:16Z - Forgot to quote URL containing query string in zsh command
- What happened: I ran `curl http://localhost:3000/api/domains?project=cloud-console` without quotes and zsh treated `?` as glob syntax, causing `no matches found`.
- Root cause: I skipped shell-escaping for a URL with special characters.
- How to avoid it: Always wrap URLs containing `?`, `&`, or `*` in single quotes when running shell commands.
- What to do instead: Use `curl 'http://localhost:3000/api/domains?project=cloud-console'`.
- Verification: Re-ran with quoted URL and reached network connectivity check correctly.

## 2026-03-18T07:36:06Z - Linked new Linear child issues to the wrong parent on first pass
- What happened: I created the mailbox MCP child tickets with `parentId` pointing at `ANM-175` instead of the newly created parent issue `ANM-181`.
- Root cause: I pre-filled the expected parent identifier before the actual Linear create responses returned and reused the wrong placeholder.
- How to avoid it: Always wait for the created parent issue ID before attaching child issues; never assume the final Linear identifier.
- What to do instead: Create the parent first, capture the real returned ID, then create or update children against that exact ID.
- Verification: Updated `ANM-182`, `ANM-183`, and `ANM-184` so each now has `parentId: ANM-181`.

## 2026-03-18T07:06:08Z - Patched services/mcp/project.json against a stale snapshot
- What happened: I attempted an `apply_patch` against `services/mcp/project.json` using an outdated command string and the patch failed because the file had already changed to a different `MCP_ROOT` value.
- Root cause: I reused an earlier read of the file instead of re-reading the live contents immediately before editing.
- How to avoid it: Before every patch, re-open the exact file if I have reason to believe another agent or earlier edit may have changed it.
- What to do instead: Read the current file, patch only the live lines, then validate with a targeted diff.
- Verification: Re-read `services/mcp/project.json`, applied the corrected patch successfully, and validated the updated Nx build outputs plus `./platform mcp` smoke test afterward.

## 2026-03-18T07:12:27Z - Used `nx affected --projects` as if it were a hard filter for run-commands targets
- What happened: I initially rewired CI jobs to use `npx nx affected -t <target> --projects=...`, but Nx treated `--projects=...` as an extra task argument and forwarded it into `nx:run-commands` targets such as `go build`, breaking the command invocation.
- Root cause: I assumed `affected` scoped target execution the same way `run-many` does and did not verify how Nx v22 handles `--projects` with run-commands executors.
- Preventive rule/check added: For scoped affected CI, first compute the project list with `nx show projects --affected --withTarget=... --projects=...`, then call `nx run-many` on the resulting list. Do not rely on `nx affected --projects=...` for run-commands targets.
- Verification: Replaced the workflow logic to use `nx show projects ... --sep=,` plus `nx run-many`, and local spot checks succeeded for docs and mcp scopes without forwarding stray CLI flags into underlying commands.

## 2026-03-10T06:06:30Z - Used wrong frontend root path while inspecting domains implementation
- What happened: I attempted to inspect and patch `services/omnichannel/frontend/...` for domains files, but this workspace keeps the active frontend under `apps/omnichannel/frontend/...`.
- Root cause: I reused an older repository layout assumption from prior tasks without confirming current paths first.
- How to avoid it: Before reading/editing files, verify path ownership with `rg --files` or `ls` in the expected subtree.
- What to do instead: Resolve target files from current search results first, then run edits only on confirmed paths.
- Verification: Re-ran inspection commands on `apps/omnichannel/frontend/...` and completed changes there.

## 2026-03-09T02:00:00Z - Used scripts/verify instead of Nx targets for app checks
- What happened: CI and scripts/verify apps used custom shell logic (nx run sdk:generate-es + npm run build) instead of Nx targets.
- Root cause: Did not treat Nx as the single source of truth for build/verify orchestration.
- How to avoid it: Prefer Nx targets for all build, test, and verify steps. Use nx run-many -t <target> instead of ad-hoc scripts.
- What to do instead: Define verify/build as Nx targets in project.json; CI runs nx run-many directly. scripts/verify becomes a thin nx wrapper or is removed.
- Verification: Migrated verify_apps to nx run-many -t build; CI apps job now runs nx directly. Ticket: docs/backlog/verify-nx-targets-tickets.json for full migration.

## 2026-03-09T02:00:00Z - Did not push and check workflow logs after changes
- What happened: Made multiple changes (CI fix, AGENTS.md, etc.) but did not push until the user asked. Did not run gh run list / gh run view --log-failed to verify CI.
- Root cause: Focused on edits without completing the loop: commit → push → verify.
- How to avoid it: After every logical change, commit, push, and check workflow status. Per AGENTS.md Ralph loop: complete the full loop.
- What to do instead: Immediately after committing: git push; gh run list --limit 3; if failure, gh run view <id> --log-failed. Do not stop with uncommitted/unpushed work.
- Verification: AGENTS.md Ralph Loop Friendly section; this entry.

## 2026-03-08T22:38:06Z - Used project-style ticket prefix instead of team-key issue identifiers
- What happened: I initially tracked execution items as `XPLAT-*` in the ExecPlan and did not publish corresponding Linear issues before execution.
- Root cause: I treated the project label as the issue identifier prefix and deferred external ticket creation.
- Preventive rule/check added: Before editing plan ticket IDs, confirm team key vs project naming. Use Linear MCP to create tickets at task start.
- Verification: Updated the plan tracker to `ANM`-scoped placeholders, prepared `docs/backlog/x-platform-repo-hygiene-tickets.json`, and executed the Linear dry-run command with `--team-key ANM`.

## 2026-03-08T22:29:00Z - Used compile-only checks instead of task-scoped validation commands
- What happened: I validated frontend changes mainly with `npm run build` without running broader task-scoped validation/workflow checks.
- Root cause: I optimized for fast compile feedback and skipped validation commands that better matched the task scope.
- Preventive rule/check added: For validation claims, always run and report task-scoped workflow/verification commands (`platform` workflow commands, `scripts/verify`, `scripts/deploy-preflight`) before finalizing.
- Verification: Superseded by `ANM-126` policy update in `AGENTS.md` and `docs/runbooks/platform-cli-workflow.md` to remove local-runtime-first requirements.

## 2026-03-08T22:14:08Z - Committed in wrong repository
- What happened: I tried to commit frontend route changes from the top-level repository, but the files live in a nested Git repository at `services/omnichannel/frontend`.
- Root cause: I did not verify repository boundaries before staging and committing.
- Preventive rule/check added: Before every commit, run `git rev-parse --show-toplevel` in the working directory that contains the edited files.
- Verification: Confirmed separate top-level paths for `/Users/andrewho/repos/projects/x` and `/Users/andrewho/repos/projects/x/services/omnichannel/frontend`.

## 2026-03-08T22:27:59Z - Unquoted bracket path in git add
- What happened: I ran `git add` with an unquoted path containing `[domain]`, and zsh expanded it as a glob, causing `no matches found`.
- Root cause: I forgot shell globbing behavior for dynamic-route paths.
- Preventive rule/check added: Always quote any path containing `[` or `]` when running shell commands.
- Verification: Retried staging using quoted dynamic-route paths.

## 2026-03-08T22:30:08Z - Invalid background smoke-check command
- What happened: I attempted a one-liner background startup/smoke check for `control-plane serve` that failed (`nice(5) failed`) and did not produce a valid process id for cleanup.
- Root cause: I used a brittle shell one-liner instead of a deterministic start/poll/stop sequence.
- Preventive rule/check added: Use explicit command steps (`start`, `poll`, `stop`) and verify PID capture before issuing `kill`.
- Verification: Subsequent validation relies on deterministic compile/test checks; runtime smoke is marked as blocked by sandbox.

## 2026-03-08T22:15:06Z - Missed one dark shell class during bulk replacement
- What happened: My initial automated replacement left `services/omnichannel/frontend/app/domains/page.tsx` still using `bg-[#0a0a0a]`.
- Root cause: I validated only a subset of grep output and did not run a final zero-match check for the hardcoded color pattern.
- Preventive rule/check added: After any bulk class migration, run a repository-wide `rg` for the old pattern and require zero matches before moving on.
- Verification: Re-ran `rg -n "bg-\\[#0a0a0a\\]" apps/cloud-console/app services/omnichannel/frontend/app`; zero results returned.

## 2026-03-08T22:25:53Z - Shell interpreted literal `<plan>` during validation
- What happened: I ran `rg -n "For every task, ensure an ExecPlan exists at `plans/<plan>.md`" AGENTS.md`, and zsh interpreted `` `...` `` plus `<plan>` as command/input redirection, producing `no such file or directory: plan`.
- Root cause: I used backticks inside double quotes instead of shell-safe single quoting for a literal Markdown snippet.
- Preventive rule/check added: When searching for literal strings that include backticks or `<...>`, wrap the entire pattern in single quotes.
- Verification: Re-ran validation with `rg -n 'For every task, ensure an ExecPlan exists at `plans/<plan>.md`' AGENTS.md` and received the expected single match.

## 2026-03-08T22:28:07Z - Ran TypeScript check without selecting project
- What happened: I executed `npm --prefix services/omnichannel/frontend exec tsc --noEmit` from the repo root and got the TypeScript CLI help output instead of a project typecheck.
- Root cause: I assumed `npm --prefix ... exec tsc` would always resolve the project context without explicitly running in the package directory.
- Preventive rule/check added: For frontend typechecks, run `./node_modules/.bin/tsc --noEmit` from the package directory (or provide `-p <tsconfig>` explicitly).
- Verification: Re-ran from `services/omnichannel/frontend` with `./node_modules/.bin/tsc --noEmit`; command exited successfully.

## 2026-03-08T22:28:48Z - Repeated shell quoting error while searching for backtick text
- What happened: I ran an `rg` pattern in double quotes that contained `` `ANM-28` ``, and zsh tried to execute `ANM-28` as a command.
- Root cause: I reused unsafe double-quoted search syntax despite already documenting this shell behavior.
- Preventive rule/check added: Treat all `rg` patterns containing backticks as single-quoted literals, no exceptions.
- Verification: Subsequent command output confirmed the shell error source and this entry was added immediately.

## 2026-03-08T22:37:52Z - Repeated backtick pattern quoting failure in rg validation
- What happened: I ran `rg` validations for updated `AGENTS.md` rules using double-quoted patterns with backticks, and zsh attempted to execute `Anmho`, `Backlog`, and `In` as commands.
- Root cause: I failed to apply the existing single-quote rule for literal Markdown text containing backticks.
- Preventive rule/check added: Before running any `rg` command, if the pattern contains backticks, require single-quoted pattern syntax and avoid shell interpolation.
- Verification: Re-ran validations using single-quoted patterns and obtained expected matches without shell command-substitution errors.
