package yara

import (
	yr "github.com/hillu/go-yara/v4"
	"github.com/shirou/gopsutil/v3/process"

	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/report"
	"github.com/spyre-project/spyre/scanner"

  "crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
  "strconv"
	"strings"
	"time"

)

func init() { scanner.RegisterProcScanner(&procScanner{}) }

func hash_file_md5(filePath string) (string, error) {
	var returnMD5String string
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}
	hashInBytes := hash.Sum(nil)[:16]
	returnMD5String = hex.EncodeToString(hashInBytes)
	return returnMD5String, nil

}

type procScanner struct{ rules *yr.Rules }

func (s *procScanner) Name() string { return "YARA-proc" }

func (s *procScanner) Init() error {
	var err error
	s.rules, err = compile(procscan, config.YaraProcRules)
	return err
}

func (s *procScanner) ScanProc(proc int32) error {
	var matches yr.MatchRules
	handle, err := process.NewProcess(pid)
	if err != nil {
		return err
	}
	exe, err := handle.Name()
	if err != nil {
		exe = ""
	}
	if !(stringInSlice(exe, config.ProcIgnoreList)) {
		return fmt.Errorf("Skipping process (found on ignore list) %s[%d].",exe,pid)
	}
	ppidx, err := handle.Ppid()
	ppid := ""
	if err == nil {
		ppid = strconv.FormatInt(int64(ppidx), 10)
	}
	pcmdline := ""
	pexe := ""
	ppathexe := ""
	pusername := ""
	phandle, err := handle.Parent()
	if (err == nil)  {
		pcmdline, err = phandle.Cmdline()
		if err != nil {
			pcmdline = ""
		}
		pexe, err = phandle.Name()
		if err != nil {
			pexe = ""
		}
		ppathexe, err = phandle.Exe()
		if err != nil {
			ppathexe = ""
		}
		pusername, err = phandle.Username()
		if err != nil {
			pusername = ""
		}
	}
	cmdline, err := handle.Cmdline()
	if err != nil {
		cmdline = ""
	}
	pathexe, err := handle.Exe()
	if err != nil {
		pathexe = ""
	}
	username, err := handle.Username()
	if err != nil {
		username = ""
	}
	crt_time, err := handle.CreateTime()
	if err != nil {
		crt_time = 0
	}
	childrens, err := handle.Children()
	var child_cmdline []string
	var child_pathexe []string
	var child_username []string
	var child_exe []string
	if err == nil {
		for _, handlechild := range childrens {
			cmdline, err := handlechild.Cmdline()
			if err == nil {
				if stringInSlice(cmdline, child_cmdline) {
					child_cmdline = append(child_cmdline, cmdline)
				}
			}
			exe, err := handlechild.Name()
			if err == nil {
				if stringInSlice(exe, child_exe) {
					child_exe = append(child_exe, exe)
				}
			}
			pathexe, err := handlechild.Exe()
			if err == nil {
				if stringInSlice(pathexe, child_pathexe) {
					child_pathexe = append(child_pathexe, pathexe)
				}
			}
			username, err := handlechild.Username()
			if err == nil {
				if stringInSlice(username, child_username) {
					child_username = append(child_username, username)
				}
			}
		}
	}
	if ppid == strconv.FormatInt(int64(pid), 10) {
		ppid = ""
	} else if ppid == "0" {
		ppid = ""
	}
	for _, v := range []struct {
		name  string
		value interface{}
	}{
		{"pid", strconv.FormatInt(int64(pid), 10)},
		{"pathexe", pathexe},
		{"cmdline", cmdline},
		{"executable", exe},
		{"username", username},
		{"ppid", ppid},
		{"ppathexe", ppathexe},
		{"pcmdline", pcmdline},
		{"pexecutable", pexe},
		{"pusername", pusername},
		{"ccmdline", strings.Join(child_cmdline, "|")},
		{"cpathexe", strings.Join(child_pathexe, "|")},
		{"cusername", strings.Join(child_username, "|")},
		{"cexecutable", strings.Join(child_exe, "|")},
	} {
		if err := s.rules.DefineVariable(v.name, v.value); err != nil {
			return err
		}
	}
	err = s.rules.ScanProc(int(pid), yr.ScanFlagsProcessMemory, 4*time.Minute, &matches)
	for _, m := range matches {
		var matchx []string
		for _, ms := range m.Strings {
			if stringInSlice(ms.Name+"-->"+string(ms.Data), matchx) {
				matchx = append(matchx, ms.Name+"-->"+string(ms.Data))
			}
		}
		matched := strings.Join(matchx[:], " | ")
		message := m.Rule+" (yara) matched on process: "+exe+"["+pathexe+"]("+username+")"
		if strings.HasPrefix("m.Rule", "kill_") {
			err = handle.Kill()
			if err == nil {
				message = "Killed process by "+m.Rule+" (yara) matched on process: "+exe+"["+pathexe+"]("+username+")"
			} else {
				message = "Error to kill process by "+m.Rule+" (yara) matched on process: "+exe+"["+pathexe+"]("+username+")"
			}
		}
		md5sum, err := hash_file_md5(pathexe)
		if err != nil {
			md5sum = ""
		}
		report.AddProcInfo("yara_on_pid", message,
			"rule", m.Rule,
			"string_match", string(matched),
			"PID", strconv.FormatInt(int64(pid), 10),
			"PPID", ppid,
			"Filehash", md5sum,
			"pathexe", pathexe,
			"cmdline", cmdline,
			"Process", exe,
			"username", username,
			"real_date", strconv.FormatInt(int64(crt_time),10),
			"Parent_pathexe", ppathexe,
			"Parent_cmdline", pcmdline,
			"Parent_Process", pexe,
			"Parent_username", pusername,
			"Child_cmdline", strings.Join(child_cmdline, "|"),
			"Child_pathexe", strings.Join(child_pathexe, "|"),
			"Child_username", strings.Join(child_username, "|"),
			"Child_Process", strings.Join(child_exe, "|"),
			)
		}
		if err != nil {
			message := fmt.Sprintf("Error yara proc scan [%v] on process: %s[%s](%s)",err,exe,pathexe,username)
			md5sum, err := hash_file_md5(pathexe)
			if err != nil {
				md5sum = ""
			}
			report.AddProcInfo("yara_on_pid", message,
				"PID", strconv.FormatInt(int64(pid), 10),
				"PPID", ppid,
				"Filehash", md5sum,
				"pathexe", pathexe,
				"cmdline", cmdline,
				"Process", exe,
				"username", username,
				"real_date", strconv.FormatInt(int64(crt_time),10),
				"Parent_pathexe", ppathexe,
				"Parent_cmdline", pcmdline,
				"Parent_Process", pexe,
				"Parent_username", pusername,
				"Child_cmdline", strings.Join(child_cmdline, "|"),
				"Child_pathexe", strings.Join(child_pathexe, "|"),
				"Child_username", strings.Join(child_username, "|"),
				"Child_Process", strings.Join(child_exe, "|"),
				)
		}
	return err
}
