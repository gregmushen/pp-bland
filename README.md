# Bland CLI

**Send a task-driven call and follow it through to the result.**

Create one outbound call from a natural-language task, wait for its status, and inspect the transcript and outcome. The CLI keeps locally searchable records for calls it makes.

## Install

Requires Go 1.26.6 or newer. Once this pull request is merged, install the CLI directly from this repository:

```bash
go install github.com/gregmushen/pp-bland/cmd/bland-pp-cli@latest
```

For a local checkout, run `make build` and use `./bin/bland-pp-cli`. Set `BLAND_API_KEY` in your environment, then check connectivity with `bland-pp-cli doctor --json`. The CLI never needs the key committed to this repository.

To run the MCP server, install it with:

```bash
go install github.com/gregmushen/pp-bland/cmd/bland-pp-mcp@latest
```

Configure your MCP host to run `bland-pp-mcp` with `BLAND_API_KEY` in its environment. The agent instructions in [`SKILL.md`](SKILL.md) describe the task call and local recall commands.

## Quick Start

```bash
# Check the CLI setup without contacting Bland.
bland-pp-cli doctor --dry-run

# Preview the task workflow without contacting anyone.
bland-pp-cli calls task run --phone-number +12065550123 --task "Ask whether a table is available tonight" --dry-run

```

## Unique Features

These commands combine task completion with local recall for calls made through this CLI.

### Task to outcome
- **`calls task run`** — Start one task call, wait for its final status, and return the transcript and outcome together.

  _Use this after the user authorizes a call when you need to complete a phone task and report what happened._

  ```bash
  bland-pp-cli calls task run --phone-number +12065550123 --task "Ask whether a table is available tonight" --agent
  ```
- **`calls search-task`** — Find prior calls made through this CLI by task wording and inspect their returned outcomes.

  _Use this to find a previous CLI-originated call by what it was asked to do._

  ```bash
  bland-pp-cli calls search-task reservation --agent
  ```

## Recipes

### Preview a reservation call

```bash
bland-pp-cli calls task run --phone-number +12065550123 --task "Ask whether a table for two is available tonight" --dry-run
```

Check the exact request without placing a call.

### Send and follow one task

```bash
bland-pp-cli calls task run --phone-number +12065550123 --task "Ask whether a table for two is available tonight" --agent
```

Places a real call; invoke only after the user approves the destination and task.

### Review recent calls

```bash
bland-pp-cli calls list --limit 5 --agent
```

List recent account calls as structured data.

### Inspect a call

```bash
bland-pp-cli calls get d1f7e65b-ec29-48a2-8c38-0e93d9438ae1 --agent --select status,transcripts,summary
```

Narrow a call record to status and the human-readable result.

### Inspect call events

```bash
bland-pp-cli event-stream d1f7e65b-ec29-48a2-8c38-0e93d9438ae1 --agent
```

Read the event stream for an individual call.

### Search local task history

```bash
bland-pp-cli calls search-task reservation --agent --select call_id,status,task,summary
```

Find a prior locally cached call by its task wording.

## Usage

Run `bland-pp-cli --help` for the full command reference and flag list.

## Paths & environment variables

This CLI separates local files into four path kinds:

| Kind | Contents |
|------|----------|
| `config` | User-editable settings such as `config.toml` and saved profiles |
| `data` | Durable local data: `credentials.toml`, `data.db`, cookies, browser-session proof files, and other auth sidecars |
| `state` | Runtime state such as persisted queries, jobs, and `teach.log` |
| `cache` | Regenerable HTTP/cache files |

Each kind resolves independently. The ladder is:

1. Per-kind env var: `BLAND_CONFIG_DIR`, `BLAND_DATA_DIR`, `BLAND_STATE_DIR`, or `BLAND_CACHE_DIR`
2. `--home <dir>` for this invocation
3. `BLAND_HOME` for a flat relocated root
4. XDG env vars: `XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `XDG_STATE_HOME`, `XDG_CACHE_HOME`
5. Platform defaults matching existing installs

For containers and agent sandboxes, prefer a single relocated root:

```bash
export BLAND_HOME=/srv/bland
bland-pp-cli doctor
```

Under `BLAND_HOME=/srv/bland`, the four dirs resolve to `/srv/bland/config`, `/srv/bland/data`, `/srv/bland/state`, and `/srv/bland/cache`.

MCP servers do not receive CLI flags from the host. Put relocation in the host `env` block:

```json
{
  "mcpServers": {
    "bland": {
      "command": "bland-pp-mcp",
      "env": {
        "BLAND_HOME": "/srv/bland"
      }
    }
  }
}
```

Precedence matters in fleets: an ambient per-kind variable such as `BLAND_DATA_DIR` overrides an explicit `--home` for that kind. Use `BLAND_HOME` or the per-kind variables for durable fleet relocation; treat `--home` as the weaker per-invocation lever.

Relocation is one-way. Unsetting `BLAND_HOME` does not move files back to platform defaults, and `doctor` cannot find credentials left under a former root. Move the files manually before unsetting relocation variables.

Existing installs keep working because the platform-default rung matches the legacy layout. On the first auth write, stored secrets leave `config.toml` and are consolidated into `credentials.toml` under the data directory. Run `bland-pp-cli doctor --fail-on warn` to check path and credential-location warnings in automation.

## Commands

### active

Manage active

- **`bland-pp-cli active`** - List calls not marked complete

### calls

Manage calls

- **`bland-pp-cli calls create`** - Places a real outbound call. May incur charges.
- **`bland-pp-cli calls get`** - Get call status and outcome
- **`bland-pp-cli calls list`** - List call history

### event-stream

Manage event stream

- **`bland-pp-cli event-stream <call_id>`** - Get a call event stream


### Self-learning loop

This CLI caches per-question discovery so repeat queries skip the walk and structurally similar queries get answered via entity substitution. The loop also self-captures: every invocation is journaled locally, and failed-flag corrections plus fresh teaches surface as candidates on the next `recall` for confirm/reject judgment. Agents call `recall` before discovery and fire `teach &` after answering. See the `## Automatic learning` section in `SKILL.md` for the full protocol.

- **`bland-pp-cli recall <query>`** - Look up cached resources for a query before running discovery
- **`bland-pp-cli teach`** - Record a query -> resource mapping (silent on success, safe to background with `&`)
- **`bland-pp-cli learnings list`** - Inspect taught rows
- **`bland-pp-cli learnings forget <query>`** - Undo a teach
- **`bland-pp-cli learnings candidates`** - List auto-captured candidates awaiting confirm/reject
- **`bland-pp-cli learnings stats`** - Local loop metrics: recall hit rate, teach-to-reuse, playbook resolution, candidate counts
- **`bland-pp-cli teach-pattern`** - Install a query/resource template up front
- **`bland-pp-cli teach-lookup`** - Add an entity mapping (e.g. country code, team alias) for pattern substitution

Pass `--no-learn` or set `BLAND_NO_LEARN=true` to disable the loop for deterministic flows.

The local store's schema version stamp is one-way: once this version of `bland-pp-cli` opens the database, older binaries refuse it with a version error — upgrade the binary rather than downgrading.

## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
bland-pp-cli calls list

# JSON for scripting and agents
bland-pp-cli calls list --json
# Filter to specific fields by name
bland-pp-cli calls list --json --select <field>[,<field>...]

# Dry run — show the request without sending
bland-pp-cli calls list --dry-run

# Agent mode — JSON + compact + no prompts in one flag
bland-pp-cli calls list --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select <field>[,<field>...]` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Explicit retries** - add `--idempotent` to create retries when a no-op success is acceptable
- **Explicit confirmation** - `--agent` does not imply `--yes`; pass `--yes` separately only after the target, arguments, and side effects are clear
- **Piped input** - write commands can accept structured input when their help lists `--stdin`
- **Offline-friendly** - sync/search commands can use the local SQLite store when available
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `6` partial failure, `7` rate limited, `10` config error.

## Health Check

```bash
bland-pp-cli doctor
```

Verifies configuration, credentials, and connectivity to the API.

## Configuration

Run `bland-pp-cli doctor` to see the resolved config, data, state, and cache directories. The platform-default config path is `~/.config/bland-task-calls-pp-cli/config.toml`; `--home`, `BLAND_HOME`, and per-kind env vars can relocate it.

Static request headers can be configured under `headers`; per-command header overrides take precedence.

Environment variables:

| Name | Kind | Required | Description |
| --- | --- | --- | --- |
| `BLAND_API_KEY` | per_call | Yes | Set to your API credential. |

### agentcookie (optional)

If you use agentcookie to sync secrets across machines, this CLI auto-adopts agentcookie-managed credentials with no extra setup. When the daemon writes to this CLI's config, `bland-pp-cli doctor` reports `agentcookie: detected` and `auth-status` labels the source as `agentcookie`. Skip this section if you don't use agentcookie - the CLI works the same as any other.

## Troubleshooting
**Authentication errors (exit code 4)**
- Run `bland-pp-cli doctor` to check credentials
- Check whether the environment variable is set without printing it: `test -n "${BLAND_API_KEY:-}" && echo set || echo missing`
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

### API-specific
- **The API returns 401 Unauthorized.** — Set BLAND_API_KEY in the environment and run bland-pp-cli doctor.
- **A call is still queued or in progress.** — Run bland-pp-cli calls get <call-id> again or inspect bland-pp-cli event-stream <call-id>.
