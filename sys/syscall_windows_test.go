package sys

import (
	"testing"
)

func TestGetDriveLetters(t *testing.T) {
	drivestrings, err := GetLogicalDriveStrings()
	if err != nil {
		t.Error(err)
	}
	t.Logf("%+[1]v %#[1]v", drivestrings)

	for _, s := range drivestrings {
		typ, err := GetDriveType(s)
		if err != nil {
			t.Error(err)
			continue
		}
		t.Logf("- type(%s) = %d", s, typ)
	}
}
