package benchmark

import (
	_ "embed"
	"encoding/json"
)

//go:embed testdata/http_traffic.json
var testCasesJSON []byte

//go:embed testdata/edge_cases.json
var edgeCasesJSON []byte

func GetTestCases() []TestCase {
	var httpTraffic []TestCase
	if err := json.Unmarshal(testCasesJSON, &httpTraffic); err != nil {
		panic("Failed to load http_traffic.json: " + err.Error())
	}

	var edgeCases []TestCase
	if err := json.Unmarshal(edgeCasesJSON, &edgeCases); err != nil {
		panic("Failed to load edge_cases.json: " + err.Error())
	}

	return append(httpTraffic, edgeCases...)
}
