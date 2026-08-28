package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/NITTC-Robosemi/stcm-viewer/src/output"
	"github.com/NITTC-Robosemi/stcm-viewer/src/parser"
)

// ANSI color codes
const (
	colorReset = "\033[0m"
	colorBold  = "\033[1m"
	colorCyan  = "\033[36m"
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
)

// spinnerFrames is the animation sequence.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// spinner provides a loading animation in the terminal.
type spinner struct {
	msg    string
	msgMu  sync.Mutex
	stopCh chan struct{}
	wg     sync.WaitGroup
	start  time.Time
	prefix string
}

// newSpinner creates a new spinner with the given message and color prefix.
func newSpinner(msg string) *spinner {
	return &spinner{
		msg:    msg,
		stopCh: make(chan struct{}),
		prefix: colorCyan,
	}
}

// SetMessage updates the spinner message safely.
func (s *spinner) SetMessage(msg string) {
	s.msgMu.Lock()
	s.msg = msg
	s.msgMu.Unlock()
}

// Start begins the spinner animation.
func (s *spinner) Start() {
	s.start = time.Now()
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		i := 0
		for {
			select {
			case <-ticker.C:
				s.msgMu.Lock()
				msg := s.msg
				s.msgMu.Unlock()
				frame := s.prefix + spinnerFrames[i%len(spinnerFrames)] + colorReset
				elapsed := time.Since(s.start).Round(time.Second)
				fmt.Printf("\r%s %s (%s)%s", frame, msg, elapsed, colorReset)
				i++
			case <-s.stopCh:
				return
			}
		}
	}()
}

// Stop stops the spinner and clears the line.
func (s *spinner) Stop() {
	close(s.stopCh)
	s.wg.Wait()
	fmt.Print("\033[2K\r")
}

// Fail stops the spinner and leaves an error message.
func (s *spinner) Fail(err error) {
	close(s.stopCh)
	s.wg.Wait()
	fmt.Printf("\r%s✗%s %s %s%s%s\n", colorRed, colorReset, s.msg, colorRed, err.Error(), colorReset)
}

func printUsage(program string) {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s <stcm_file> [output_name] [--keep] [--pdf]\n", program)
	fmt.Fprintf(os.Stderr, "  %s <folder> [--keep] [--pdf] [--recursive]\n", program)
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  --keep       変換されたCSVフォルダを残す\n")
	fmt.Fprintf(os.Stderr, "  --pdf        HTMLに加えてPDFも生成する\n")
	fmt.Fprintf(os.Stderr, "  --recursive  フォルダ指定時、サブフォルダも再帰的に探索する (-r)\n")
}

func printDone() {
	fmt.Printf("\n%s✓%s %sDone%s\n\n", colorGreen, colorReset, colorBold, colorReset)
}

func collectSTCMFiles(dir string, recursive bool) ([]string, error) {
	var files []string
	if recursive {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.EqualFold(filepath.Ext(path), ".stcm") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if strings.EqualFold(filepath.Ext(e.Name()), ".stcm") {
				files = append(files, filepath.Join(dir, e.Name()))
			}
		}
	}
	sort.Strings(files)
	return files, nil
}

// uniqueFilePath returns a file path that does not yet exist by appending _1, _2, ...
func uniqueFilePath(base string) string {
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return base
	}
	ext := filepath.Ext(base)
	withoutExt := strings.TrimSuffix(base, ext)
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s_%d%s", withoutExt, i, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

// uniqueDirPath returns a directory path that does not yet exist.
func uniqueDirPath(base string) string {
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return base
	}
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s_%d", base, i)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

func resolveCSVFolderName(stcmFileName string) string {
	if idx := strings.Index(stcmFileName, "Log_"); idx != -1 && len(stcmFileName) > 26 {
		afterLog := stcmFileName[idx+len("Log_"):]
		if secondUnderscore := strings.Index(afterLog, "_"); secondUnderscore != -1 {
			rest := afterLog[secondUnderscore+1:]
			if len(rest) > 5 {
				return rest[:len(rest)-5] // without .stcm
			}
		}
	}
	// Fallback: use file name without extension so renamed/custom files keep their name.
	trimmed := strings.TrimSuffix(stcmFileName, filepath.Ext(stcmFileName))
	trimmed = strings.TrimSpace(trimmed)
	if trimmed != "" {
		return trimmed
	}
	return "Converted"
}

