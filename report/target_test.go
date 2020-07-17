package report

import (
	"reflect"
	"testing"
)

func TestTargetSpec(t *testing.T) {
	tests := []struct {
		spec     string
		expected target
	}{
		{"C:/Program Files/nobody/spyre.log",
			target{
				formatter: &formatterPlain{},
				writer:    &fileWriter{path: "C:/Program Files/nobody/spyre.log"},
			}},
	}
	for _, test := range tests {
		got, err := mkTarget(test.spec)
		if err != nil {
			t.Errorf("parse '%s': %v", test.spec, err)
		} else if !reflect.DeepEqual(got, test.expected) {
			t.Errorf("parse '%s': got %+v, expected %+v", test.spec, got, test.expected)
		}
	}
}
