# SKILLS.md

This file describes the MCP tools available in this server and how Claude agents should use them effectively.

## What is Collibra?

Collibra is a data governance platform — a central catalog where an organization documents, classifies, and governs its data assets. It is the authoritative source for:

- **What data exists**: tables, columns, datasets, reports, APIs, and other data assets across the organization
- **What data means**: a rich business glossary of terms, acronyms, KPIs, and definitions that captures how the business interprets and communicates about data — the authoritative place to resolve ambiguity around business language
- **How data relates**: lineage between physical columns, semantic data attributes, business terms, and measures
- **Who owns and trusts it**: stewards, data contracts, classifications, and quality rules

Reach for Collibra tools when the user's question is about **understanding, discovering, or governing data in the organization** — e.g. "what customer data do we have?", "what does this metric measure?", "which columns contain PII?", or "where does this KPI come from?". These tools are not appropriate for querying the actual data values in a database; they operate on the metadata and governance layer above the data.

## Tool Inventory

### Asset Creation

**`prepare_create_asset`** — Resolve asset type and domain by name or ID, hydrate the full attribute schema, and check for duplicates. Returns a structured status (`ready`, `incomplete`, `needs_clarification`, `duplicate_found`) with pre-fetched options for missing fields. **Always call this before `create_asset`** to obtain the resolved UUIDs and validate inputs. Read-only.

**`create_asset`** — Create a new data asset in Collibra with optional attributes. Requires the resolved asset type UUID, domain UUID, and asset name — use the values returned by `prepare_create_asset`. Destructive (creates a new asset).

**`prepare_add_business_term`** — Validate business term data, resolve domains by name, check for duplicates, and hydrate the attribute schema for the Business Term type. Returns structured status with pre-fetched options for missing fields. **Always call this before `add_business_term`**. Read-only.

**`add_business_term`** — Create a business term asset with an optional definition and additional attributes. Requires the domain UUID — use the resolved domain from `prepare_add_business_term`. Destructive (creates a new asset).

### Discovery & Search

**`discover_data_assets`** — Natural language semantic search over data assets (tables, columns, datasets). Use when the user asks open-ended questions like "what data do we have about customers?". Requires `dgc.ai-copilot` permission.

**`discover_business_glossary`** — Natural language semantic search over the business glossary (terms, acronyms, KPIs, definitions). Use when the user asks about the meaning of a business concept. Requires `dgc.ai-copilot` permission.

**`search_asset_keyword`** — Wildcard keyword search. Returns names, IDs, and metadata but not full asset details. Use this to find an asset's UUID when you only know its name. Supports filtering by resource type, community, domain, asset type, status, and creator. Paginated via `limit`/`offset`.

**`list_asset_types`** — List all asset type names and UUIDs. Use this when you need a type UUID to filter `search_asset_keyword` results.

### Asset Details

**`get_asset_details`** — Retrieve full details for a single asset by UUID: attributes, relations, and metadata. Returns a direct link to the asset in the Collibra UI. Relations are paginated (50 per page); use `outgoingRelationsCursor` and `incomingRelationsCursor` from the previous response to page through them.

### Semantic Graph Traversal

These tools walk the Collibra asset relation graph to answer lineage and semantic questions. All require asset UUIDs as input.

**`get_column_semantics`** — Given a column UUID, returns all connected Data Attributes with their descriptions, linked Measures, and generic business assets. Use to answer "what does this column mean semantically?".

**`get_table_semantics`** — Given a table UUID, returns all columns with their Data Attributes and connected Measures. Use to answer "what metrics use data from this table?" or "what is the semantic context of this table?".

**`get_measure_data`** — Given a measure UUID, traces backward through Data Attributes to the underlying Columns and their parent Tables. Use to answer "what physical data feeds this metric?".

**`get_business_term_data`** — Given a business term UUID, traces through Data Attributes to connected Columns and Tables. Use to answer "what physical data is associated with this business term?".

### Data Classification

**`search_data_class`** — Search for data classes by name or description. Use this to find a classification UUID before applying it to an asset. Requires `dgc.data-classes-read` permission.

**`search_data_classification_match`** — Search existing classification matches (associations between data classes and assets). Filter by asset IDs, classification IDs, or status (`ACCEPTED`, `REJECTED`, `SUGGESTED`). Requires `dgc.classify` + `dgc.catalog`.

**`add_data_classification_match`** — Apply a data class to an asset. Requires both the asset UUID and classification UUID. Requires `dgc.classify` + `dgc.catalog`.

