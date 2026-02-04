# ccbackup

CLI tool to backup ~/.claude/ history with Git version control.

## Commands

```bash
# Build
go build -o ccbackup .

# Test
go test ./...

# Run (dry-run by default)
./ccbackup backup
./ccbackup backup --exec  # Actually execute
```

## Architecture

```
cmd/           # Cobra commands (init, backup, restore, config)
internal/
  git/         # Git/LFS wrapper (run() includes stderr in errors)
  paths/       # Path utilities (ExpandHome)
  sync/        # File sync with include-pattern filtering
```

## Key Patterns

- **Dry-run by default**: All mutating commands require `--exec` flag
- **Config**: viper-based, `~/.config/ccbackup/config.yaml`
- **Include filter**: Only sync files matching `include` patterns (projects, history.jsonl, plans, todos)

## Gotchas

- `backup --exec` requires `init --exec` first (validates .git exists)
- Source directory must exist before backup
- Tests use minimal .git structure (not full git init)
