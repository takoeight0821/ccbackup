# Release Readiness Fixes

## 概要

リリースレビューで指摘された問題をすべて修正する。

---

## MUST Fix (4件)

### 1. README.md の作成

**ファイル**: `README.md` (新規作成)

```markdown
# ccbackup

Claude Code (~/.claude/) の履歴をバックアップするCLIツール。

## インストール

go install github.com/takoeight0821/ccbackup@latest

## 使い方

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

## 設定

~/.config/ccbackup/config.yaml

## 要件

- Go 1.21+
- Git
```

---

### 2. Git エラーメッセージの改善

**ファイル**: `internal/git/git.go`

`run()` 関数で stderr を取得し、エラーメッセージに含める:

```go
func (g *Git) run(args ...string) error {
    cmd := exec.Command("git", args...)
    cmd.Dir = g.Dir
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
    }
    return nil
}
```

**ファイル**: `internal/git/lfs.go`

同様に `run()` 関数を修正。

---

### 3. ソースディレクトリの存在確認

**ファイル**: `cmd/backup.go`

`syncer` 作成前に確認:

```go
if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
    return fmt.Errorf("source directory does not exist: %s", sourceDir)
}
```

---

### 4. バックアップディレクトリの初期化確認

**ファイル**: `cmd/backup.go`

Git 操作前に `.git` の存在を確認:

```go
gitDir := filepath.Join(backupDir, ".git")
if _, err := os.Stat(gitDir); os.IsNotExist(err) {
    return fmt.Errorf("backup directory not initialized, run 'ccbackup init --exec' first")
}
```

---

## SHOULD Fix (3件)

### 5. setDefaults() のエラー処理

**ファイル**: `cmd/root.go`

`initConfig()` で既にエラー処理されているため、`setDefaults()` は `initConfig()` から home を受け取る形にリファクタ:

```go
func setDefaults(home string) {
    viper.SetDefault("source_dir", filepath.Join(home, ".claude"))
    // ...
}
```

---

### 6. runConfigShow() のエラー処理

**ファイル**: `cmd/config.go`

エラーを適切にハンドル（表示目的なので警告として扱う）:

```go
sourceDir, err := paths.ExpandHome(viper.GetString("source_dir"))
if err != nil {
    sourceDir = viper.GetString("source_dir") + " (expansion failed)"
}
```

---

### 7. デフォルト backup_dir の汎用化

**ファイル**: `cmd/root.go`

組織固有のパスを汎用的なパスに変更:

```go
viper.SetDefault("backup_dir", filepath.Join(home, "claude-backup"))
```

---

## 変更ファイル一覧

| ファイル | 変更内容 |
|----------|----------|
| `README.md` | 新規作成 |
| `internal/git/git.go` | run() でエラー詳細を含める |
| `internal/git/lfs.go` | run() でエラー詳細を含める |
| `cmd/backup.go` | ソース/バックアップディレクトリの検証追加 |
| `cmd/root.go` | setDefaults のエラー処理、デフォルトパス変更 |
| `cmd/config.go` | ExpandHome のエラー処理 |

---

## 検証方法

```bash
# テスト実行
go test ./...

# ビルド
go build -o ccbackup .

# エラーメッセージ確認（存在しないディレクトリ）
./ccbackup backup --exec
# → "source directory does not exist: ..." が表示されること

# 未初期化バックアップディレクトリ
./ccbackup backup --exec
# → "backup directory not initialized..." が表示されること

# 正常動作確認
./ccbackup init --exec
./ccbackup backup
```

---

## コミット

```
fix: improve error handling and add README for release

- Add README.md with installation and usage instructions
- Include stderr in git command error messages
- Validate source directory exists before backup
- Check backup directory is initialized before git operations
- Fix ignored errors in setDefaults() and runConfigShow()
- Change default backup_dir to ~/claude-backup
```