**`remove_data_classification_match`** — Remove a classification match. Requires `dgc.classify` + `dgc.catalog`.

### Technical Lineage

These tools query the technical lineage graph — a map of all data objects and transformations across external systems, including unregistered assets, temporary tables, and source code. Unlike business lineage (which only covers assets in the Collibra Data Catalog), technical lineage covers the full physical data flow.

**Workflow**: Almost all lineage questions follow the same pattern: **(1)** `search_lineage_entities` → **(2)** `get_lineage_upstream` or `get_lineage_downstream` → **(3)** optionally `get_lineage_entity` for the most relevant entities only. Do not resolve every entity ID — summarize from the graph structure and only look up entities the user specifically needs details on. Only call `get_lineage_transformation` when the user asks to see actual SQL or logic.

**IMPORTANT — ID types**: Lineage tools use their own internal entity IDs, which are **not** the same as DGC asset UUIDs. You cannot pass a DGC asset UUID directly to `get_lineage_upstream` or `get_lineage_downstream`. To bridge from the catalog to the lineage graph, call `search_lineage_entities` with the asset's UUID as `dgcId` to obtain the lineage entity ID first.

**LIMITATION — Column-level lineage**: Columns cannot be searched by name in `search_lineage_entities` (`nameContains` does not work for columns). The `dgcId` parameter also does not reliably resolve columns because there is no consistent mapping between Collibra catalog column UUIDs and technical lineage entity IDs. To reach a column in the lineage graph, first find its parent table (by name or `dgcId`), then use `get_lineage_upstream` or `get_lineage_downstream` on the table to discover its columns in the lineage graph.

**`search_lineage_entities`** *(entry point)* — Search by name, type, or DGC UUID. **Start here** for almost all lineage questions to resolve an entity name or DGC asset UUID to a lineage entity ID. Supports partial name matching and type filtering (e.g. `table`, `column`, `report`). Paginated. **Note**: name search and DGC UUID lookup do not work reliably for columns — see limitation above.

**`get_lineage_upstream`** *(step 2: trace sources)* — Given a lineage entity ID (not a DGC UUID), returns all upstream source entities and connecting transformations. Use to answer "where does this data come from?". Results contain entity IDs only. Paginated.

**`get_lineage_downstream`** *(step 2: trace consumers)* — Given a lineage entity ID (not a DGC UUID), returns all downstream consumer entities and connecting transformations. Use for impact analysis: "what depends on this?", "what breaks if this changes?". Results contain entity IDs only. Paginated.

**`get_lineage_entity`** *(follow-up: resolve IDs)* — Get full metadata for a specific lineage entity by its lineage ID (not a DGC UUID): name, type, source systems, parent entity, and linked DGC identifier. Only call this for the most relevant entity IDs from upstream/downstream results — do not resolve every ID.

**`get_lineage_transformation`** *(terminal: view logic)* — Get the full details of a transformation, including its SQL or script logic. Only call when the user explicitly asks about the transformation code. Do not call just to understand the lineage graph.

**`search_lineage_transformations`** *(specialized)* — Search for transformations by name. Only use when the user explicitly asks about a transformation by name. This is **not** a general entry point for lineage questions — start with `search_lineage_entities` instead.

### Data Contracts

**`list_data_contract`** — List data contracts with cursor-based pagination. Filter by `manifestId`. Use this to find a contract's UUID.

**`pull_data_contract_manifest`** — Download the manifest for a data contract by UUID.

**`push_data_contract_manifest`** — Upload/update a manifest for a data contract by UUID.

---

## Integration Lifecycle Management

These tools manage GENERIC integration instances — Databricks Unity Catalog and Google Dataplex Universal Catalog syncs running on Collibra Edge. They cover three API surfaces: Edge Management (`/edge/api/rest/v2`), Catalog Cloud Ingestions (`/rest/catalog/1.0/genericIntegration`), and Jobs (`/rest/jobs/v1`).

> **Key mapping**: A GENERIC integration instance is an Edge *capability*. The `capability.Id` from the Edge API is the same UUID as the `ingestibleId` used in the Catalog and Jobs APIs. Always use this UUID to move between the three APIs.

### Edge: Capabilities (integrations)