// processSingleFileBatch handles one file in batch mode (no spinner, verbose per-file output).
func processSingleFileBatch(stcmFile string, outputName string, keep bool, pdf bool) error {
	fmt.Println("============================================================")
	fmt.Printf("処理対象: %s\n", stcmFile)
	fmt.Println("============================================================")

	fmt.Println("\n[ステップ1] STCMファイルをCSVに変換中...")
	allData, err := parser.ParseSTCMFile(stcmFile)
	if err != nil {
		return fmt.Errorf("変換に失敗しました: %w", err)
	}

	stcmFileName := filepath.Base(stcmFile)
	baseName := output.ResolveOutputName(stcmFileName, outputName)
	csvFolderName := resolveCSVFolderName(stcmFileName)

	parentDir := filepath.Dir(stcmFile)
	csvDir := filepath.Join(parentDir, csvFolderName)
	if keep {
		csvDir = uniqueDirPath(csvDir)
	}
	csvDir, err = output.WriteCSV(csvDir, allData)
	if err != nil {
		return fmt.Errorf("CSV書き込みに失敗しました: %w", err)
	}
	fmt.Printf("変換済みフォルダ: %s\n", csvDir)

	fmt.Println("\n[ステップ2] インタラクティブグラフを生成中...")
	htmlPath := uniqueFilePath(filepath.Join(parentDir, baseName+".html"))
	if err := output.GenerateHTML(allData, htmlPath, "All Data"); err != nil {
		return fmt.Errorf("HTML生成に失敗しました: %w", err)
	}
	fmt.Printf("インタラクティブグラフ出力完了: %s\n", htmlPath)

	if pdf {
		fmt.Println("\n[ステップ3] PDFレポートを生成中...")
		pdfPath := uniqueFilePath(filepath.Join(parentDir, baseName+".pdf"))
		if err := output.GeneratePDF(allData, pdfPath); err != nil {
			return fmt.Errorf("PDF生成に失敗しました: %w", err)
		}
		fmt.Printf("PDFレポート出力完了: %s\n", pdfPath)
	}

	if !keep {
		stepLabel := "[ステップ3]"
		if pdf {
			stepLabel = "[ステップ4]"
		}
		fmt.Printf("\n%s CSVフォルダを削除中...\n", stepLabel)
		if err := os.RemoveAll(csvDir); err != nil {
			fmt.Fprintf(os.Stderr, "警告: フォルダの削除に失敗しました: %v\n", err)
		} else {
			fmt.Printf("フォルダを削除しました: %s\n", csvDir)
		}
	}

	fmt.Println("\n--- 完了 ---")
	return nil
}

