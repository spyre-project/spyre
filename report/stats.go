package report

var Stats struct {
	File struct {
		ScanCount uint64
		SkipCount uint64
		// EvidenceCount uint64

		ScanBytes uint64
		// SkipBytes uint64
		// EvidenceBytes uint64

		Matches  uint64
		NoAccess uint64
	}
	Process struct {
		ScanCount uint64
		SkipCount uint64

		Matches uint64
	}
}
