package output

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/NITTC-Robosemi/stcm-viewer/src/model"
)

const defaultPlotlyURL = "https://cdn.plot.ly/plotly-2.27.0.min.js"

// Trace holds data for one Plotly trace.
type Trace struct {
	Name  string
	X     []float64
	Y     []float64
	Color string
}

// HTMLData is passed to the HTML template.
type HTMLData struct {
	Title    string
	PlotlyJS template.URL
	Traces   []Trace
}

// colors matches the original Plotly default color cycle.
var colors = []string{
	"#1f77b4", "#ff7f0e", "#2ca02c", "#d62728", "#9467bd",
	"#8c564b", "#e377c2", "#7f7f7f", "#bcbd22", "#17becf",
}

// sanitizeForJS escapes single quotes in a string for safe JS output.
func sanitizeForJS(s string) string {
	return strings.ReplaceAll(s, "'", `\\'`)
}

// buildTraces flattens parsed data into Plotly traces, sorted by variable name.
func buildTraces(allData model.ParsedData) []Trace {
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
		for _, varName := range vars {
			data := allData[group][varName]
			if len(data.X) == 0 {
				continue
			}
			start := data.X[0]
			xRel := make([]float64, len(data.X))
			for i, v := range data.X {
				xRel[i] = (v - start) / 1000.0
			}
			traces = append(traces, Trace{
				Name:  sanitizeForJS(varName),
				X:     xRel,
				Y:     data.Y,
				Color: colors[colorIdx%len(colors)],
			})
			colorIdx++
		}
	}
	return traces
}

const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <script src="{{.PlotlyJS}}"></script>
    <title>{{.Title}}</title>
</head>
<body>
    <div id="plotly-div" style="width:1200px;height:600px;"></div>
    <script>
        var data = [
{{- range $i, $t := .Traces }}
            {
                x: [{{range $j, $v := $t.X}}{{if $j}},{{end}}{{$v}}{{end}}],
                y: [{{range $j, $v := $t.Y}}{{if $j}},{{end}}{{$v}}{{end}}],
                mode: 'lines',
                name: '{{$t.Name}}',
                line: {width: 2, color: '{{$t.Color}}'}
            }{{if not (isLast $i $.Traces)}},{{end}}
{{- end }}
        ];
        var layout = {
            title: '{{.Title}}',
            xaxis: {title: 'Time (s)'},
            yaxis: {title: 'Value'},
            hovermode: 'x unified',
            hoverlabel: {namelength: -1},
            plot_bgcolor: 'rgba(240,240,240,0.5)',
            font: {size: 11}
        };
        Plotly.newPlot('plotly-div', data, layout);
    </script>
</body>
</html>
`

// GenerateHTML creates an interactive Plotly HTML file.
func GenerateHTML(allData model.ParsedData, outputPath string, title string) error {
	traces := buildTraces(allData)

	funcMap := template.FuncMap{
		"isLast": func(i int, list []Trace) bool {
			return i == len(list)-1
		},
	}

	tmpl, err := template.New("stcm").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse html template: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create html file: %w", err)
	}
	defer file.Close()

	data := HTMLData{
		Title:    title,
		PlotlyJS: template.URL(defaultPlotlyURL),
		Traces:   traces,
	}

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute html template: %w", err)
	}

	return nil
}

// ResolveOutputName determines the base output name from the STCM file name.
func ResolveOutputName(stcmFileName string, explicitName string) string {
	if explicitName != "" {
		return explicitName
	}

	re := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}_\d{2}h\d{2}m\d{2}s)`)
	matches := re.FindStringSubmatch(stcmFileName)
	if len(matches) > 1 {
		return matches[1]
	}

	// Fallback: use file name without extension so renamed files keep their name.
	// e.g. "my_log.stcm" -> "my_log", "test.stcm" -> "test"
	trimmed := strings.TrimSuffix(stcmFileName, filepath.Ext(stcmFileName))
	trimmed = strings.TrimSpace(trimmed)
	if trimmed == "" {
		return "stcm_viewer_interactive"
	}
	return trimmed
}
