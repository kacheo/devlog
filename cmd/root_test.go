package cmd

import (
	"bytes"
	"testing"
)

func TestRootCmd_NoArgs_PrintsHelp(t *testing.T) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("devlog")) {
		t.Errorf("help output does not contain 'devlog', got:\n%s", buf.Bytes())
	}
}

func TestRootCmd_GlobalFlags_Registered(t *testing.T) {
	f := rootCmd.PersistentFlags()
	if f.Lookup("json") == nil {
		t.Error("--json flag not registered")
	}
	if f.Lookup("date") == nil {
		t.Error("--date flag not registered")
	}
}