**`edge_list_capabilities`** — List all Edge capabilities (integration instances). Use this as the entry point to discover `ingestibleId` values. Returns `id`, `name`, `type.Id`, `edgeSiteId`, and `parameters` for each capability. `type.Id` is a human-readable string (e.g. `"databricks-edge-capability"`) — use it to filter by integration platform directly from the response.

**`edge_find_capabilities`** — Filter capabilities by `edgeSiteId`, `labels`, or `parameters`. All filters are ANDed; multiple values within `labels` or `parameters` are ORed. To filter by integration type, pass `labels: {"capability-type": "<type-id>"}` — the label value is the same string as `type.Id` on the capability (e.g. `"databricks-edge-capability"`). If you don't know the exact type ID, first call `edge_find_capabilities` with only `edgeSiteId` to get a sample, read `type.Id` from any result, then use that as the label value in a second call. Name filtering is not supported server-side — filter by name client-side after receiving results.

**`edge_get_capability`** — Get full details for a single capability by UUID.

**`edge_create_capability`** — Create a new capability. Requires `name`, `typeId`, and `edgeSiteId`. Pass integration-specific parameters in the `parameters` map.

**`edge_update_capability`** — Update an existing capability's name, type, site, or parameters.

**`edge_delete_capability`** — Delete a capability permanently. Confirm with the user before calling.

**`edge_run_capability`** — Trigger a capability run directly via the Edge API. Returns the Edge job UUID. For GENERIC integrations, prefer `catalog_generic_start_job` which returns richer job details.

### Edge: Connections

**`edge_list_connections`** — List all Edge connections. Use to discover connection UUIDs before creating or updating capabilities that reference them.

**`edge_find_connections`** — Filter connections by `edgeSiteId`, `name`, or `nameMatchMode` (`EXACT` or `ANYWHERE`).

**`edge_get_connection`** — Get full details for a single connection by UUID.

**`edge_create_connection`** — Create a new connection. Requires `name`, `typeId`, and `edgeSiteId`. Pass credentials and endpoint parameters in the `parameters` map.

**`edge_update_connection`** — Update an existing connection.

**`edge_delete_connection`** — Delete a connection permanently. Confirm with the user before calling.

**`edge_test_connection`** — Test a connection's reachability. Pass `timeoutSec` to wait for the result synchronously; omit to run asynchronously. Returns `success`, `message`, and a `jobId` for async tracking.

### Edge: Jobs

**`edge_cancel_job`** — Cancel a running Edge job by its Edge job UUID (not the catalog job ID).

**`edge_get_job_status`** — Get the current status log entry for an Edge job. Returns `status`, `message`, and `lastUpdatedDateTime`.

**`edge_get_job_status_history`** — Get the full status history for an Edge job. Use to trace how a job progressed through states.

### Catalog: GENERIC Integration Config

**`catalog_generic_get_config`** — Get the current configuration for a GENERIC integration by `ingestibleId`.

**`catalog_generic_save_config`** — Create or update the configuration. Pass the config as a JSON string in the `configuration` field.

**`catalog_generic_delete_config`** — Delete the configuration. Confirm with the user before calling.

**`catalog_generic_get_schema`** — Get the JSON schema that describes valid configuration values for this integration type. Use this before calling `catalog_generic_save_config` to understand what fields are required.

### Catalog: GENERIC Integration Schedules

**`catalog_generic_get_schedule`** — Get the active schedule for an integration. Returns `cronExpression`, `cronTimeZone`, `lastRunTimeStamp`, and `nextRunDateLongValue` (both as Unix epoch milliseconds).

**`catalog_generic_get_all_schedules`** — Get all schedules (including per-workflow schedules) for an integration.

**`catalog_generic_add_schedule`** — Create a new schedule. Requires `cronExpression` and `cronTimeZone`.

**`catalog_generic_update_schedule`** — Update the existing schedule. Same inputs as add.

**`catalog_generic_delete_schedule`** — Delete the schedule. Pass `workflow` to target a specific workflow schedule.

### Catalog: GENERIC Integration Jobs

**`catalog_generic_start_job`** — Trigger an immediate sync run. Returns 202 Accepted with a `LegacyJobDto` — the job is queued, not yet complete. Always call `catalog_generic_get_schedule` first to confirm no job is currently running.

**`catalog_generic_cancel_job`** — Cancel the currently running sync job. Returns 404 if no job is running — treat this as a success (already stopped). Confirm with the user before calling.

### Jobs API

