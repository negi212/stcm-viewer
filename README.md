# STCM Viewer

STM32CubeMonitor で記録した `.stcm` ファイルを、ブラウザで操作できる HTML グラフや印刷・保存向け PDF に変換するツールです。

## できること

- `.stcm` ファイルから HTML グラフを生成
- ブラウザ上でズームや凡例の表示・非表示ができるインタラクティブなグラフを確認
- `--pdf` オプションで各変数ごとの PDF レポートを出力

## インストール

[Releases](https://github.com/NITTC-Robosemi/stcm-viewer/releases) から、お使いの環境に合ったファイルをダウンロードしてください。

### Linux

#### おすすめ：.deb パッケージ

Debian 系（Ubuntu など）をお使いの場合は `.deb` を推奨します。

```bash
# 64bit PC の場合
sudo dpkg -i stcm-viewer-linux-amd64.deb
sudo apt-get install -f
```

インストール後、ターミナルで `stcm-viewer` と入力して使えます。

#### その他の Linux

バイナリファイルをダウンロードし、PATH の通った場所に置いてください。

```bash
chmod +x stcm-viewer-linux-amd64
sudo mv stcm-viewer-linux-amd64 /usr/local/bin/stcm-viewer
```

| ファイル名 | 対応環境 |
|---|---|
| `stcm-viewer-linux-amd64` | 64bit PC（x86_64） |
| `stcm-viewer-linux-arm64` | ARM64 |
| `stcm-viewer-linux-armv7` | ARM 32bit |

### Windows

`stcm-viewer-setup.exe` をダウンロードして実行してください。インストーラーが PATH を自動で設定するので、コマンドプロンプトや PowerShell から `stcm-viewer` と入力して使えます。

## 使い方

### 基本的な使い方

```bash
stcm-viewer your_log_file.stcm
```

実行すると、以下のファイルが生成されます。

- `<日時>.html` — ブラウザで開くグラフ
- 一時的な CSV フォルダは自動で削除されます

### オプション

| オプション | 説明 |
|---|---|
| `--pdf` | HTML に加えて PDF も生成します |
| `--keep` | 変換された CSV フォルダを残します |

### 使用例

```bash
# PDF も一緒に出力
stcm-viewer your_log_file.stcm --pdf

# CSV も残す
stcm-viewer your_log_file.stcm --keep

# 両方指定
stcm-viewer your_log_file.stcm --pdf --keep
```

## 出力ファイル

| ファイル | 内容 |
|---|---|
| `<日時>.html` | ブラウザで操作できるグラフ |
| `<日時>.pdf` | 各変数ごとのグラフレポート（`--pdf` 時） |
| `<日時>/` | 変換された CSV ファイル群（`--keep` 時） |

## PDF で日本語が表示されない場合

PDF 出力には日本語フォントが必要です。`.deb` を使っている場合は自動で入りますが、手動でインストールする場合は以下を実行してください。

```bash
sudo apt-get update
sudo apt-get install fonts-noto-cjk
```

## 困ったとき

### ファイルが見つからないと言われる

`.stcm` ファイルのパスが正しいか確認してください。ファイル名にスペースが含まれる場合は `"` で囲ってください。

```bash
stcm-viewer "my log file.stcm"
```

### Windows で保護されましたと表示される

初めてダウンロードした実行ファイルは Windows Defender が警告を出すことがあります。「詳細情報」→「実行」を選択してください。

### グラフが表示されない

生成された `.html` ファイルをブラウザで開き直してください。インターネット接続が必要です（Plotly.js を CDN から読み込みます）。
