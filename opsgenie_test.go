package main

import (
	"testing"
)

func Test_countAlertsParams_ToQuery(t *testing.T) {
	testcases := []struct {
		Name   string
		Input  countAlertsParams
		Output string
	}{
		{
			Name:   "default params",
			Input:  countAlertsParams{},
			Output: "",
		},
		{
			Name: "status only",
			Input: countAlertsParams{
				Status: "open",
			},
			Output: "status: open",
		},
		{
			Name: "team only",
			Input: countAlertsParams{
				Team: "Everyone",
			},
			Output: "teams: Everyone",
		},
		{
			Name: "priority only",
			Input: countAlertsParams{
				Priority: "P2",
			},
			Output: "priority: P2",
		},
		{
			Name: "status + team",
			Input: countAlertsParams{
				Status:   "closed",
				Team:     "Everyone",
				Priority: "P3",
			},
			Output: "status: open AND teams: Everyone AND priority: P3",
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			if actual := testcase.Input.ToQuery(); actual != testcase.Output {
				t.Errorf("ToQuery() failed: %s != %s", actual, testcase.Output)
			}
		})
	}
}