**`jobs_find`** — Search jobs by name, state (`WAITING`, `RUNNING`, `COMPLETED`, `FAILED`, `DELETED`), result (`SUCCESS`, `COMPLETED_WITH_ERROR`, `FAILURE`, `ABORTED`), type, or user. Paginated via `cursor` and `pageSize`. Use to find the last job for an integration by searching for its name.

**`jobs_get`** — Get full details for a specific job by UUID. Use this after `jobs_find` to get the `message` field for error details.

---

## Integration Lifecycle Workflows

### List integrations
1. `list_integrations` — returns all capabilities with schedule + last run enriched. Returns all types unless you pass `platform` (e.g. `"databricks"`, `"dataplex"`). Client instances have ≤20 capabilities so a full fetch is safe.
2. Each result includes `ingestibleId`, `typeId`, `hasSchedule`, `lastRunAt` (ISO 8601), `lastRunState`, `lastRunResult`, and `nextRun`.
3. Known type IDs: `"databricks-edge-capability"` (Databricks Unity Catalog), `"dataplex-synchronization"` (Dataplex metadata sync), `"dataplex-lineage-synchronization"` (Dataplex lineage sync).

### Find integrations that ran in the last 24h
Do **not** fetch all integrations and filter by `lastRunAt` client-side. Instead use the Jobs API as the primary source:
1. `jobs_find` with `sortField: "START_DATE"`, `sortOrder: "DESC"`, and a reasonable page size (e.g. 50).
2. Filter the results to jobs whose `startDate` is within the last 24h.
3. The `name` field of each job is `"Synchronization for <capability name>"` — strip the prefix to get the capability name.
4. Optionally cross-reference with `list_integrations` if you need `ingestibleId` or schedule details for those capabilities.

### Find integrations on a specific Edge site
1. `edge_find_capabilities` with `edgeSiteId` → filtered list
2. Use `type.Id` from results to distinguish integration types

### Check integration status (schedule + last run)
1. `catalog_generic_get_schedule` with `ingestibleId` → current schedule, `lastRunTimeStamp`, `nextRunDateLongValue`
2. `jobs_find` with the integration name → find the most recent job
3. `jobs_get` with the job UUID → full state, result, progress, and message

### Trigger a sync run
1. `catalog_generic_get_schedule` → confirm state is not actively running
2. `jobs_find` with state `RUNNING` and the integration name → double-check no job is in flight
3. Confirm with the user: _"I'll trigger a run for `<name>`. Proceed?"_
4. `catalog_generic_start_job` → returns job details in `WAITING` or `RUNNING` state
5. Poll `jobs_find` or `jobs_get` to report progress

### Cancel a running sync
1. `jobs_find` with state `RUNNING` and the integration name → confirm a job is actually running
2. Confirm with the user before cancelling
3. `catalog_generic_cancel_job` → 204 on success, 404 means already stopped (treat as success)

### Manage a schedule
1. `catalog_generic_get_schedule` → see current schedule
2. `catalog_generic_update_schedule` to change cron expression/timezone, or `catalog_generic_delete_schedule` to remove it
3. `catalog_generic_add_schedule` if no schedule exists yet

### Diagnose a failed sync
1. `jobs_find` with result `FAILURE` or `COMPLETED_WITH_ERROR` and the integration name
2. `jobs_get` with the job UUID → read the `message` field for error detail
3. `catalog_generic_get_config` → verify the configuration is still valid
4. `catalog_generic_get_schema` → check config against the schema if the error suggests misconfiguration

### Set up a new integration
1. `edge_list_connections` → find or confirm the connection UUID to use
2. `edge_test_connection` with the connection UUID → confirm it is reachable
3. `edge_create_capability` with `typeId`, `edgeSiteId`, and `parameters` (including the connection UUID)
4. `catalog_generic_get_schema` → understand required configuration fields
5. `catalog_generic_save_config` → apply the configuration
6. `catalog_generic_add_schedule` → set a cron schedule if needed
7. `catalog_generic_start_job` → trigger the first run

### Test and update a connection
1. `edge_test_connection` → check current reachability
2. `edge_update_connection` → change parameters (endpoint, credentials)
3. `edge_test_connection` again → confirm the updated connection works

---

## Common Workflows

### Create any asset
1. `prepare_create_asset` with the asset name, asset type (publicId), and domain ID → check status is `ready`
2. `create_asset` with the resolved `assetTypeId` and `domainId` from step 1

### Add a business term
1. `prepare_add_business_term` with the term name and domain name or ID → check status is `ready`
2. `add_business_term` with the resolved `domainId` from step 1, plus optional definition and attributes

