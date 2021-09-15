package spyre

import (
	"github.com/spf13/afero"
)

var Version = "1.2.0"

var FS = afero.NewOsFs()
