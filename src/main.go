package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NITTC-Robosemi/stcm-viewer/src/output"
	"github.com/NITTC-Robosemi/stcm-viewer/src/parser"
)

func printUsage(program string) {
	fmt.Fprintf(os.Stderr, "Usage: %s <stcm_file> [output_name] [--keep] [--pdf]\n", program)
}

func main() {
	if len(os.Args) < 2 {
		printUsage(os.Args[0])
		os.Exit(1)
	}

	stcmFile := os.Args[1]
	outputName := ""
	keep := false
	pdf := false
	argIdx := 2
	if argIdx < len(os.Args) {
		arg := os.Args[argIdx]
		if arg != "--keep" && arg != "--pdf" {
			outputName = arg
			argIdx++
		}
	}

	for i := argIdx; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--keep":
			keep = true
		case "--pdf":
			pdf = true
		}
	}

	if _, err := os.Stat(stcmFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "エラー: ファイルが見つかりません: %s\n", stcmFile)
		os.Exit(1)
	}

	fmt.Println("============================================================")
	fmt.Println("STM32CubeMonitor STCM to CSV Converter")
	fmt.Println("============================================================")

	fmt.Println("\n[ステップ1] STCMファイルをCSVに変換中...")
	allData, err := parser.ParseSTCMFile(stcmFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: 変換に失敗しました: %v\n", err)
		os.Exit(1)
	}

	stcmFileName := filepath.Base(stcmFile)
	baseName := output.ResolveOutputName(stcmFileName, outputName)

	// Determine CSV output folder name.
	csvFolderName := "Converted"
	if idx := strings.Index(stcmFileName, "Log_"); idx != -1 && len(stcmFileName) > 26 {
		afterLog := stcmFileName[idx+len("Log_"):]
		if secondUnderscore := strings.Index(afterLog, "_"); secondUnderscore != -1 {
			rest := afterLog[secondUnderscore+1:]
			if len(rest) > 5 {
				csvFolderName = rest[:len(rest)-5] // without .stcm
			}
		}
	}

	parentDir := filepath.Dir(stcmFile)
	csvDir := filepath.Join(parentDir, csvFolderName)
	csvDir, err = output.WriteCSV(csvDir, allData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: CSV書き込みに失敗しました: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("変換済みフォルダ: %s\n", csvDir)

	fmt.Println("\n[ステップ2] インタラクティブグラフを生成中...")
	htmlPath := filepath.Join(parentDir, baseName+".html")
	if err := output.GenerateHTML(allData, htmlPath, "All Data"); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: HTML生成に失敗しました: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("インタラクティブグラフ出力完了: %s\n", htmlPath)

	pdfPath := ""
	if pdf {
		fmt.Println("\n[ステップ3] PDFレポートを生成中...")
		pdfPath = filepath.Join(parentDir, baseName+".pdf")
		if err := output.GeneratePDF(allData, pdfPath); err != nil {
			fmt.Fprintf(os.Stderr, "エラー: PDF生成に失敗しました: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("PDFレポート出力完了: %s\n", pdfPath)
	}

	if !keep {
		fmt.Println("\n[ステップ3] CSVフォルダを削除中...")
		if err := os.RemoveAll(csvDir); err != nil {
			fmt.Fprintf(os.Stderr, "警告: フォルダの削除に失敗しました: %v\n", err)
		} else {
			fmt.Printf("フォルダを削除しました: %s\n", csvDir)
		}
	}

	fmt.Println("\n============================================================")
	fmt.Println("処理が完了しました")
	fmt.Println("============================================================")
}

