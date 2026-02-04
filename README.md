# ccbackup

Claude Code (~/.claude/) の履歴をバックアップするCLIツール。

## インストール

```bash
go install github.com/takoeight0821/ccbackup@latest
```

## 使い方

```bash
# 初期化
ccbackup init --exec

# バックアップ (dry-run)
ccbackup backup

# バックアップ (実行)
ccbackup backup --exec

# リストア
ccbackup restore --exec

# 設定確認
ccbackup config show
```

## 設定

~/.config/ccbackup/config.yaml

## 要件

- Go 1.21+
- Git
