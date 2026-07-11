package output

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/NITTC-Robosemi/stcm-viewer/src/model"
	"github.com/signintech/gopdf"
)

const (
	pdfPageWidth  = 595.28 // A4 width in points
	pdfPageHeight = 841.89 // A4 height in points
	pdfMargin     = 36.0
	pdfLeftMargin = 82.0  // wider left margin for Y-axis labels
	pdfTopMargin  = 30.0
	pdfPlotHeight = 100.0
	tickCount     = 5
)

// pdfColor represents an RGB color.
type pdfColor struct {
	R uint8
	G uint8
	B uint8
}

// hexToRGB converts a hex color string to RGB values.
func hexToRGB(hex string) (pdfColor, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return pdfColor{}, fmt.Errorf("invalid hex color: %s", hex)
	}
	var r, g, b uint8
	_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return pdfColor{}, fmt.Errorf("failed to parse hex color %s: %w", hex, err)
	}
	return pdfColor{R: r, G: g, B: b}, nil
}

// flattenTraces returns all traces as a flat slice, sorted by group and variable name.
func flattenTraces(allData model.ParsedData) []Trace {
	groups := make([]string, 0, len(allData))
	for group := range allData {
		groups = append(groups, group)
	}
	sort.Strings(groups)

	traces := make([]Trace, 0)
	colorIdx := 0
	for _, group := range groups {
		vars := make([]string, 0, len(allData[group]))
		for v := range allData[group] {
			vars = append(vars, v)
		}
		sort.Strings(vars)
		for _, v := range vars {
			data := allData[group][v]
			if len(data.X) == 0 {
				continue
			}
			start := data.X[0]
			xRel := make([]float64, len(data.X))
			for i, val := range data.X {
				xRel[i] = (val - start) / 1000.0
			}
			traces = append(traces, Trace{
				Name:  v,
				X:     xRel,
				Y:     data.Y,
				Color: colors[colorIdx%len(colors)],
			})
			colorIdx++
		}
	}
	return traces
}

// minMax returns the minimum and maximum values of a slice.
func minMax(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}
	min := values[0]
	max := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

// niceBounds returns axis limits that cover the data and align to nice tick values.
func niceBounds(min, max float64, count int) (float64, float64) {
	if min == max {
		min -= 1
		max += 1
	}
	step := niceNum((max-min)/float64(count-1), true)
	if step == 0 {
		step = 1
	}
	lo := math.Floor(min/step) * step
	hi := math.Ceil(max/step) * step
	if hi == max {
		hi += step
	}
	return lo, hi
}

// mapValue maps a value from one range to another.
func mapValue(v, srcMin, srcMax, dstMin, dstMax float64) float64 {
	if srcMax == srcMin {
		return (dstMin + dstMax) / 2
	}
	return dstMin + (v-srcMin)*(dstMax-dstMin)/(srcMax-srcMin)
}

// niceNum returns a "nice" number near x (round if round=true, ceiling otherwise).
func niceNum(x float64, round bool) float64 {
	if x == 0 {
		return 0
	}
	exp := math.Floor(math.Log10(math.Abs(x)))
	f := x / math.Pow(10, exp)
	var nf float64
	if round {
		if f < 1.5 {
			nf = 1
		} else if f < 3 {
			nf = 2
		} else if f < 7 {
			nf = 5
		} else {
			nf = 10
		}
	} else {
		if f <= 1 {
			nf = 1
		} else if f <= 2 {
			nf = 2
		} else if f <= 5 {
			nf = 5
		} else {
			nf = 10
		}
	}
	return nf * math.Pow(10, exp)
}

// niceTicks returns human-friendly tick positions for the range [min, max].
func niceTicks(min, max float64, count int) []float64 {
	rangeVal := niceNum(max-min, false)
	if rangeVal == 0 || count <= 1 {
		return []float64{min, max}
	}
	step := niceNum(rangeVal/float64(count-1), true)
	if step == 0 {
		step = 1
	}
	lo := math.Floor(min/step) * step
	hi := math.Ceil(max/step) * step
	n := int(math.Round((hi-lo)/step)) + 1
	ticks := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		tick := lo + step*float64(i)
		if tick < min-1e-12 || tick > max+1e-12 {
			continue
		}
		ticks = append(ticks, tick)
	}
	return ticks
}

