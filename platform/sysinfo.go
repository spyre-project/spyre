package platform

type SystemInformation []kv

type kv struct {
	Key   string
	Value string
}

func (si SystemInformation) String() (s string) {
	if len(si) == 0 {
		return "(n/a)"
	}
	for _, kv := range si {
		if len(s) > 0 {
			s += " "
		}
		s += kv.Key
		s += `="`
		s += kv.Value
		s += `"`
	}
	return
}
