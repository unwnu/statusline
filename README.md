# statusline

Git status line tool for Claude Code.

## Example Output

```
statusline on ⎇ main ↑2 ↓1
```

- `⎇` icon color indicates repository status: green (clean), yellow (tracked changes), red (untracked files)
- `↑2` ahead of upstream, `↓1` behind upstream

## Environment Variables

- `STATUSLINE_NO_COLOR=1` — disable colors
- `STATUSLINE_FETCH=1` — fetch upstream (slower, but accurate ↑/↓)

## Claude Code Integration

Add to your `settings.json`:

**macOS/Linux:**
```json
{
  "statusLine": {
    "type": "command",
    "command": "~/.claude/bin/statusline",
    "padding": 0
  },
  "env": {
    "STATUSLINE_FETCH": "1"
  }
}
```

**Windows (PowerShell):**
```json
{
  "statusLine": {
    "type": "command",
    "command": "%USERPROFILE%\\.claude\\statusline.exe",
    "padding": 0
  },
  "env": {
    "STATUSLINE_FETCH": "1"
  }
}
```

**Windows (Git Bash/MinGW):**
```json
{
  "statusLine": {
    "type": "command",
    "command": "$HOME/.claude/statusline.exe",
    "padding": 0
  },
  "env": {
    "STATUSLINE_FETCH": "1"
  }
}
```

## Cross-platform Builds

```bash
make xwin    # Windows
make xlinux  # Linux  
make xmac    # macOS
```

## Requirements

- `git` in PATH

## License

MIT