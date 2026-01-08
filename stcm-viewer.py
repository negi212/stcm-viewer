import pandas as pd
import matplotlib.pyplot as plt
from pathlib import Path
import numpy as np
from matplotlib.backends.backend_pdf import PdfPages
import plotly.graph_objects as go
import json
import os
import argparse
import shutil
import sys
from time import sleep

# 日本語フォント設定（OS別）
import platform
if platform.system() == 'Windows':
    plt.rcParams['font.sans-serif'] = ['Yu Gothic', 'MS Gothic']
elif platform.system() == 'Darwin':  # macOS
    plt.rcParams['font.sans-serif'] = ['Hiragino Sans', 'Hiragino Kaku Gothic Pro']
else:  # Linux
    plt.rcParams['font.sans-serif'] = ['Noto Sans CJK JP', 'IPAexGothic', 'DejaVu Sans']
plt.rcParams['axes.unicode_minus'] = False

# variablesフォルダのパス
variables_dir = Path(__file__).parent / "variables"
config_file = Path(__file__).parent / "graph_config.json"


def parse_stcm_file(filepath):
    """STCMファイルをCSVに変換"""
    print("stcmファイルを読み込み中...")
    
    try:
        with open(filepath, 'r') as file_in:
            data_string = file_in.read()
        
        if data_string[-3] == ",":
            data_string = data_string[:-3] + data_string[-2:]
        
        data = json.loads(data_string)
        print("データが読み込まれました。")
        
        list_group_name = []
        group_var_names = []
        list_column_groups = []
        
        for item in data:
            group_name = item['groupname']
            variable_name = item['variablename']
            variable_data = item['variabledata']
            
            if group_name not in list_group_name:
                list_group_name.append(group_name)
                group_var_names.append([])
                list_column_groups.append([])
            
            index_group = list_group_name.index(group_name)
            
            if variable_name not in group_var_names[index_group]:
                group_var_names[index_group].append(variable_name)
                list_column_groups[index_group].append([])
            
            index = group_var_names[index_group].index(variable_name)
            list_column_groups[index_group][index].extend(variable_data)
        
        print(".stcmファイルをパースしました。")
        
        # Pathlibを使用してクロスプラットフォーム対応
        filepath_path = Path(filepath)
        find_log_substr = filepath.find("Log_")
        log_substr = filepath[find_log_substr:]
        
        if find_log_substr > 0 and len(log_substr) > 26:
            pos = log_substr.find('_', log_substr.find('_') + 1)
            dir_name = log_substr[pos + 1: -5]
        else:
            dir_name = "Converted"
        
        dir_name = filepath_path.parent / dir_name
        if not dir_name.exists():
            dir_name.mkdir(parents=True)
        
        len_group_name = len(list_group_name)
        
        for g in range(len_group_name):
            name_group = list_group_name[g]
            print(f"グループ {name_group} を書き込み中...")
            
            dir_group_name = dir_name / name_group.strip().replace(',', '_').replace('.', '_')
            if not dir_group_name.exists():
                dir_group_name.mkdir(parents=True)
            
            len_variable_name = len(group_var_names[g])
            
            for i in range(len_variable_name):
                file_name = group_var_names[g][i].replace(".", "_").replace(":", "_").strip()
                csv_file_path = dir_group_name / f"{file_name}.csv"
                with open(csv_file_path, 'w') as file_out:
                    file_out.write(f"Time; {group_var_names[g][i]} \n")
                    for item in list_column_groups[g][i]:
                        time = item['x']
                        val = item['y']
                        string_out = f"{time};{val}\n"
                        file_out.write(string_out)
        
        print("変換が完了しました!")
        return str(dir_name)
    except Exception as e:
        print(f"エラーが発生しました: {str(e)}")
        return None


def delete_folder(folder_path):
    """フォルダを再帰的に削除"""
    try:
        if os.path.isdir(folder_path):
            shutil.rmtree(folder_path)
            return True
    except Exception as e:
        print(f"削除エラー: {str(e)}")
        return False