// formatTick formats a tick value for display.
func formatTick(v float64) string {
	if v == 0 {
		return "0"
	}
	absV := math.Abs(v)
	switch {
	case absV >= 10000 || absV < 0.001:
		return fmt.Sprintf("%.2e", v)
	case absV >= 100:
		return fmt.Sprintf("%.0f", v)
	case absV >= 1:
		return fmt.Sprintf("%.2f", v)
	case absV >= 0.1:
		return fmt.Sprintf("%.3f", v)
	default:
		return fmt.Sprintf("%.4f", v)
	}
}

type fontCandidate struct {
	path   string
	name   string
	option *gopdf.TtfOption
}

// findFont searches for an available font on the current system.
func findFont(pdf *gopdf.GoPdf) (string, string, error) {
	candidates := []fontCandidate{}

	// Common Windows font directories.
	windir := os.Getenv("WINDIR")
	if windir == "" {
		windir = `C:\Windows`
	}

	switch runtime.GOOS {
	case "windows":
		// Prefer .ttf fonts on Windows; .ttc requires TtfOption with Style.
		candidates = append(candidates, []fontCandidate{
			{filepath.Join(windir, `Fonts\segoeui.ttf`), "segoeui", nil},
			{filepath.Join(windir, `Fonts\arial.ttf`), "arial", nil},
			{filepath.Join(windir, `Fonts\meiryo.ttc`), "meiryo", &gopdf.TtfOption{Style: gopdf.Regular}},
			{filepath.Join(windir, `Fonts\msgothic.ttc`), "msgothic", &gopdf.TtfOption{Style: gopdf.Regular}},
			{filepath.Join(windir, `Fonts\YuGothM.ttc`), "YuGothM", &gopdf.TtfOption{Style: gopdf.Regular}},
			{filepath.Join(windir, `Fonts\msyh.ttc`), "msyh", &gopdf.TtfOption{Style: gopdf.Regular}},
		}...)
	case "darwin":
		candidates = append(candidates, []fontCandidate{
			{"/System/Library/Fonts/Helvetica.ttc", "Helvetica", &gopdf.TtfOption{Style: gopdf.Regular}},
			{"/System/Library/Fonts/Supplemental/Arial.ttf", "Arial", nil},
			{"/Library/Fonts/Arial.ttf", "Arial2", nil},
		}...)
	default:
		// Linux and others.
		candidates = append(candidates, []fontCandidate{
			{"/usr/share/fonts/truetype/noto/NotoSans-Regular.ttf", "noto-latin", nil},
			{"/usr/share/fonts/opentype/noto/NotoSansCJK-Medium.ttc", "noto", &gopdf.TtfOption{Style: gopdf.Regular}},
			{"/usr/share/fonts/truetype/noto/NotoSansCJK-Regular.ttc", "noto-cjk", &gopdf.TtfOption{Style: gopdf.Regular}},
			{"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc", "noto-cjk2", &gopdf.TtfOption{Style: gopdf.Regular}},
		}...)
	}

	for _, c := range candidates {
		if _, err := os.Stat(c.path); err != nil {
			continue
		}
		if c.option != nil {
			if err := pdf.AddTTFFontWithOption(c.name, c.path, *c.option); err == nil {
				return c.path, c.name, nil
			}
		} else {
			if err := pdf.AddTTFFont(c.name, c.path); err == nil {
				return c.path, c.name, nil
			}
		}
	}
	return "", "", fmt.Errorf("no usable font found")
}

