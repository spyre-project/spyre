package config

import (
	"errors"
	"fmt"
	"math"
)

type FileSize int64

var suffixes = []string{"", "k", "M", "G", "T", "P", "E", "Z", "Y"}

func (f *FileSize) Set(val string) error {
	if val == "none" {
		*f = 0
		return nil
	}
	var sz float64
	var suffix string
	if n, _ := fmt.Sscanf(val, "%f%s", &sz, &suffix); n < 1 {
		return errors.New("could not parse size")
	}
	for i := len(suffixes) - 1; i >= 0; i-- {
		if suffix == suffixes[i] || suffix == suffixes[i]+"B" {
			*f = FileSize(sz * float64(uint(1)<<uint(10*i)))
			return nil
		}
	}
	return errors.New("could not parse size")
}

func (f *FileSize) UnmarshalText(text []byte) error {
	return f.Set(string(text))
}

func (f *FileSize) String() string {
	if *f <= 0 {
		return "none"
	}
	sz := float64(*f)
	var suffix string
	for i := len(suffixes) - 1; i >= 0; i-- {
		m := math.Exp2(float64(10 * i))
		if sz >= m {
			sz /= m
			suffix = suffixes[i]
			break
		}
	}
	return fmt.Sprintf("%.1f%sB", sz, suffix)
}

func (f *FileSize) Type() string { return "" }