def generate_pdf(csv_folder=None, stcm_filename=None, generate_pdf_file=False):
    """HTMLグラフを生成（オプションでPDFも生成）"""
    try:
        # CSVフォルダを指定（デフォルトはvariablesフォルダ）
        if csv_folder is None:
            csv_folder = variables_dir
        else:
            csv_folder = Path(csv_folder)
        
        # CSVファイルを再帰的に取得
        csv_files = sorted(csv_folder.glob("**/*.csv"))
        
        if not csv_files:
            print(f"エラー: {csv_folder} 内にCSVファイルが見つかりません")
            return False
        
        # データを読み込む
        data_dict = {}
        for csv_file in csv_files:
            df = pd.read_csv(csv_file, sep=';')
            df.columns = df.columns.str.strip()
            data_dict[csv_file.stem] = df
        
        # ユーザーにタイトルを入力させる
        print("\nグラフのタイトルを入力してください（Enterキーでデフォルト名を使用）:")
        titles = {}
        for filename in data_dict.keys():
            user_input = input(f"{filename}: ").strip()
            titles[filename] = user_input if user_input else filename
        
        n_files = len(data_dict)
        
        # プロット作成（個別グラフ）
        n_cols = 2
        n_rows = (n_files + n_cols - 1) // n_cols
        
        fig1, axes = plt.subplots(n_rows, n_cols, figsize=(14, 4*n_rows))
        
        if n_files == 1:
            axes = np.array([axes])
        elif n_rows == 1:
            axes = axes.reshape(1, -1)
        
        axes = axes.flatten()
        
        # 各ファイルのデータをプロット
        for idx, (filename, df) in enumerate(data_dict.items()):
            ax = axes[idx]
            
            cols = df.columns.tolist()
            time_col = cols[0]
            value_col = cols[1]
            time_relative = (df[time_col] - df[time_col].iloc[0]) / 1000.0
            
            ax.plot(time_relative, df[value_col], linewidth=1, alpha=0.8)
            ax.set_xlabel('Time (s)')
            
            # ユーザーが指定したタイトルを使用
            title = titles.get(filename, filename)
            
            ax.set_ylabel(title)
            ax.set_title(f"図{idx + 1}："+title)
            ax.grid(True, alpha=0.3)
        
        # 余った領域を非表示に
        for idx in range(n_files, len(axes)):
            axes[idx].set_visible(False)
        
        # サブプロット間隔を調整（余白を大きくして線が見えるようにする）
        plt.subplots_adjust(hspace=0.5, wspace=0.5)
        
        # グラフ同士を区切る線を引く（すべての行と列の間）
        # 垂直方向（列）の区切り線
        if n_cols > 1:
            # 最初の行のグラフペアから位置を取得して全体に適用
            ax_left = axes[0] if 0 < n_files else None
            ax_right = axes[1] if 1 < n_files else None
            
            if ax_left and ax_right:
                x_left = ax_left.get_position().x1
                x_right = ax_right.get_position().x0
                x_pos = (x_left + x_right) / 2
                # すべての行を通して線を引く
                y_min = 0.02
                y_max = 0.98
                fig1.add_artist(plt.Line2D([x_pos, x_pos], [y_min, y_max], 
                                           transform=fig1.transFigure,
                                           color='black', linewidth=2, zorder=10))
        
        # 水平方向（行）の区切り線
        if n_rows > 1:
            for row in range(1, n_rows):
                # この行と前の行のグラフの位置を取得
                top_idx = (row - 1) * n_cols
                bottom_idx = row * n_cols
                
                ax_top = axes[top_idx] if top_idx < n_files else None
                ax_bottom = axes[bottom_idx] if bottom_idx < n_files else None
                
                if ax_top and ax_bottom:
                    y_top = ax_top.get_position().y0
                    y_bottom = ax_bottom.get_position().y1
                    y_pos = (y_top + y_bottom) / 2
                    # すべての列を通して線を引く
                    x_min = 0.02
                    x_max = 0.98
                    fig1.add_artist(plt.Line2D([x_min, x_max], [y_pos, y_pos], 
                                               transform=fig1.transFigure,
                                               color='black', linewidth=2, zorder=10))
        
        # Plotlyで全てのデータをインタラクティブに表示
        fig_interactive = go.Figure()
        
        colors = ['#1f77b4', '#ff7f0e', '#2ca02c', '#d62728', '#9467bd', '#8c564b', '#e377c2', '#7f7f7f', '#bcbd22', '#17becf']
        
        for idx, (filename, df) in enumerate(data_dict.items()):
            cols = df.columns.tolist()
            time_col = cols[0]
            value_col = cols[1]
            time_relative = (df[time_col] - df[time_col].iloc[0]) / 1000.0
            
            label = titles.get(filename, filename)
            color = colors[idx % len(colors)]
            
            fig_interactive.add_trace(go.Scatter(
                x=time_relative,
                y=df[value_col],
                mode='lines',
                name=label,
                line=dict(width=2, color=color),
                hovertemplate='<b>%{fullData.name}</b><br>Time: %{x:.2f}s<br>Value: %{y:.4f}<extra></extra>'
            ))
        
        fig_interactive.update_layout(
            title='All Data',
            xaxis_title='Time (s)',
            yaxis_title='Value',
            hovermode='x unified',
            plot_bgcolor='rgba(240,240,240,0.5)',
            width=1200,
            height=600,
            font=dict(size=11),
            legend=dict(x=1.05, y=1, xanchor='left', yanchor='top')
        )
        
        # HTMLファイル名を決定
        if stcm_filename:
            # ログファイル名から日時を抽出（例：Log_variables_2026-01-07_17h42m20s.stcm -> 2026-01-07_17h42m20s）
            import re
            match = re.search(r'(\d{4}-\d{2}-\d{2}_\d{2}h\d{2}m\d{2}s)', stcm_filename)
            if match:
                datetime_str = match.group(1)
                output_html = csv_folder.parent / f"{datetime_str}.html"
                output_pdf = csv_folder.parent / f"{datetime_str}.pdf"
            else:
                output_html = csv_folder.parent / "stcm_viewer_interactive.html"
                output_pdf = csv_folder.parent / "stcm_viewer_output.pdf"
        else:
            output_html = csv_folder.parent / "stcm_viewer_interactive.html"
            output_pdf = csv_folder.parent / "stcm_viewer_output.pdf"
        
        # インタラクティブグラフをHTMLで出力
        fig_interactive.write_html(str(output_html))
        print(f"インタラクティブグラフ出力完了: {output_html}")
        
        # PDFを生成する場合
        if generate_pdf_file:
            with PdfPages(output_pdf) as pdf:
                pdf.savefig(fig1)
                d = pdf.infodict()
                d['Title'] = 'STM32CubeMonitor Log Viewer'
                d['Producer'] = 'matplotlib'
            print(f"PDF出力完了: {output_pdf}")
        
        plt.close('all')
        return True
    
    except Exception as e:
        print(f"エラーが発生しました: {str(e)}")
        return False


