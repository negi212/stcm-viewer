# STCM Viewer

STM32CubeMonitor で記録した `.stcm` ファイルを、ブラウザで操作できる HTML グラフや印刷・保存向け PDF に変換するツールです。

## できること

- `.stcm` ファイルから HTML グラフを生成
- ブラウザ上でズームや凡例の表示・非表示ができるインタラクティブなグラフを確認
- `--pdf` オプションで各変数ごとの PDF レポートを出力
- 出力ファイル名を自由に指定可能

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
# 単一ファイルを変換
stcm-viewer your_log_file.stcm

# フォルダ内の全 .stcm を一括変換
stcm-viewer ./logs/

# サブフォルダも再帰的に処理
stcm-viewer ./logs/ --recursive
```

実行すると、以下のファイルが生成されます。

- `<ファイル名>.html` — ブラウザで開くグラフ
- 一時的な CSV フォルダは自動で削除されます
- フォルダ指定時は、フォルダ内の各 `.stcm` ごとに `<日時>.html`（と `--pdf` 時は `.pdf`）が生成されます。同名ファイルが既に存在する場合は `_1`, `_2` … が自動で付与されます。

### 出力名を指定する

2 番目の引数で出力ファイル名を指定できます。

```bash
stcm-viewer your_log_file.stcm my_output
# → my_output.html が生成される
```

### オプション

| オプション | 説明 |
|---|---|
| `--pdf` | HTML に加えて PDF も生成します |
| `--keep` | 変換された CSV フォルダを残します |
| `--recursive` / `-r` | フォルダ指定時、サブフォルダも再帰的に探索します |

### 使用例

```bash
# PDF も一緒に出力
stcm-viewer your_log_file.stcm --pdf

# 出力名を指定して PDF も生成
stcm-viewer your_log_file.stcm my_output --pdf

# CSV も残す
stcm-viewer your_log_file.stcm --keep

# 両方指定
stcm-viewer your_log_file.stcm --pdf --keep

# フォルダ内の全ファイルを変換（PDF 付き）
stcm-viewer ./logs/ --pdf

# フォルダを再帰的に変換し CSV も残す
stcm-viewer ./logs/ --recursive --keep

# フォルダを再帰的に変換し PDF も出力
stcm-viewer ./logs/ --recursive --pdf
```

## 出力ファイル

| ファイル | 内容 |
|---|---|
| `<出力名>.html` | ブラウザで操作できるグラフ |
| `<出力名>.pdf` | 各変数ごとのグラフレポート（`--pdf` 時） |
| `<日時>/` | 変換された CSV ファイル群（`--keep` 時） |

## 困ったとき

### ファイルが見つからないと言われる

`.stcm` ファイルのパスが正しいか確認してください。ファイル名にスペースが含まれる場合は `"` で囲ってください。

```bash
stcm-viewer "my log file.stcm"
```

### PDF 生成でフォントエラーになる（Windows）

Windows 環境で以下のようなエラーが出る場合は、対応するフォントが見つからないためです。

```
failed to load any font: CreateFile /usr/share/fonts/...: The system cannot find the path specified.
```

最新版では Windows のシステムフォントを自動で検出するようになっています。問題が続く場合は、`C:\Windows\Fonts` に `segoeui.ttf` または `arial.ttf` が存在することを確認してください。

### Windows で保護されましたと表示される

初めてダウンロードした実行ファイルは Windows Defender が警告を出すことがあります。「詳細情報」→「実行」を選択してください。

### グラフが表示されない

生成された `.html` ファイルをブラウザで開き直してください。インターネット接続が必要です（Plotly.js を CDN から読み込みます）。
