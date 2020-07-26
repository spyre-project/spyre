package yara

import (
	"github.com/spf13/afero"

	yr "github.com/hillu/go-yara/v4"

	"testing"
)

func TestInclude(t *testing.T) {
	c, err := yr.NewCompiler()
	if err != nil {
		t.Fatalf("Could not create YARA compiler: %s", err)
	}
	is := &includeState{fs: afero.NewBasePathFs(afero.NewOsFs(), "testdata")}
	c.SetIncludeCallback(is.IncludeCallback)
	if err = c.AddString(`include "/a.yar"`, ""); err != nil {
		t.Errorf("Could not parse /a.yar -> b/b.yar -> ../c.yar: %s", err)
	}
}