def main():
    """メイン処理"""
    parser = argparse.ArgumentParser(description="STM32CubeMonitorのSTCMファイルをCSVに変換し、インタラクティブグラフを生成します")
    parser.add_argument("stcm_file", help="変換するSTCMファイルのパス")
    parser.add_argument("--keep", action="store_true", help="変換後のCSVフォルダを削除しない")
    parser.add_argument("--pdf", action="store_true", help="PDFも生成する")
    
    args = parser.parse_args()
    
    # STCMファイルの存在確認
    if not os.path.isfile(args.stcm_file):
        print(f"エラー: ファイルが見つかりません: {args.stcm_file}")
        sys.exit(1)
    
    print("=" * 60)
    print("STM32CubeMonitor STCM to CSV Converter")
    print("=" * 60)
    
    # ステップ1: STCMファイルをCSVに変換
    print("\n[ステップ1] STCMファイルをCSVに変換中...")
    converted_folder = parse_stcm_file(args.stcm_file)
    
    if not converted_folder:
        print("エラー: 変換に失敗しました")
        sys.exit(1)
    
    print(f"変換済みフォルダ: {converted_folder}")
    
    # ステップ2: HTMLグラフを生成（常に生成、--pdfでPDFも生成）
    print("\n[ステップ2] インタラクティブグラフを生成中...")
    stcm_filename = os.path.basename(args.stcm_file)
    if not generate_pdf(converted_folder, stcm_filename, args.pdf):
        print("警告: グラフの生成に失敗しました")
    
    # ステップ3: 必要に応じてCSVフォルダを削除
    if not args.keep:
        print("\n[ステップ3] CSVフォルダを削除中...")
        if delete_folder(converted_folder):
            print(f"フォルダを削除しました: {converted_folder}")
        else:
            print("警告: フォルダの削除に失敗しました")
    
    print("\n" + "=" * 60)
    print("処理が完了しました")
    print("=" * 60)


if __name__ == "__main__":
    main()
    # sleep(10)