### Find an asset and get its details
1. `search_asset_keyword` with the asset name → get UUID from results
2. `get_asset_details` with the UUID → get full attributes and relations

### Classify a column
1. `search_asset_keyword` to find the column UUID
2. `search_data_class` to find the data class UUID
3. `add_data_classification_match` with both UUIDs

### Understand what a table means
1. `search_asset_keyword` to find the table UUID
2. `get_table_semantics` → columns → data attributes → measures

### Trace a metric to its source data
1. `search_asset_keyword` to find the measure UUID
2. `get_measure_data` → data attributes → columns → tables

### Trace a business term to physical data
1. `search_asset_keyword` to find the business term UUID
2. `get_business_term_data` → data attributes → columns → tables

### Trace upstream lineage for a data asset
1. `search_lineage_entities` with the asset name → get entity ID
2. `get_lineage_upstream` → relations with source entity IDs and transformation IDs
3. Summarize based on the graph structure — only call `get_lineage_entity` for the most relevant source entities, not all of them
4. Only call `get_lineage_transformation` if the user explicitly asks to see the SQL or logic

### Perform impact analysis (downstream)
1. `search_lineage_entities` with the asset name → get entity ID
2. `get_lineage_downstream` → relations with consumer entity IDs
3. Summarize based on the graph structure — only call `get_lineage_entity` for the most relevant consumers, not all of them

### Manage a data contract
1. `list_data_contract` to find the contract UUID
2. `pull_data_contract_manifest` to download, edit, then `push_data_contract_manifest` to update

---

## Tips

- **Always prepare before creating.** Call `prepare_create_asset` before `create_asset` and `prepare_add_business_term` before `add_business_term`, even if you already have the UUIDs. The prepare tools validate inputs, check for duplicates, and return the attribute schema.
- **UUIDs are required for most tools.** When you only have a name, start with `search_asset_keyword` or the natural language discovery tools to get the UUID first.
- **`discover_data_assets` vs `search_asset_keyword`**: Prefer `discover_data_assets` for open-ended semantic questions; prefer `search_asset_keyword` when you know the exact name or need to filter by type/community/domain.
- **Permissions**: `discover_data_assets` and `discover_business_glossary` require the `dgc.ai-copilot` permission. Classification tools require `dgc.classify` + `dgc.catalog`. If a tool fails with a permission error, let the user know which permission is needed.
- **Pagination**: `search_asset_keyword`, `list_asset_types`, `search_data_class`, and `search_data_classification_match` use `limit`/`offset`. `list_data_contract` and `get_asset_details` (for relations) use cursor-based pagination — carry the cursor from the previous response. Lineage tools (`search_lineage_entities`, `get_lineage_upstream`, `get_lineage_downstream`, `search_lineage_transformations`) also use cursor-based pagination.
- **Error handling**: Validation errors are returned in the output `error` field (not as Go errors), so always check `error` and `success`/`found` fields in the response before using the data.
- **Integration IDs**: The `capability.Id` from the Edge API is the same UUID as `ingestibleId` in the Catalog and Jobs APIs. One UUID, three API surfaces.
- **Job monitoring**: Use the Jobs API (`jobs_find`, `jobs_get`) for catalog-level job state and error messages. Use `edge_get_job_status`/`edge_get_job_status_history` for low-level Edge execution status.
- **Timestamps**: `lastRunTimeStamp` and `nextRunDateLongValue` in schedule responses are Unix epoch milliseconds — convert to human-readable time before presenting to the user.
- **Async jobs**: `catalog_generic_start_job` returns 202 Accepted. The job is queued, not complete. Poll `jobs_get` for progress rather than assuming success immediately.
- **Cancel 404 is not an error**: `catalog_generic_cancel_job` returns 404 when no job is running. This is expected — treat it as "already stopped".
- **Confirm before destructive actions**: Always ask the user before calling `catalog_generic_start_job`, `catalog_generic_cancel_job`, `edge_delete_capability`, `edge_delete_connection`, or any delete/cancel operation.
- **Capability type filtering**: Use `edge_find_capabilities` with `labels: {"capability-type": "<type-id>"}` — the label value is the same as `type.Id` in the response (e.g. `"databricks-edge-capability"`). This avoids fetching the full list. If the exact type ID is unknown, do a discovery call with `edgeSiteId` first to read `type.Id` from a sample result. Filter by name client-side after.
