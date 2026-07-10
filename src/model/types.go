package model

// Point represents a single (x, y) data point from an STCM file.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// STCMEntry represents one entry in the STCM JSON array.
type STCMEntry struct {
	GroupName    string  `json:"groupname"`
	VariableName string  `json:"variablename"`
	VariableData []Point `json:"variabledata"`
}

// VariableData holds the time/value series for one variable.
type VariableData struct {
	X []float64
	Y []float64
}

// ParsedData maps group -> variable -> data series.
type ParsedData map[string]map[string]*VariableData