func main() {
	if len(os.Args) < 2 {
		printUsage(os.Args[0])
		os.Exit(1)
	}

	inputPath := os.Args[1]
	if inputPath == "--help" || inputPath == "-h" {
		printUsage(os.Args[0])
		os.Exit(0)
	}

	keep := false
	pdf := false
	recursive := false
	var positional []string

	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--keep":
			keep = true
		case "--pdf":
			pdf = true
		case "--recursive", "-r":
			recursive = true
		case "--help", "-h":
			printUsage(os.Args[0])
			os.Exit(0)
		default:
			if strings.HasPrefix(os.Args[i], "-") {
				fmt.Fprintf(os.Stderr, "エラー: 不明なオプション: %s\n", os.Args[i])
				printUsage(os.Args[0])
				os.Exit(1)
			}
			positional = append(positional, os.Args[i])
		}
	}

	info, err := os.Stat(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sエラー:%s ファイル/フォルダが見つかりません: %s\n", colorRed, colorReset, inputPath)
		os.Exit(1)
	}

	if info.IsDir() {
		if len(positional) > 0 {
			fmt.Fprintf(os.Stderr, "警告: フォルダ指定時は出力名指定は無視されます: %v\n", positional)
		}
		stcmFiles, err := collectSTCMFiles(inputPath, recursive)
		if err != nil {
			fmt.Fprintf(os.Stderr, "エラー: フォルダの読み取りに失敗しました: %v\n", err)
			os.Exit(1)
		}
		if len(stcmFiles) == 0 {
			fmt.Fprintf(os.Stderr, "エラー: フォルダ内に .stcm ファイルが見つかりませんでした: %s\n", inputPath)
			if recursive {
				fmt.Fprintf(os.Stderr, "（再帰的に探索しました）\n")
			}
			os.Exit(1)
		}

		fmt.Println("============================================================")
		fmt.Println("STM32CubeMonitor STCM to CSV Converter (Batch Mode)")
		fmt.Println("============================================================")
		fmt.Printf("対象フォルダ: %s\n", inputPath)
		fmt.Printf("再帰的探索: %v\n", recursive)
		fmt.Printf("対象ファイル数: %d\n", len(stcmFiles))
		for _, f := range stcmFiles {
			fmt.Printf("  - %s\n", f)
		}
		fmt.Println()

		successCount := 0
		failCount := 0
		for idx, f := range stcmFiles {
			fmt.Printf("\n[%d/%d] ", idx+1, len(stcmFiles))
			if err := processSingleFileBatch(f, "", keep, pdf); err != nil {
				fmt.Fprintf(os.Stderr, "エラー [%s]: %v\n", f, err)
				failCount++
			} else {
				successCount++
			}
		}

		fmt.Println("\n============================================================")
		fmt.Printf("バッチ処理が完了しました: 成功 %d / 失敗 %d / 合計 %d\n", successCount, failCount, len(stcmFiles))
		fmt.Println("============================================================")
		if failCount > 0 {
			os.Exit(1)
		}
		return
	}

	// Single file mode - use spinner for nice UX
	if len(positional) > 0 {
		// positional[0] is outputName if present
		if len(positional) > 1 {
			fmt.Fprintf(os.Stderr, "警告: 余分な引数は無視されます: %v\n", positional[1:])
		}
	}
	outputName := ""
	if len(positional) > 0 {
		outputName = positional[0]
	}

	// Check recursive flag is ignored in single file mode
	if recursive {
		fmt.Fprintf(os.Stderr, "警告: 単一ファイル指定時は --recursive は無視されます\n")
	}

	s := newSpinner("parsing...")
	s.Start()

	allData, err := parser.ParseSTCMFile(inputPath)
	if err != nil {
		s.Fail(fmt.Errorf("変換に失敗しました: %w", err))
		os.Exit(1)
	}

	stcmFileName := filepath.Base(inputPath)
	baseName := output.ResolveOutputName(stcmFileName, outputName)
	csvFolderName := resolveCSVFolderName(stcmFileName)

	parentDir := filepath.Dir(inputPath)
	csvDir := filepath.Join(parentDir, csvFolderName)
	if keep {
		csvDir = uniqueDirPath(csvDir)
	}
	csvDir, err = output.WriteCSV(csvDir, allData)
	if err != nil {
		s.Fail(fmt.Errorf("CSV書き込みに失敗しました: %w", err))
		os.Exit(1)
	}

	s.SetMessage("generating HTML...")
	htmlPath := uniqueFilePath(filepath.Join(parentDir, baseName+".html"))
	if err := output.GenerateHTML(allData, htmlPath, "All Data"); err != nil {
		s.Fail(fmt.Errorf("HTML: %w", err))
		os.Exit(1)
	}

	pdfPath := ""
	if pdf {
		s.SetMessage("generating PDF...")
		pdfPath = uniqueFilePath(filepath.Join(parentDir, baseName+".pdf"))
		if err := output.GeneratePDF(allData, pdfPath); err != nil {
			s.Fail(fmt.Errorf("PDF: %w", err))
			os.Exit(1)
		}
	}

	if !keep {
		s.SetMessage("cleaning up...")
		if err := os.RemoveAll(csvDir); err != nil {
			s.Fail(fmt.Errorf("cleanup: %w", err))
			os.Exit(1)
		}
	}

	s.Stop()
	printDone()

	fmt.Printf("  %s→%s %s%s%s\n", colorGreen, colorReset, colorBold, htmlPath, colorReset)
	if pdfPath != "" {
		fmt.Printf("  %s→%s %s%s%s\n", colorGreen, colorReset, colorBold, pdfPath, colorReset)
	}
}
