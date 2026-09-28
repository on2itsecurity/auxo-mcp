---
name: auxo-case-management
description: Use when the user wants to look up, open, update, escalate, de-escalate or close an AUXO support case or ticket with ON2IT, for example "show my open cases", "raise a security incident", "change the priority of case X", "add a note to my ticket" or "close this case". Covers the AUXO case tools, confirming before escalation or closure, and summarising cases instead of dumping raw output.
---

# Manage AUXO cases

These tools work on support cases (tickets) with the ON2IT Security Operations Center. They need the Tickets API token. If a tool reports a missing `AUXO_TICKET_TOKEN`, tell the user to enter the Tickets API token in the plugin's settings.

## Rules

- Confirm before any change. Before `createCase`, `updateCasePriority`, `updateCasePrimaryContact`, `updateCaseSubject` or `addNoteToCase`, show the user what you will send and wait for a yes.
- `escalateCase`, `deescalateCase` and `closeCase` change how ON2IT handles the case. Always state the case, its current priority and state, and the effect, then wait for an explicit yes before calling them. Never escalate or close on your own initiative.
- Summarise. Present cases as short tables or bullet lists with the fields the user needs. Do not paste raw JSON from the API.
- Never invent a case ID, priority or contact. Look them up or ask.

## Look up cases

A case has two identifiers: `id`, the unique ID that every case tool takes, and `case_number`, the five-character ON2IT number people quote. When the user gives a case number, find the matching `id` with `getCases` first.

- `getCases` returns all cases. Summarise them, for example as a table of case number, subject, `case_type`, `priority`, `state` and whether it is `escalated`, sorted by priority (1 is highest). Treat every `state` other than `closed` as open. Offer to filter by what the user asked for, such as open cases or security incidents.
- `getCase` with `id` returns one case in full. Summarise the subject, type, priority, state, primary contact and the latest entries in `history_of_notes`.

## Create a case

`createCase` requires:

- `subject`: a short title.
- `note`: the initial description. Include what happened, when, which systems are affected and what the user needs from ON2IT.
- `priority`: 1 to 4, where 1 is the highest. Propose one based on impact and urgency, and let the user decide.
- `case_type`: one of
  - `securityincident`: a security incident needing immediate attention
  - `incident`: a service disruption or issue
  - `change`: a planned change needing approval
  - `standardchange`: a pre-approved routine change
  - `inforequest`: a request for information or assistance
- `primary_contact_email`: must match a user known to AUXO. Ask the user for it if you do not know it.

Optionally pass `id`, a friendly customer reference for tracking. It is not the case number, which AUXO generates. After creating, report the case number and a one-line summary.

For a suspected security incident, ask the minimum needed to open the case quickly (what, when, which systems) and suggest priority 1 or 2. Do not delay creating it to collect every detail; further information can be added with notes.

## Update a case

All updates take the case `id`:

- `updateCasePriority` with `new_priority` (1 to 4).
- `updateCasePrimaryContact` with `new_primary_contact_email`.
- `updateCaseSubject` with `new_subject`.
- `addNoteToCase` with `additional_note`. Write notes that stand on their own: facts, times and what is being asked.

Show the old and new value when confirming an update.

## Escalate or de-escalate

- `escalateCase` with `id` asks ON2IT for higher priority attention. Use it when the user reports that impact has grown or a deadline is at risk. Suggest adding a note that explains why, before or after escalating.
- `deescalateCase` with `id` returns the case to normal handling.

Confirm first, as described in the rules.

## Close a case

`closeCase` with `id` requests closure. An ON2IT engineer reviews the request and closes the case if no further work is needed, so tell the user the case may show `request_close` until then. Confirm first, and suggest a closing note with `addNoteToCase` that records the outcome.

## Report back

After each action, reply with the case number, what changed and the current state, in one or two sentences.
