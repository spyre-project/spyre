package spyre

import (
	"github.com/spf13/afero"
)

var Version = "1.2.6~pre"

var FS = afero.NewOsFs()
