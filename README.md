# STCM Viewer

STM32CubeMonitor で記録した `.stcm` ファイルを、ブラウザで操作できる HTML グラフおよび印刷・保存向け PDF に変換するクロスプラットフォーム CLI ツールです。

## 主な機能

- `.stcm` ファイルの解析と CSV への一時変換
- Plotly.js によるインタラクティブ HTML グラフの生成
- 各変数ごとの PDF レポート出力（`--pdf` オプション）
- Linux（amd64 / arm64 / armv7 / .deb）および Windows インストーラー対応

## インストール

### Linux（.deb）

[Releases](https://github.com/NITTC-Robosemi/stcm-viewer/releases) からお使いのアーキテクチャに合わせた `.deb` ファイルをダウンロードしてください。

```bash
# amd64
sudo dpkg -i stcm-viewer-linux-amd64.deb
sudo apt-get install -f
```

インストール後、`stcm-viewer` コマンドが使えるようになります。

### Linux（バイナリ）

実行ファイルをダウンロードしてPATHの通った場所に配置します。

| ファイル名 | 対応環境 |
|---|---|
| `stcm-viewer-linux-amd64` | Linux x86_64 |
| `stcm-viewer-linux-arm64` | Linux ARM64 |
| `stcm-viewer-linux-armv7` | Linux ARM 32bit |

```bash
chmod +x stcm-viewer-linux-amd64
sudo mv stcm-viewer-linux-amd64 /usr/local/bin/stcm-viewer
```

### Windows

[Releases](https://github.com/NITTC-Robosemi/stcm-viewer/releases) から `stcm-viewer-setup.exe` をダウンロードして実行してください。

インストーラーが PATH を自動で設定します。コマンドプロンプトや PowerShell から `stcm-viewer` を使えます。

### ソースからビルド

必要な環境：

- Go 1.22 以上

```bash
git clone https://github.com/NITTC-Robosemi/stcm-viewer.git
cd stcm-viewer
go build -o stcm-viewer ./src/main.go
```

## 使い方

### 基本的な使い方

```bash
stcm-viewer your_log_file.stcm
```

実行結果：

- `<日時>.html` が生成されます
- 一時的な CSV フォルダは自動的に削除されます

### オプション

| オプション | 説明 |
|---|---|
| `--keep` | 変換後の CSV フォルダを保持します |
| `--pdf` | HTML に加えて PDF ファイルも生成します |

### 使用例

```bash
# CSV フォルダを保持
stcm-viewer your_log_file.stcm --keep

# PDF も生成
stcm-viewer your_log_file.stcm --pdf

# 両方指定
stcm-viewer your_log_file.stcm --pdf --keep
```

## 出力ファイル

| ファイル | 内容 |
|---|---|
| `<日時>.html` | Plotly.js を使用したインタラクティブグラフ |
| `<日時>.pdf` | 各変数ごとの折れ線グラフレポート（`--pdf` 時） |
| `<日時>/` | 変換された CSV ファイル群（`--keep` 時） |

## 日本語フォントについて（Linux）

PDF 出力で日本語を正しく表示するため、日本語フォントがインストールされている必要があります。

`.deb` パッケージをインストールすると `fonts-noto-cjk` が依存関係として導入されます。

手動でインストールする場合：

**Ubuntu / Debian:**

```bash
sudo apt-get update
sudo apt-get install fonts-noto-cjk
```

## 開発

### テスト

```bash
go test ./...
```

### Linux 向けクロスコンパイル

```bash
# amd64
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o stcm-viewer-linux-amd64 ./src/main.go

# arm64
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o stcm-viewer-linux-arm64 ./src/main.go

# armv7
GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w" -o stcm-viewer-linux-armv7 ./src/main.go
```

### Windows インストーラーのビルド（NSIS）

```bash
# Windows バイナリを先にビルド
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/windows/stcm-viewer.exe ./src/main.go

# NSIS でインストーラー作成
makensis packaging/windows/installer.nsi
```

### .deb パッケージのビルド

```bash
# amd64 の例
chmod +x stcm-viewer-linux-amd64
mkdir -p packaging/deb/amd64/DEBIAN packaging/deb/amd64/usr/local/bin
cp packaging/deb/DEBIAN/control packaging/deb/amd64/DEBIAN/control
sed -i 's/Architecture: .*/Architecture: amd64/' packaging/deb/amd64/DEBIAN/control
cp stcm-viewer-linux-amd64 packaging/deb/amd64/usr/local/bin/stcm-viewer
dpkg-deb --build packaging/deb/amd64 stcm-viewer-linux-amd64.deb
```

## GitHub Actions

タグを push すると、以下のアセットを自動でビルド・リリースします。

- Linux バイナリ（amd64 / arm64 / armv7）
- Linux .deb パッケージ（amd64 / arm64）
- Windows インストーラー（`stcm-viewer-setup.exe`）

```bash
git tag v2.0.0
git push origin v2.0.0
```

## トラブルシューティング

### エラー: ファイルが見つかりません

STCM ファイルのパスが正しいか確認してください。

### エラー: failed to load any font

PDF 生成時に日本語フォントが見つからない場合に発生します。`fonts-noto-cjk` などの日本語フォントをインストールしてください。

### Windows Defender の警告が出る

初めてダウンロードした実行ファイルは Windows Defender が警告を表示することがあります。「詳細情報」→「実行」をクリックして実行してください。
