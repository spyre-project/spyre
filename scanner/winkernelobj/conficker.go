//go:build windows
// +build windows

package winkernelobj

import (
	"github.com/spyre-project/spyre/log"

	"golang.org/x/sys/windows"

	"github.com/mitchellh/go-ps"

	"fmt"
	"hash/crc32"
)

// MSVCRT compatible Random number generator, per
// https://stackoverflow.com/a/1280765
type rng struct {
	seed uint32
}

func NewRNG(seed uint32) rng {
	return rng{seed}
}

func (r *rng) rand() uint32 {
	r.seed = r.seed*0x343fd + 0x269EC3
	return (r.seed >> 0x10) & 0x7FFF
}

func startup_mutex(pid uint32, use_seed bool) string {
	var buf [20]byte
	seed := pid
	if use_seed {
		seed ^= 0x630063
	}
	r := NewRNG(seed)
	end := r.rand()%7 + 10

	for i := range buf[:end] {
		buf[i] = uint8(r.rand()%26) + 'a'
	}
	return string(buf[:end])
}

func simplified_crc(s string) uint32 {
	var result uint32 = 0xffffffff
	bs := []byte(s)
	for _, c := range bs {
		for s := 8; s > 0; s -= 1 {
			if ((uint32(c) ^ result) & 0x01) != 0 {
				result = (result >> 1) ^ 0x0EDB88320
			} else {
				result = result >> 1
			}
			c >>= 1
		}
	}
	return result
}

func (s *systemScanner) addConfickerIOCs() {
	cn, _ := windows.ComputerName()
	log.Debugf("%s: Adding Conficker IOCs", s.ShortName())
	s.IOCs["Conficker.A global mutex"] = Obj{"Mutant", fmt.Sprintf("\\BaseNamedObjects\\%d-%d", crc32.ChecksumIEEE([]byte(cn)), 7)}
	s.IOCs["Conficker.B global mutex"] = Obj{"Mutant", fmt.Sprintf("\\BaseNamedObjects\\%d-%d", simplified_crc(cn), 7)}
	s.IOCs["Conficker.C global mutex"] = Obj{"Mutant", fmt.Sprintf("\\BaseNamedObjects\\%d-%d", simplified_crc(cn), 99)}
	if procs, err := ps.Processes(); err != nil {
		log.Errorf("%s: Error listing processes: %v", s.ShortName, err)
	} else {
		for _, proc := range procs {
			pid := uint32(proc.Pid())
			s.IOCs[fmt.Sprintf("Conficker.B pid=%d mutex", pid)] = Obj{"Mutant", "\\BaseNamedObjects\\" + startup_mutex(pid, false)}
			s.IOCs[fmt.Sprintf("Conficker.C pid=%d mutex", pid)] = Obj{"Mutant", "\\BaseNamedObjects\\" + startup_mutex(pid, true)}
		}
	}
}
