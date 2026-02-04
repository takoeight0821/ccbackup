# ccbackup - 設計ドキュメント

## 概要

`~/.claude/` のClaude Code履歴をOneDriveにバックアップし、Gitで世代管理するGoツール。

## CLI設計

```
ccbackup init [--exec]           # バックアップリポジトリ初期化 (Git + LFS)
ccbackup backup [--exec] [-v]    # バックアップ実行
ccbackup restore [--exec] [-v]   # リストア実行
ccbackup config show             # 設定表示
ccbackup config path             # 設定ファイルパス表示
```

### 安全設計: dry-runがデフォルト

- **デフォルト動作**: 変更内容を表示するのみ（dry-run）
- **`--exec`フラグ**: 実際にファイル操作・Git操作を実行
- **`-v, --verbose`**: 詳細出力

```bash
# 何が起こるか確認（デフォルト）
$ ccbackup backup
Would copy: projects/session.jsonl (1.2MB)
Would copy: history.jsonl (50KB)
...
Run with --exec to apply changes.

# 実際に実行
$ ccbackup backup --exec
Copying projects/session.jsonl...
Copying history.jsonl...
Committed: "Backup 2024-01-15 14:30"
```

## バックアップ対象

明示的にincludeされたファイル/ディレクトリのみがバックアップされます。

| 対象 | 内容 |
|------|------|
| `projects/` | セッション履歴 (JSONL) |
| `history.jsonl` | グローバル履歴 |
| `plans/` | プラン (Markdown) |
| `todos/` | タスク (JSON) |

**Note**: 上記以外のファイル（`file-history/`, `debug/`, `cache/`, `settings.json`など）は自動的に除外されます。

## 設定ファイル

`~/.config/ccbackup/config.yaml`:
```yaml
backup_dir: "/Users/y002168/OneDrive - Cybozu/claude-backup"
source_dir: "~/.claude"
include:
  - projects
  - history.jsonl
  - plans
  - todos
```

## プロジェクト構造

```
ccbackup/
├── go.mod
├── main.go
├── cmd/
│   ├── root.go        # Cobra/Viper初期化
│   ├── backup.go
│   ├── restore.go
│   ├── init_cmd.go    # initは予約語のためinit_cmd
│   └── config.go
└── internal/
    ├── sync/
    │   ├── sync.go
    │   ├── filter.go
    │   └── sync_test.go
    ├── git/
    │   ├── git.go
    │   └── lfs.go
    └── paths/
        ├── paths.go
        └── paths_test.go
```

**Note**: 設定管理はViperが担当するため、`internal/config`は不要。

## `ccbackup init` の処理内容

```
ccbackup init [--backup-dir PATH] [--exec]
```

**デフォルト動作（dry-run）:**
何が作成されるかを表示するのみ。

**`--exec`指定時の実行手順:**
1. 設定ファイル作成
   - `~/.config/ccbackup/config.yaml` が存在しなければ作成
   - `--backup-dir` 指定時はその値を使用、未指定時はデフォルト値
2. バックアップディレクトリ作成
   - `backup_dir` で指定されたパスを作成（存在しなければ）
3. Gitリポジトリ初期化
   - `git init` を実行
   - `.gitignore` を作成（.DS_Store等を除外）
4. Git LFS設定
   - `git lfs install` を実行
   - `.gitattributes` を作成してLFSパターンを設定
   - `file-history/**/*` をLFS追跡対象に
5. 初期コミット作成
   - 設定ファイル（.gitignore, .gitattributes）をコミット

**出力例:**
```bash
# dry-run（デフォルト）
$ ccbackup init
Would create config: ~/.config/ccbackup/config.yaml
Would create directory: /Users/.../OneDrive - Cybozu/claude-backup
Would run: git init
Would run: git lfs install
Would create: .gitattributes (LFS patterns)
Run with --exec to apply changes.

# 実際に実行
$ ccbackup init --exec
Created config: ~/.config/ccbackup/config.yaml
Created backup directory: /Users/.../OneDrive - Cybozu/claude-backup
Initialized git repository
Configured Git LFS for file-history/**/*
Ready! Run 'ccbackup backup --exec' to start backing up.
```

## Cobra/Viper構成例

```go
// cmd/root.go
var rootCmd = &cobra.Command{
    Use:   "ccbackup",
    Short: "Claude Code history backup tool",
}

func init() {
    cobra.OnInitialize(initConfig)
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
        "config file (default $HOME/.config/ccbackup/config.yaml)")
    rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
    rootCmd.PersistentFlags().Bool("exec", false, "actually execute (default: dry-run)")
    viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
    viper.BindPFlag("exec", rootCmd.PersistentFlags().Lookup("exec"))
}

func initConfig() {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        viper.AddConfigPath("$HOME/.config/ccbackup")
        viper.SetConfigName("config")
        viper.SetConfigType("yaml")
    }
    viper.SetEnvPrefix("CCBACKUP")
    viper.AutomaticEnv()
    viper.ReadInConfig()
}
```

## 主要アルゴリズム

### ファイル比較
```go
func NeedsSync(src, dst *FileInfo) bool {
    if dst == nil { return true }           // 新規
    if src.Size != dst.Size { return true } // サイズ差
    if src.ModTime.After(dst.ModTime) { return true } // 更新
    return false
}
```

### フィルター処理（Include方式）
- `projects` → projects/以下を含める
- `history.jsonl` → history.jsonlを含める
- `plans` → plans/以下を含める
- `todos` → todos/以下を含める
- 上記以外のファイルは自動的に除外される

## 依存関係

```go
// go.mod
require (
    github.com/spf13/cobra v1.10.0
    github.com/spf13/viper v1.20.0
)
```

- **Cobra**: CLIフレームワーク（サブコマンド、フラグ）
- **Viper**: 設定管理（YAML、環境変数、フラグの統合）

Note: YAMLはViperの依存として自動的に含まれる。直接使用する場合は `go.yaml.in/yaml/v3` を推奨（gopkg.in/yaml.v3 は非推奨）。
