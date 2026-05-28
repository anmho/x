# Build A Chrome Side Panel Extension For Vibecoding Linear Tickets

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

This document follows the requirements in `PLANS.md` at the repository root.

Linear linkage:

- Bootstrap issue: `ANM-170` (`https://linear.app/anmho/issue/ANM-170/medium-build-chrome-side-panel-extension-for-vibecoding-linear-tickets`)
- Current slice issue: `ANM-175` (`https://linear.app/anmho/issue/ANM-175/medium-replace-pasted-linear-api-key-setup-in-sidepanel-with-safer-low`)
- Canonical tracking project: `Project X Agent Execution` (`https://linear.app/anmho/team/ANM/projects/all`)

## Purpose / Big Picture

After this change, the repo contains a loadable Manifest V3 Chrome extension that opens in the browser side panel and gives a vibecoding-friendly workspace directly next to the current page. The first slice focuses on the essential loop: capture browser context, optionally load Linear ticket context, prepare a text-only vibe payload, and keep a clean seam for a future MCP proxy.

Observable outcome:

- Chrome can load the built extension as an unpacked extension.
- Clicking the extension action opens a side panel instead of a popup.
- The side panel shows active-tab title and URL.
- A user can save a Linear personal API key in local extension storage.
- A user can save an optional proxy URL/token in local extension storage.
- The side panel can fetch teams and assigned issues from Linear, select a ticket as vibe context, and create a new issue using the current page context.
- The side panel can copy a structured context payload and optionally send a text-only request to a proxy endpoint.

## Progress

- [x] (2026-03-18 07:14Z) Confirmed `Anmho` team and `X Platform` project prerequisites, created `ANM-170`, assigned it to `me`, recorded the Codex session UUID, and set status to `In Progress`.
- [x] (2026-03-18 07:15Z) Audited the repo for an existing extension scaffold and confirmed no tracked Chrome extension app currently exists.
- [x] (2026-03-18 07:18Z) Added `apps/linear-ticket-sidepanel` with MV3 manifest, background worker, side panel UI, build/verify scripts, README, and Nx targets.
- [x] (2026-03-18 07:22Z) Implemented page-context capture, Linear issue loading/selection, issue creation flow, context-pack copying, and an optional text-only proxy request seam.
- [x] (2026-03-18 07:24Z) Integrated extension verification into repo validation via `scripts/verify` and documented unpacked-extension usage.
- [x] (2026-03-18 07:26Z) Ran targeted validation, performed post-change audit/rescan, created follow-up tickets `ANM-174` and `ANM-175`, and updated active ticket context.
- [x] (2026-03-18 07:36Z) Added branded PNG manifest icons, restricted sidepanel activation to `linear.app`, and added low-latency Linear issue-creation open triggers via a lightweight content script.
- [x] (2026-03-18 07:44Z) Broadened the `linear.app` content-script heuristics so user actions that open create-issue or edit/open-issue flows also open the sidepanel.
- [x] (2026-03-24 07:25Z) Revalidated the extension in the current shell: targeted Nx verify still passes; repo-level `./scripts/verify apps` remains blocked until Node is upgraded to the repo minimum.
- [x] (2026-03-24 07:35Z) Added low-setup live session mode so the sidepanel can consume current `linear.app` issue/draft context without requiring a pasted API key.

## Surprises & Discoveries

- Observation: No existing browser-extension app or `manifest.json` exists in the checked-in repo state, so this request needs a fresh scaffold rather than a continuation patch.
  Evidence: repository search across `apps/`, `agents/`, `packages/`, and `services/` returned no extension artifacts outside dependencies.

- Observation: The root `scripts/verify apps` flow currently validates only `cloud-console`.
  Evidence: `scripts/verify` runs `npx nx run-many -t build --projects=cloud-console`.

- Observation: Supporting a user-specified proxy with minimal setup requires broad extension host permissions.
  Evidence: Manifest V3 network access is gated by `host_permissions`, so a future arbitrary proxy target needs explicit wildcard or known-host entries.

- Observation: Repo-level `./scripts/verify apps` is blocked in this shell by the repository's required Node version, even though the targeted extension Nx targets pass.
  Evidence: command output first reported `node version too old: v20.12.2 (required >= v24.14.0)` on 2026-03-18, and revalidation on 2026-03-24 still reports `node version too old: v22.22.0 (required >= v24.14.0)`.

