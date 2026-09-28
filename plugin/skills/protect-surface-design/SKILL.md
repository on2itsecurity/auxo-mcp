---
name: protect-surface-design
description: Use when the user wants to define, scope, create or refine a Zero Trust protect surface in AUXO, for example "create a protect surface for our CRM", "what should we protect", "assign measures to the payment platform" or "map the flows for this application". Covers identifying the DAAS element, reusing existing protect surfaces and locations, creating the protect surface and its states, assigning security measures and mapping transaction flows with the AUXO tools.
---

# Design a protect surface in AUXO

A protect surface is the smallest set of Data, Applications, Assets and Services (DAAS) that one business function depends on and that needs the same protection. ON2IT's method defines the protect surface first, then maps its transaction flows, then selects and implements measures. Follow the steps below in order.

## Rule: confirm before every change

Before you call any tool that creates, updates or deletes data (`createProtectSurface`, `updateProtectSurface`, `deleteProtectSurface`, `createLocation`, `updateLocation`, `deleteLocation`, `createState`, `updateState`, `deleteState`, `updateProtectSurfaceMeasure`, `removeMeasureFromProtectSurface`, `createTransactionFlow`, `createExternalFlow`, `deleteTransactionFlow`, `deleteExternalFlow`), show the user exactly what you are about to send, in plain language, and wait for an explicit yes. Batch related changes into one confirmation when that is clearer. Never delete anything the user did not name.

## 1. Identify the DAAS element

Ask what business function or crown jewel the user wants to protect and why. Establish:

- The DAAS element: the data, application, asset or service at the centre, for example "customer PII in the CRM".
- Relevance (0 to 100): how important it is to the business.
- Confidentiality, integrity and availability, each normally scored 1 to 5.
- Data tags such as `PII` or `PCI`, and compliance tags such as `GDPR`, `PCI-DSS` or `SOX`.
- Whether it is inside the control boundary (the organisation is responsible for securing it) and whether it is in Zero Trust focus (it should be measured and reported).
- The main contact and security contact.

Keep one business function per protect surface. If the user describes several unrelated systems, propose separate protect surfaces.

## 2. Check what already exists

Before proposing anything new:

1. `listProtectSurfaces` with `name_contains` (and `data_tags` or `compliance_tags` when useful) to find duplicates or a surface to extend. Use `getProtectSurface` for full details of a candidate.
2. `listLocations` to reuse an existing location instead of creating a near-duplicate.
3. `listContacts` to find the contact IDs for `main_contact_person_id` and `security_contact_person_id`, and the e-mail addresses needed later for measures.

If a matching protect surface exists, offer to update it rather than create a second one.

## 3. Create the protect surface

Call `createProtectSurface` with `name`, `description`, `relevance`, `confidentiality`, `integrity`, `availability`, `in_control_boundary`, `in_zero_trust_focus`, `data_tags`, `compliance_tags` and the contact IDs. Send numbers and booleans as JSON numbers and booleans, not strings. Use `customer_labels` for key/value grouping the user asks for.

## 4. Describe where it lives (location and states)

A state records what makes up the protect surface at a location.

1. If no suitable location exists, create one with `createLocation` (`name`, optionally `latitude` and `longitude`).
2. Create one state per kind of content with `createState`: `protectsurface_id`, `location_id`, `content_type` (`ipv4`, `ipv6`, `azure_cloud`, `aws_cloud`, `gcp_cloud`, `container`, `hostname` or `user_identity`), `content` as an array such as `["10.10.0.1", "10.10.0.2"]`, and a short `description`.
3. Use `listStates` to review the result.

## 5. Assign measures

1. Find candidate measures with `listMeasures`, using `search_query` for a topic (for example "segmentation" or "MFA") or `category` for a group. Use `id` for an exact measure such as `ztmm-1.1`.
2. Propose a short list that fits the protect surface's CIA scores and tags, and explain why each one applies.
3. After confirmation, call `updateProtectSurfaceMeasure` per measure with `protect_surface_id`, `measure_name`, `assigned: true` and `assignment_person_email`. Record implementation or evidence only when the user states it, with `implemented: true` plus `implementation_person_email`, or `evidenced: true` plus `evidence_person_email`.
4. Risk acceptance (`risk_no_implementation_accepted` or `risk_no_evidence_accepted`) needs `risk_accepted_comment` and `risk_acceptance_person_email`. Only set it when the user explicitly accepts the risk.
5. Check progress with `listProtectSurfaceMeasures`, using `assigned_only` or `implemented_only` to filter.

## 6. Map transaction flows

Ask which other protect surfaces this one talks to, and whether it receives traffic from or sends traffic to outside the organisation.

- Between protect surfaces: `createTransactionFlow` with `source_protect_surface_id`, `destination_protect_surface_id` and `allow` (true to allow, false to deny). Look up IDs with `listProtectSurfaces`.
- To or from outside: `createExternalFlow` with `protect_surface_id`, `direction` (`inbound` or `outbound`) and `allow`.
- Review with `listTransactionFlows` for the protect surface.

Default to deny for flows the user cannot justify, and say so.

## 7. Summarise

End with a short summary: the protect surface name and ID, its scores and tags, the states per location, the measures assigned and their status, and the flows. Do not paste raw API responses. List any open questions, such as missing contacts or measures still to be implemented.

## Errors

- A missing token error means the Zero Trust API token is not configured. Tell the user to enter it in the plugin's settings.
- Validation errors name the missing field. Ask the user for it rather than guessing.
