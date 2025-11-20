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
	httpTraffic := GetHTTPTrafficTestCases()
	edgeCases := GetEdgeCaseTestCases()
	return append(httpTraffic, edgeCases...)
}

func GetHTTPTrafficTestCases() []TestCase {
	var httpTraffic []TestCase
	if err := json.Unmarshal(testCasesJSON, &httpTraffic); err != nil {
		panic("Failed to load http_traffic.json: " + err.Error())
	}
	return httpTraffic
}

func GetEdgeCaseTestCases() []TestCase {
	var edgeCases []TestCase
	if err := json.Unmarshal(edgeCasesJSON, &edgeCases); err != nil {
		panic("Failed to load edge_cases.json: " + err.Error())
	}
	return edgeCases
}