- Observation: Chrome requires `sidePanel.open()` to run in response to a user action, so truly instant opening must be tied to real page interactions instead of passive route watching.
  Evidence: official `chrome.sidePanel.open()` docs say it "may only be called in response to a user action."

- Observation: Linear issue-opening surfaces are heterogeneous enough that low-latency auto-open works best from click/focus heuristics based on issue identifiers, issue links, and editor-field hints rather than a single stable selector.
  Evidence: current Linear UI exposes both `New issue` modal flows and issue-card/detail flows with different surface text and markup.

- Observation: A practical low-setup path exists before OAuth by using live `linear.app` DOM/session context for the current issue or draft, without attempting to scrape cookies or impersonate the user's session outside the active page.
  Evidence: content-script snapshotting can extract current draft/issue metadata from the active Linear tab and hand it to the sidepanel via extension messaging plus local state.

## Decision Log

- Decision: Start with a plain MV3 extension using static HTML/CSS/JS plus small Node build/verify scripts instead of introducing a frontend bundler.
  Rationale: The requested behavior is extension-specific, the repo has no existing extension toolchain to continue from, and a static implementation minimizes integration risk while still fitting Nx.
  Date/Author: 2026-03-18 / Codex

- Decision: Use a Linear personal API key stored in `chrome.storage.local` for the first slice instead of implementing OAuth immediately.
  Rationale: It keeps the first in-browser workflow functional without adding an external auth backend or secret material to the repo. OAuth can follow as a separate ticket once the side-panel UX is proven.
  Date/Author: 2026-03-18 / Codex

- Decision: Keep the first vibe workflow text-only and target a generic proxy seam instead of wiring browser-side tool execution now.
  Rationale: The user wants the extension to stay low-setup while a separate MCP proxy is still being designed. Shipping text payloads plus page/ticket context keeps the MVP testable without overcommitting to a premature proxy contract.
  Date/Author: 2026-03-18 / Codex

- Decision: Use a tiny `linear.app` content script for issue-creation clicks/focus events to open the side panel with minimal latency.
  Rationale: It preserves Chrome's user-action requirement while making the panel feel immediate on the Linear flows the user actually cares about.
  Date/Author: 2026-03-18 / Codex

- Decision: Treat issue-card clicks, issue links/identifiers, and issue-editor field focus as valid edit-flow triggers for sidepanel opening.
  Rationale: The user wants the panel to appear whenever a create or edit issue flow is being opened. These heuristics align with actual Linear usage better than only watching the `New issue` affordance.
  Date/Author: 2026-03-18 / Codex

- Decision: Add a live session-context mode before OAuth so the sidepanel can operate on the active Linear issue/draft without a stored API key.
  Rationale: It materially lowers setup cost for the user's main workflow while avoiding the unsafe cookie-scraping path and deferring the larger OAuth/proxy auth design to follow-up work.
  Date/Author: 2026-03-24 / Codex

## Outcomes & Retrospective

Outcome:

- The repo now contains a buildable unpacked Chrome extension at `dist/apps/linear-ticket-sidepanel`.
- The side panel captures the active tab, loads assigned Linear issues via personal API key, lets the user select an issue as vibe context, and can create a new Linear issue from the current page.
- The side panel can also copy a structured context payload and POST the same payload to a user-configured proxy URL for text-only vibe responses.
- The extension now has branded manifest/action icons, stays inactive off `linear.app`, and opens quickly on common Linear issue-creation interactions.
- The extension now opens on a broader set of Linear create/edit issue interactions, including issue-card and issue-link clicks plus editor-field focus.
- The sidepanel now supports a no-key live session mode on `linear.app`, using the current issue/draft context from the active page before falling back to API-key-backed issue listing.
- Targeted Nx validation passes for the extension app.

Remaining gaps:

- OAuth-based Linear auth is out of scope for this initial slice.
- Direct browser-side MCP tool execution is out of scope for this initial slice.
- Chrome Web Store packaging/publishing is out of scope for this initial slice.
- Repo-wide `./scripts/verify apps` was not fully executable in this shell because Node `v20.12.2` is below the repo minimum `v24.14.0`.
- Repo-wide `./scripts/verify apps` is still not executable in this shell because current Node `v22.22.0` is below the repo minimum `v24.14.0`.
- Follow-up tickets created: `ANM-174` for the authenticated MCP proxy contract and `ANM-175` for a safer low-setup Linear auth flow.
- The low-setup auth follow-up (`ANM-175`) is now partially delivered via live session mode; full OAuth/proxy-backed auth is still open as future work.
- Additional reprioritization: hotkey ticket `ANM-176` moved back to `Backlog`; current UI/activation polish tracked in `ANM-177`.
- Additional refinement tracked separately in `ANM-187`: broader create/edit-flow auto-open heuristics on Linear.

