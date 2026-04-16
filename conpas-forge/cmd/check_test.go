package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestCheckCommand_HasJSONFlag(t *testing.T) {
	flag := checkCmd.Flags().Lookup("json")
	if flag == nil {
		t.Fatal("check command missing --json flag")
	}
}

func TestCheckCommand_TableHeaders(t *testing.T) {
	checkJSONFlag = false
	buf := &bytes.Buffer{}
	checkCmd.SetOut(buf)
	defer checkCmd.SetOut(nil)

	_ = checkCmd.RunE(checkCmd, []string{})

	output := buf.String()
	if !strings.Contains(output, "MODULE") || !strings.Contains(output, "STATUS") {
		t.Errorf("table output missing expected headers, got:\n%q", output)
	}
}

func TestCheckCommand_JSONOutput(t *testing.T) {
	checkJSONFlag = true
	defer func() { checkJSONFlag = false }()

	buf := &bytes.Buffer{}
	checkCmd.SetOut(buf)
	defer checkCmd.SetOut(nil)

	_ = checkCmd.RunE(checkCmd, []string{})

	output := buf.String()
	if !strings.Contains(output, `"modules"`) {
		t.Errorf("JSON output missing top-level 'modules' key, got:\n%q", output)
	}
	if !strings.Contains(output, `"name"`) || !strings.Contains(output, `"status"`) {
		t.Errorf("JSON output missing expected fields, got:\n%q", output)
	}
}