// GeneratePDF creates a PDF report with one plot per variable.
func GeneratePDF(allData model.ParsedData, outputPath string) error {
	traces := flattenTraces(allData)
	if len(traces) == 0 {
		return fmt.Errorf("no data to plot")
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	_, fontName, err := findFont(&pdf)
	if err != nil {
		return fmt.Errorf("failed to load any font: %w", err)
	}
	if err := pdf.SetFont(fontName, "", 10); err != nil {
		return fmt.Errorf("failed to set font: %w", err)
	}

	pdf.AddPage()
	pdf.SetXY(pdfMargin, pdfTopMargin)
	pdf.Cell(nil, "STCM Viewer PDF Report")

	y := pdfTopMargin + 24
	plotWidth := pdfPageWidth - pdfLeftMargin - pdfMargin

	for _, trace := range traces {
		blockHeight := pdfPlotHeight + 55
		if y+blockHeight > pdfPageHeight-pdfMargin {
			pdf.AddPage()
			y = pdfTopMargin
		}

		xMin, xMax := minMax(trace.X)
		xMin, xMax = niceBounds(xMin, xMax, tickCount)
		yMin, yMax := minMax(trace.Y)
		yMin, yMax = niceBounds(yMin, yMax, tickCount)

		plotLeft := pdfLeftMargin
		plotTop := y + 12

		// Center the trace title above the plot area.
		titleWidth := float64(len(trace.Name)) * 3.5
		pdf.SetXY(plotLeft+(plotWidth-titleWidth)/2, y)
		pdf.SetFont("", "", 10)
		pdf.SetTextColor(0, 0, 0)
		pdf.Cell(nil, trace.Name)

		plotRight := plotLeft + plotWidth
		plotBottom := plotTop + pdfPlotHeight

		// Draw axes.
		pdf.SetLineWidth(0.5)
		pdf.SetStrokeColor(0, 0, 0)
		pdf.Line(plotLeft, plotBottom, plotRight, plotBottom) // X axis
		pdf.Line(plotLeft, plotTop, plotLeft, plotBottom)     // Y axis

		// Draw grid lines and Y-axis ticks/labels.
		yTicks := niceTicks(yMin, yMax, tickCount)
		pdf.SetLineWidth(0.2)
		pdf.SetStrokeColor(200, 200, 200)
		for _, tick := range yTicks {
			gy := mapValue(tick, yMin, yMax, plotBottom, plotTop)
			pdf.Line(plotLeft, gy, plotRight, gy)
			label := formatTick(tick)
			textWidth := float64(len(label)) * 3.5
			pdf.SetXY(plotLeft-textWidth-8, gy-3)
			pdf.SetTextColor(80, 80, 80)
			pdf.SetFont("", "", 7)
			pdf.Cell(nil, label)
		}

		// Draw X-axis ticks and labels.
		xTicks := niceTicks(xMin, xMax, tickCount)
		pdf.SetLineWidth(0.5)
		pdf.SetStrokeColor(0, 0, 0)
		for _, tick := range xTicks {
			gx := mapValue(tick, xMin, xMax, plotLeft, plotRight)
			pdf.Line(gx, plotBottom, gx, plotBottom+3)
			label := fmt.Sprintf("%.2f", tick)
			textWidth := float64(len(label)) * 3.5
			pdf.SetXY(gx-textWidth/2, plotBottom+6)
			pdf.SetTextColor(80, 80, 80)
			pdf.SetFont("", "", 7)
			pdf.Cell(nil, label)
		}

		// Draw data line.
		color, err := hexToRGB(trace.Color)
		if err != nil {
			color = pdfColor{R: 0, G: 0, B: 0}
		}
		pdf.SetLineWidth(0.8)
		pdf.SetStrokeColor(color.R, color.G, color.B)

		for i := 1; i < len(trace.X); i++ {
			x1 := mapValue(trace.X[i-1], xMin, xMax, plotLeft, plotRight)
			y1 := mapValue(trace.Y[i-1], yMin, yMax, plotBottom, plotTop)
			x2 := mapValue(trace.X[i], xMin, xMax, plotLeft, plotRight)
			y2 := mapValue(trace.Y[i], yMin, yMax, plotBottom, plotTop)
			if math.IsNaN(x1) || math.IsNaN(y1) || math.IsNaN(x2) || math.IsNaN(y2) {
				continue
			}
			pdf.Line(x1, y1, x2, y2)
		}

		// Axis titles.
		pdf.SetFont("", "", 8)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetXY(plotRight-15, plotBottom+17)
		pdf.Cell(nil, "Time (s)")
		pdf.SetXY(8, plotTop-18)
		pdf.Cell(nil, "Value")

		y += blockHeight
	}

	if err := pdf.WritePdf(outputPath); err != nil {
		return fmt.Errorf("failed to write pdf: %w", err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("failed to stat pdf: %w", err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("pdf was written but has zero bytes")
	}

	return nil
}
