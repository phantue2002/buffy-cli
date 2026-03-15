# buffy-cli

**buffy** is the official CLI for [Buffy](https://buffyai.org) — a personal behavior agent for habits, tasks, and routines. Use it from your terminal or scripts to talk to Buffy, manage settings, and API keys.

**API:** [api.buffyai.org](https://api.buffyai.org)

---

## Install

### Download a binary (recommended)

1. Go to [Releases](https://github.com/phantue2002/buffy-agent/releases) (tags `cli/v*`) and download the build for your OS/arch, for example:
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
buffy api-key create --user-id YOUR_USER_ID --label my-cli
```

Then set `BUFFY_API_KEY` to the new key, or pass `--api-key KEY` on each command.

---

## Commands

| Command | Description |
|--------|-------------|
| `buffy version` | Show version |
| `buffy message --user-id ID --text "..."` | Send a message to the agent |
| `buffy user-settings get --user-id ID` | Get user settings (JSON) |
| `buffy user-settings set --user-id ID [--name ...] [--timezone ...]` | Update settings |
| `buffy api-key list --user-id ID` | List API keys (ID, label, type) |
| `buffy api-key create --user-id ID [--label ...] [--type user\|system]` | Create API key |
| `buffy api-key revoke --id KEY_ID` | Revoke key by ID (from list) |

**Global flags:** `--api-base URL`, `--api-key KEY`, `--as-user USER_ID` (for system keys acting on behalf of a user).

---

## Examples

Send a message (e.g. create a reminder):

```bash
export BUFFY_API_KEY=your_key
buffy message --user-id YOUR_USER_ID --text "remind me to drink water every day"
```

List your API keys:

```bash
buffy api-key list --user-id YOUR_USER_ID
```

Use a different API base (e.g. self‑hosted):

```bash
buffy --api-base https://api.example.com message --user-id u1 --api-key KEY --text "hello"
```

---

## License

Same license as the [Buffy](https://github.com/phantue2002/buffy-agent) project.
