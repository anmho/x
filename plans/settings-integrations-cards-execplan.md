# Add Settings Integrations Cards For PostHog, Stripe, And Vercel

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

This document follows the requirements in `PLANS.md` at the repository root.

Linear linkage:

- Primary issue: `ANM-58` (`https://linear.app/anmho/issue/ANM-58/minor-add-settings-integrations-cards-for-posthog-stripe-and-vercel`)
- Canonical tracking project: `Project X Agent Execution` (`https://linear.app/anmho/team/ANM/projects/all`)

## Purpose / Big Picture

After this change, users visiting `Settings > Integrations` can see explicit integration entries for Google Cloud, PostHog, Stripe, and Vercel instead of a single Google card plus a generic "coming soon" placeholder. This makes available integrations discoverable and gives each provider a clear status and action affordance.

Observable outcome:

- `/settings?tab=integrations` renders four integration cards.
- Google Cloud retains its existing connect/disconnect behavior.
- PostHog, Stripe, and Vercel are listed with "not connected" status and explicit "coming soon" actions.
- Both mirrored frontends (`apps/cloud-console` and `services/omnichannel/frontend`) render the same integrations UI.

## Progress

- [x] (2026-03-08 23:40Z) Confirmed team/project prerequisites and created Linear ticket `ANM-58`; moved ticket to `In Progress`.
- [x] (2026-03-08 23:41Z) Audited current settings integrations UI and confirmed both frontend files are currently identical.
- [x] (2026-03-08 23:42Z) Implemented reusable integrations card rendering and added PostHog/Stripe/Vercel entries in `apps/cloud-console/app/settings/page.tsx`.
- [x] (2026-03-08 23:42Z) Mirrored integrations updates to `services/omnichannel/frontend/app/settings/page.tsx` and re-verified parity.
- [x] (2026-03-08 23:43Z) Ran frontend build validation for both workspaces and captured outputs.
- [x] (2026-03-08 23:43Z) Performed post-change bug/unfinished rescan and ticket reconciliation (no new unique follow-up tickets required).
- [ ] Sync final plan + Linear status and document outcomes.

## Surprises & Discoveries

- Observation: Both settings pages are byte-identical right now, so mirroring can be done by copying a single validated implementation.
  Evidence: `diff -u apps/cloud-console/app/settings/page.tsx services/omnichannel/frontend/app/settings/page.tsx` returned no diff output.

- Observation: Frontend builds still emit stale baseline-browser-mapping warnings, but the issue is already tracked.
  Evidence: Build output warning matches existing backlog tickets `ANM-45` and `ANM-36`; no duplicate ticket created.

## Decision Log

- Decision: Implement provider cards directly in the existing Integrations tab component instead of creating a separate shared package component in this slice.
  Rationale: Keeps this request small and low-risk while still preserving parity by mirroring the same implementation in both app surfaces.
  Date/Author: 2026-03-08 / Codex

## Outcomes & Retrospective

Outcome:

- Integrations tab now shows four explicit provider cards: Google Cloud, PostHog, Stripe, and Vercel.
- Google Cloud connect/disconnect behavior remained intact.
- Generic "more integrations coming soon" placeholder was replaced by concrete provider cards with disabled "coming soon" actions for providers not yet wired.
- Both mirrored settings pages are identical after implementation.

Remaining gaps:

- PostHog/Stripe/Vercel actions are intentionally placeholders until backend auth/provider flows are implemented.
- A browser-baseline dependency warning persists in builds and is tracked in existing Linear tickets (`ANM-45`, `ANM-36`).

Lesson:

- A small local `IntegrationCard` helper reduced duplication while preserving low-risk, file-local changes in mirrored frontends.

## Context And Orientation

Relevant paths:

- `apps/cloud-console/app/settings/page.tsx`: Settings route used in Cloud Console shell.
- `services/omnichannel/frontend/app/settings/page.tsx`: Mirrored Settings route used in Omnichannel frontend shell.
- `plans/settings-integrations-cards-execplan.md`: This execution plan.

Non-obvious terms:

- "Mirrored frontends": two separate Next.js app trees that intentionally keep navigation/settings UX behavior aligned.
- "Integration card": visual block showing provider identity, status, and connection action.

## Plan Of Work

1. Update `IntegrationsTab` to render a list of integration cards with shared structure.
2. Keep existing Google Cloud connection behavior and scope display intact.
3. Add PostHog/Stripe/Vercel entries as non-connected providers with disabled "Coming soon" actions and concise descriptions.
4. Remove the generic dashed placeholder section after provider cards are introduced.
5. Copy the finalized settings file into the mirrored frontend path.
6. Validate with frontend builds and capture evidence.
7. Perform required rescan/reconciliation and align plan/ticket completion state.

## Concrete Steps

Run from repository root:

    npm run build --workspace apps/cloud-console
    npm run build --workspace services/omnichannel/frontend

Expected snippets:

- Build commands complete successfully with no type/lint errors that block build.
- `/settings?tab=integrations` shows Google Cloud, PostHog, Stripe, and Vercel cards.

## Validation And Acceptance

Acceptance criteria:

- Integrations tab includes cards for Google Cloud, PostHog, Stripe, and Vercel.
- Google Cloud connected/disconnected state logic still works.
- PostHog/Stripe/Vercel cards show explicit not-connected status and non-breaking placeholder action.
- Both frontend implementations are aligned.

## Idempotence And Recovery

This work is UI-only and idempotent. Re-applying the same edits should produce the same rendered cards. If visual regressions occur, restore the previous settings pages from git history for the two touched files and re-run the build commands.

## Artifacts And Notes

Artifacts:

- `apps/cloud-console/app/settings/page.tsx`
- `services/omnichannel/frontend/app/settings/page.tsx`
- `plans/settings-integrations-cards-execplan.md`

Validation evidence:

    npm run build --workspace apps/cloud-console
    # next build completed successfully (compiled, type check, static generation)

    npm run build --workspace services/omnichannel/frontend
    # next build completed successfully (compiled, type check, static generation)

    cmp -s apps/cloud-console/app/settings/page.tsx services/omnichannel/frontend/app/settings/page.tsx && echo "settings pages are identical"
    # settings pages are identical

## Interfaces And Dependencies

No new API contracts or external dependencies are required. Existing React state in `SettingsPage` (`gcloudStatus`, `gcloudEmail`, `gcloudProject`) remains authoritative for Google Cloud integration status.

Revision Note (2026-03-08): Initial plan created and linked to ANM-58 for integrations-card expansion in settings UI.
Revision Note (2026-03-08): Recorded completed implementation, validation evidence, and ticket reconciliation results.
