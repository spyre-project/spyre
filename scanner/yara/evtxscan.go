package yara

import (
	yr "github.com/lprat/go-yara/v4"

	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/report"
	"github.com/spyre-project/spyre/scanner"
  "github.com/spyre-project/spyre/log"

  "fmt"
	"time"
	"encoding/json"
)

func init() { scanner.RegisterEvtxScanner(&evtxScanner{}) }

type evtxScanner struct{ rules *yr.Rules }

func (s *evtxScanner) Name() string { return "YARA-evtx" }

func (s *evtxScanner) Init() error {
	var err error
	s.rules, err = compile(evtxscan, config.YaraEvtxRules)
	return err
}

func (s *evtxScanner) ScanEvtx(evt string, jsonval []byte) error {
	var (
		matches yr.MatchRules
		err     error
	)
	err = s.rules.ScanMem([]byte(evt), 0, 1*time.Minute, &matches)
	for _, m := range matches {
    var data map[string]interface{}
		err = json.Unmarshal(jsonval, &data)
    if err != nil {
        log.Errorf("Error to read evtx json : %s", err)
    }
		event_id := "Unknown"
		source_name := "Unknown"
		event_date := "Unknown"
		event_level := "Unknown"
		event_sid := "Unknown"
		if val, ok := data["Event"]; ok {
       if evtmap, ok := val.(map[string]interface{}); ok {
				 if val2, ok := evtmap["System"]; ok {
					 if sysmap, ok := val2.(map[string]interface{}); ok {
						  if valx, ok := sysmap["EventID"]; ok {
                event_id = fmt.Sprintf("%s",valx)
					    }
							if valx, ok := sysmap["Channel"]; ok {
                source_name = fmt.Sprintf("%s",valx)
					    }
							if valx, ok := sysmap["Level"]; ok {
                event_level = fmt.Sprintf("%s",valx)
					    }
							if val3, ok := sysmap["TimeCreated"]; ok {
								if datemap, ok := val3.(map[string]interface{}); ok {
									if valx, ok := datemap["SystemTime"]; ok {
                    event_date = fmt.Sprintf("%s",valx)
								  }
							  }
					    }
							if val3, ok := sysmap["Security"]; ok {
								if secmap, ok := val3.(map[string]interface{}); ok {
									if valx, ok := secmap["UserID"]; ok {
                    event_sid = fmt.Sprintf("%s",valx)
								  }
							  }
					    }
					 }
         }
			 }
		}
		message := m.Rule + " (yara) matched on event windows: " + event_id + "(" + source_name + ")" + "[" + event_level + "]"
		report.AddEvtxInfo(evt, "yara_on_eventlog", message,
			"rule", m.Rule, "event_level", event_level, "event_identifier", event_id, "source_name", source_name, "real_date", event_date, "event_sid", event_sid)
	}
	return err
}
