package report

import (
	"flag"
	"testing"
)

func TestFlags(t *testing.T) {
	fs := flag.NewFlagSet("test-program", flag.ExitOnError)
	var targets = TargetList{}
	fs.Var(&targets, "target", "report target(s)")
	err := fs.Parse([]string{"-target", "/path/to/file1 /path/to/file2"})
	if err != nil {
		t.Errorf("failed to parse: %v", err)
	}
	t.Logf("%#v", targets)
}
