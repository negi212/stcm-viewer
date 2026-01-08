# STCM Viewer

## 概要

STCM Viewerは、STM32CubeMonitorで記録したデータファイル（.stcm）を、見やすいグラフに変換するツールです。

データをブラウザで表示できるHTMLファイルとして出力し、グラフの拡大・縮小や詳細確認などの操作が可能です。複数のデータを一度に表示して比較することもできます。オプションで印刷や配布に適したPDFファイルも出力できます。

### 主な機能
- データファイルの読み込みと変換
- 操作可能なHTMLグラフの生成
- PDFレポートの出力（オプション）
- データのグループ分け管理
- グラフタイトルのカスタマイズ

## インストール

### 実行ファイルのダウンロード

[Releases](https://github.com/NITTC-Robosemi/stcm-viewer/releases)から、お使いのOSに対応した実行ファイルをダウンロードしてください。

- **Windows**: `stcm-viewer.exe`
- **Linux**: `stcm-viewer`
- **macOS**: `stcm-viewer`

### Linux/macOSでの準備

ダウンロード後、実行権限を付与してください：

```bash
chmod +x stcm-viewer
```

### 日本語フォントのインストール（Linux のみ）

日本語を正しく表示するため、日本語フォントをインストールしてください：

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install fonts-noto-cjk fonts-ipafont
```

**Fedora/RHEL:**
```bash
sudo dnf install google-noto-sans-cjk-jp-fonts ipa-gothic-fonts
```

**Arch Linux:**
```bash
sudo pacman -S noto-fonts-cjk adobe-source-han-sans-jp-fonts
```

## 使い方

### 基本的な使い方

ターミナル（コマンドプロンプト）で、ダウンロードした実行ファイルにSTCMファイルのパスを指定して実行します。

**Windows:**
```cmd
stcm-viewer.exe your_log_file.stcm
```

**Linux/macOS:**
```bash
./stcm-viewer your_log_file.stcm
```

**PATHを通している場合:**
```bash
stcm-viewer your_log_file.stcm
```

### コマンドラインオプション

#### 必須引数
- `stcm_file`: 変換するSTCMファイルのパス

#### オプション引数
- `--keep`: 変換後のCSVフォルダを保持する（デフォルトでは削除されます）
- `--pdf`: HTMLに加えてPDFファイルも生成する

### 使用例

#### 例1: 基本的な変換とHTML生成
```bash
# Windows
stcm-viewer.exe your_log_file.stcm

# Linux/macOS
./stcm-viewer your_log_file.stcm
```
実行結果:
- CSVファイルが一時的に生成されます
- インタラクティブなHTMLグラフが生成されます
- CSVフォルダは自動的に削除されます

#### 例2: CSVフォルダを保持する
```bash
# Windows
stcm-viewer.exe your_log_file.stcm --keep

# Linux/macOS
./stcm-viewer your_log_file.stcm --keep
```
CSVファイルが保持され、後で確認・再利用できます。

#### 例3: PDFも生成する
```bash
# Windows
stcm-viewer.exe your_log_file.stcm --pdf

# Linux/macOS
./stcm-viewer your_log_file.stcm --pdf
```
HTMLファイルとPDFファイルの両方が生成されます。

### 機能詳細

#### 出力されるファイル
- **HTMLファイル**: `<日時>.html`
  - ブラウザで操作可能（ズーム、データ確認など）
  - すべてのデータを重ねて表示し、比較が容易
- **PDFファイル**: `<日時>.pdf`（`--pdf`オプション時のみ）
  - 印刷や配布向けの固定レイアウト
  - 各変数が個別グラフとして配置

#### グラフのカスタマイズ
プログラム実行中、各グラフのタイトルを変更できます。Enterキーでデフォルト名を使用します。

#### 使い分けの目安
- **HTML**: データ分析、詳細確認
- **PDF**: 報告書、記録保存

## トラブルシューティング

### エラー: ファイルが見つかりません
STCMファイルのパスが正しいか確認してください。相対パスまたは絶対パスで指定できます。

### エラー: CSVファイルが見つかりません
変換処理が正常に完了しているか確認してください。STCMファイルの形式が正しいことを確認してください。

### 日本語フォントが表示されない（Linux）
システムに日本語フォントがインストールされているか確認してください（インストールセクション参照）。

### Windows Defenderの警告が出る
初めてダウンロードした実行ファイルは、Windows Defenderが警告を表示することがあります。
「詳細情報」→「実行」をクリックして実行してください。

---

## 開発者向け情報

### Pythonスクリプトとして実行する

ソースコードから直接実行する場合：

#### 必要な環境
- Python 3.7以上

#### 依存パッケージのインストール
```bash
pip install pandas matplotlib numpy plotly
```

#### 実行方法
```bash
python stcm.py your_log_file.stcm
```

### 実行ファイルのビルド

PyInstallerを使用してビルドできます：

```bash
pip install pyinstaller
pyinstaller --onefile --name stcm-viewer stcm.py
```

ビルドされた実行ファイルは`dist/`フォルダに生成されます。

---
