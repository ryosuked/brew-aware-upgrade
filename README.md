# brew-aware-upgrade

**brew-aware-upgrade** は、ネットワーク環境に応じて Homebrew パッケージのアップグレードを「カテゴリ別」に制御する CLI ツールです。

たとえば、**通信が制限されているテザリング環境**では重要なパッケージのみを更新し、**Wi-Fi環境など**であれば一括アップグレードを実行できます。

---

## 🚀 特徴

- パッケージを YAML でカテゴリ分け（重要・通常・大容量）
- `brew outdated` で更新対象を取得し、カテゴリごとにアップグレードを実施
- 通信量を抑えるための `dry-run` や `verbose` モードを搭載
- 設定ファイルは柔軟に検索（環境変数や `$HOME/.brewall` にも対応）

---

## 💡 利用例

```bash
# 全てのカテゴリを一括アップグレード（= 通信制限のない Wi-Fi 環境）
brew-aware-upgrade

# 優先度の高いものだけアップグレード（= テザリング環境など）
brew-aware-upgrade -P

# 指定カテゴリのみ実行（大容量は回避）
brew-aware-upgrade -c highest_priority,priority

# dry-runで確認
brew-aware-upgrade -D -c large_size
```

---

## 🔧 インストール

```bash
go install github.com/yourusername/brew-aware-upgrade@latest
```

または手動ビルド:

```bash
go build -o brew-aware-upgrade
```

---

## 🧩 オプション

| オプション        | 説明                                                                 |
|-------------------|----------------------------------------------------------------------|
| `-c <cats>`       | カンマ区切りでカテゴリを指定（例: `-c highest_priority,large_size`）     |
| `-P`              | `highest_priority` + `priority` のみを実行                            |
| `-D`              | Dry-run（実行せずに出力確認のみ）                                     |
| `-v`              | 詳細なログ出力を有効化                                                |
| `-h`              | ヘルプ表示                                                            |

---

## 📁 設定ファイル構成（YAML）

```yaml
categories:
  highest_priority:
    - git
    - curl
  priority:
    - ffmpeg
    - jq
  large_size:
    - android-studio
    - xcode
```

### 🔍 検索される設定ファイルのパス

優先順:

1. 実行ファイルと同じディレクトリ
2. `$HOME/.brewall/packages.yaml`
3. カレントディレクトリ
4. 環境変数 `BREWALL_CONFIG_PATHS` で指定されたディレクトリ群

---

## 📦 出力例（Dry Run）

```bash
$ brew-aware-upgrade -D -c highest_priority
Using config file: /Users/foo/.brewall/packages.yaml
[VERBOSE] Executing: brew outdated --quiet --greedy
[VERBOSE] Outdated packages: map[git:true curl:true]
Upgrading highest_priority packages: [git curl]
```

---

## 🔁 エイリアス設定（Zsh 推奨）

`brew-aware-upgrade` を毎回入力するのが手間な場合は、以下のように `.zshrc` にエイリアスを設定できます。

```bash
# ~/.zshrc に追加
alias brewup="brew-aware-upgrade"
```

設定後は、以下のように簡単にコマンドを呼び出せます：

```bash
# 優先度の高いパッケージのみアップグレード
brewup -P

# 特定カテゴリだけアップグレード
brewup -c priority

# 通信なしで確認（dry run）
brewup -D -c large_size
```

設定を反映するには次を実行してください：

```bash
source ~/.zshrc
```

---

## 📘 ライセンス

MIT License

---

## 🙋‍♀️ 用途の例

- 出張先やカフェで最低限のセキュリティ更新だけ実施
- 自宅 Wi-Fi 接続時に一括アップグレード
- 月間通信量制限がある環境でのダウンロード制御