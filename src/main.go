package main

import (
	"fmt"
	"os"
	"path/filepath"
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
	fmt.Fprintf(os.Stderr, "Usage: %s <stcm_file> [output_name] [--keep] [--pdf]\n", program)
}

func printDone() {
	fmt.Printf("\n%s✓%s %sDone%s\n\n", colorGreen, colorReset, colorBold, colorReset)
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
		fmt.Fprintf(os.Stderr, "%sエラー:%s ファイルが見つかりません: %s\n", colorRed, colorReset, stcmFile)
		os.Exit(1)
	}

	s := newSpinner("parsing...")
	s.Start()

	allData, err := parser.ParseSTCMFile(stcmFile)
	if err != nil {
		s.Fail(fmt.Errorf("変換に失敗しました: %w", err))
		os.Exit(1)
	}

	stcmFileName := filepath.Base(stcmFile)
	baseName := output.ResolveOutputName(stcmFileName, outputName)

	csvFolderName := "Converted"
	if idx := strings.Index(stcmFileName, "Log_"); idx != -1 && len(stcmFileName) > 26 {
		afterLog := stcmFileName[idx+len("Log_"):]
		if secondUnderscore := strings.Index(afterLog, "_"); secondUnderscore != -1 {
			rest := afterLog[secondUnderscore+1:]
			if len(rest) > 5 {
				csvFolderName = rest[:len(rest)-5]
			}
		}
	}

	parentDir := filepath.Dir(stcmFile)
	csvDir := filepath.Join(parentDir, csvFolderName)
	csvDir, err = output.WriteCSV(csvDir, allData)
	if err != nil {
		s.Fail(fmt.Errorf("CSV書き込みに失敗しました: %w", err))
		os.Exit(1)
	}

	s.SetMessage("generating HTML...")
	htmlPath := filepath.Join(parentDir, baseName+".html")
	if err := output.GenerateHTML(allData, htmlPath, "All Data"); err != nil {
		s.Fail(fmt.Errorf("HTML: %w", err))
		os.Exit(1)
	}

	pdfPath := ""
	if pdf {
		s.SetMessage("generating PDF...")
		pdfPath = filepath.Join(parentDir, baseName+".pdf")
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