Lesson:

- The right MVP boundary is “context capture plus proxy seam,” not direct browser-side tool execution. That keeps the side panel testable now while leaving room for a safer agent/proxy architecture later.
- For perceived speed in Chrome extensions, tying panel opening to in-page user gestures beats polling or route-only heuristics.
- A staged auth strategy works better here than waiting for full OAuth: live session context removes setup friction for the main workflow while preserving a clean path toward safer long-term auth.

## Context And Orientation

Relevant paths:

- `apps/linear-ticket-sidepanel/`: new Chrome extension app.
- `plans/vibecoding-linear-sidepanel-extension-execplan.md`: this execution plan.
- `scripts/verify`: repo validation entrypoint that should cover the new app.

Non-obvious terms:

- "MV3": Chrome Extension Manifest V3, which uses a service worker background script instead of a persistent background page.
- "Side panel": Chrome’s extension UI surface that stays alongside the active tab and is configured via the `chrome.sidePanel` API.

## Plan Of Work

1. Scaffold a new extension app under `apps/` with a manifest, service worker, side panel page, styling, and README.
2. Implement extension state management for local settings and active-tab context.
3. Add Linear GraphQL integration for bootstrap data and issue creation.
4. Add a text-only vibe workflow that copies/sends structured page-plus-ticket context to a future proxy.
5. Add Node/Nx build and verify commands that produce a dist folder loadable by Chrome.
6. Extend repo app verification coverage to include the new extension.
7. Run validation, then perform the required audit/ticket reconciliation before marking completion.

## Concrete Steps

Run from repository root:

    npx nx run linear-ticket-sidepanel:build
    npx nx run linear-ticket-sidepanel:verify
    ./scripts/verify apps

Expected snippets:

- Build copies the extension package into `dist/apps/linear-ticket-sidepanel`.
- Verify confirms required manifest fields and packaged files are present.
- The built extension can be loaded from `dist/apps/linear-ticket-sidepanel` in Chrome’s unpacked-extension flow.

## Validation And Acceptance

Acceptance criteria:

- Extension manifest declares a side panel and required permissions.
- Background worker configures side-panel action behavior.
- Side panel persists Linear and proxy settings locally without checking secrets into git.
- Side panel can fetch teams and assigned issues and create a new Linear issue with current-tab metadata.
- Side panel can generate a structured vibe payload and optionally POST it to a user-configured proxy.
- Repo validation includes the extension in the apps verification path.

## Idempotence And Recovery

The extension app is file-based and deterministic. Re-running the build should recreate the same dist output from source files. If the UI or manifest regresses, restore the extension app files from git history, rebuild the dist output, and re-run verification.

## Artifacts And Notes

Artifacts:

- `apps/linear-ticket-sidepanel/**`
- `plans/vibecoding-linear-sidepanel-extension-execplan.md`
- `scripts/verify`

Validation evidence:

    npx nx run linear-ticket-sidepanel:build
    # built linear-ticket-sidepanel to dist/apps/linear-ticket-sidepanel

    npx nx run linear-ticket-sidepanel:verify
    # verify: linear-ticket-sidepanel checks passed

    ./scripts/verify apps
    # blocked in this shell: node version too old (v22.22.0; required >= v24.14.0)

## Interfaces And Dependencies

External interfaces:

- Chrome Extensions Manifest V3 side panel API (`chrome.sidePanel`) for the extension surface and action behavior.
- Linear GraphQL API (`https://api.linear.app/graphql`) for team lookup, assigned issues, and issue creation.
- User-configured proxy endpoint for text-only vibe requests that will later front MCP-backed tools.

Internal dependencies:

- Nx `project.json` targets for build/verify integration.
- Root `scripts/verify` for repo-level apps validation coverage.

Revision Note (2026-03-18): Initial plan created and linked to `ANM-170` for the Chrome side-panel extension slice.
Revision Note (2026-03-18): Expanded the MVP to a broader vibe side panel with a future MCP proxy seam and recorded targeted validation plus follow-up tickets.
Revision Note (2026-03-24): Revalidated the extension in the current shell and refreshed the repo-level apps-verify blocker evidence.
Revision Note (2026-03-24): Added low-setup live session mode for current `linear.app` issue/draft context under `ANM-175`.
