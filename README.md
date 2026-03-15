# buffy-cli

**buffy** is the official CLI for [Buffy](https://buffyai.org) — a personal behavior agent for habits, tasks, and routines. Use it from your terminal or scripts to talk to Buffy, manage settings, and API keys.

**API:** [api.buffyai.org](https://api.buffyai.org)

---

## Install

### Download a binary (recommended)

1. Go to [Releases](https://github.com/phantue2002/buffy-cli/releases) and download the build for your OS/arch, for example:
   - **Linux (x64):** `buffy-1.0.0-linux-amd64`
   - **macOS (Apple Silicon):** `buffy-1.0.0-darwin-arm64`
   - **Windows (x64):** `buffy-1.0.0-windows-amd64.exe`
2. Rename the file to `buffy` (or `buffy.exe` on Windows).
3. Put it in a directory that is in your PATH (e.g. `/usr/local/bin`, `~/bin`, or `C:\Windows`).

### Build from source

Requires [Go 1.21+](https://go.dev/dl/).

```bash
git clone https://github.com/phantue2002/buffy-cli.git
cd buffy-cli
go build -o buffy .
# On Windows: go build -o buffy.exe .
```

Or install into your GOPATH bin:

```bash
go install github.com/phantue2002/buffy-cli@latest
```

---

## Get an API key

You need a Buffy API key (Bearer token) to call the API.

1. Sign up at [buffyai.org](https://buffyai.org).
2. Go to **Account → API keys** and create a key, or use the CLI with an existing key:

```bash
export BUFFY_API_KEY=your_existing_key
buffy api-key create --label my-cli
```

Then set `BUFFY_API_KEY` to the new key, or pass `--api-key KEY` on each command.

---

## Commands

| Command | Description |
|--------|-------------|
| `buffy version` | Show version |
| `buffy message --text "..."` | Send a message (user = API key owner; optional `--user-id` for system keys) |
| `buffy user-settings get --user-id ID` | Get user settings (JSON) |
| `buffy user-settings set --user-id ID [--name ...] [--timezone ...]` | Update settings |
| `buffy api-key list --user-id ID` | List API keys (ID, label, type) |
| `buffy api-key create [--label ...] [--type user\|system]` | Create API key (user = key owner; `--user-id` only for system keys) |
| `buffy api-key revoke --id KEY_ID` | Revoke key by ID (from list) |

**Global flags:** `--api-base URL`, `--api-key KEY`, `--as-user USER_ID` (for system keys acting on behalf of a user).

---

## Examples

Send a message (e.g. create a reminder). With a user API key, no need to pass `--user-id`:

```bash
export BUFFY_API_KEY=your_key
buffy message --text "remind me to drink water every day"
```

List your API keys:

```bash
buffy api-key list --user-id YOUR_USER_ID
```

Use a different API base (e.g. self‑hosted):

```bash
buffy --api-base https://api.example.com message --api-key KEY --text "hello"
```

---

## Guide for agent use

If an **agent** (AI assistant, MCP server, OpenClaw clawbot, cron job, or script) needs to call Buffy on behalf of a user—or multiple users—use a **system API key** and the `--as-user` flag.

### 1. Create a system API key for the agent

1. Sign in at [buffyai.org](https://buffyai.org).
2. Go to **Account → Agent setup**.
3. Create an API key (this is a **system** key). Copy the key and store it securely; the dashboard shows it only once.

Alternatively, with an existing key you can create a system key via the CLI (requires `--user-id` for the account that may create system keys):

```bash
export BUFFY_API_KEY=your_existing_key
buffy api-key create --user-id YOUR_USER_ID --label my-agent --type system
```

### 2. Configure the agent

Set the key in the agent’s environment (or in its config) so the CLI can use it:

```bash
export BUFFY_API_KEY=your_system_key
```

For MCP, OpenClaw, or other frameworks, use the same key in the place where the tool is configured (e.g. Authorization header or env var).

### 3. Act on behalf of a user

With a **system** key, every request must specify **which user** the action is for. Use the global `--as-user` flag:

```bash
# Send a message as user abc123
buffy --as-user abc123 message --text "remind me to drink water every day"

# Get or set that user’s settings
buffy --as-user abc123 user-settings get --user-id abc123
buffy --as-user abc123 user-settings set --user-id abc123 --timezone "Europe/London"

# List or create API keys for that user
buffy --as-user abc123 api-key list --user-id abc123
buffy --as-user abc123 api-key create --user-id abc123 --label new-key
```

The backend uses `X-Buffy-User-ID` (set from `--as-user`) to attribute the request to that user. Without `--as-user`, a system key may be rejected or not attributed to a user.

### 4. Example: script or cron

Your script should know the Buffy **user ID** of the person the agent is helping (e.g. from your auth or session). Then run the CLI with that ID:

```bash
#!/bin/sh
BUFFY_USER_ID="${BUFFY_USER_ID:-}"  # e.g. from your app’s session
if [ -z "$BUFFY_USER_ID" ]; then
  echo "BUFFY_USER_ID not set" >&2
  exit 1
fi
export BUFFY_API_KEY="your_system_key"
buffy --as-user "$BUFFY_USER_ID" message --text "List my habits for today"
```

### 5. User key vs system key

| Key type | Use case | How to act as a user |
|----------|----------|------------------------|
| **User** | Single user (you), scripts that only act as you | No extra flag; the key owner is the user. |
| **System** | Agent, MCP, or script acting for one or many users | Always pass `--as-user USER_ID` so the backend knows which user the action is for. |

Create user keys at **Account → API keys**. Create system keys at **Account → Agent setup** (or via CLI with `--type system` where allowed).

---

## License

Proprietary. Buffy is a private project; all rights reserved.
