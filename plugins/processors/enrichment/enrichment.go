package enrichment

import (
    "encoding/json"
    "io/ioutil"
    "os"
    "bytes"
    "log"
    "time"

    "github.com/influxdata/telegraf"
    "github.com/influxdata/telegraf/plugins/processors"
)

var sampleConfig = `
  ## Enrich with external Tags from an external json file set by EnrichFilePath.
  ##
  ## Conditionnal enrichment based on source tags already added by input plugin
  ## There are 2 levels of filtering. Level1 Source Tag ---> Level2 Source Tag ---> Tags to add
  ## If one level of filtering (default) is used the plugin looks for the wellknown level2
  ## Tag "LEVEL1TAGS" in the json file.
  ## The json file as read periodically every RefreshPeriod minutes. (by default 60m)
  ## See README file for more info about the Json file structure.
  ##
  enrichfilepath = ""
  twolevels = false
  refreshperiod = 60
  ## Filtering input tags
  ## Tags set by input plugin used as filter conditions
  ## Level2TagKey is only required when TwoLevel is set to true
  level1tagkey = ""
  level2tagkey = ""
`

var enrich map[string] map[string] map[string] string

type Enrichment struct {
    EnrichFilePath string `toml:"enrichfilepath"`
    TwoLevels bool `toml:"twolevels"`
    RefreshPeriod int `toml:"refreshperiod"`
    Level1TagKey string `toml:"level1tagkey"`
    Level2TagKey string `toml:"level2tagkey"`

    initialized bool
    FileError bool
    LastUpdate time.Time
}

func(p * Enrichment) SampleConfig() string {
    return sampleConfig
}

func(p * Enrichment) Description() string {
    return "Enrich with external tags based on existing tags"
}

func addBrackets(s string) string {
    var buf bytes.Buffer
    buf.WriteString("\"")
    buf.WriteString(s)
    buf.WriteString("\"")
    result := buf.String()
    return result
}

func(p * Enrichment) Apply(metrics...telegraf.Metric)[] telegraf.Metric {
    currentTime := time.Now()
    delta := int(currentTime.Sub(p.LastUpdate).Minutes())
    if !p.initialized || delta >= p.RefreshPeriod {
        if p.RefreshPeriod <= 0 {
            p.RefreshPeriod = 60
        }

        // Open enrichment file
        jsonFile, err := os.Open(p.EnrichFilePath)
        if err != nil {
            log.Printf("E! [processors.enrichment] Error when opening enrichment file %s error is %v", p.EnrichFilePath, err)
            p.FileError = false
            p.initialized = false
        } else {
            logPrintf("Successfully Open the file %s", p.EnrichFilePath)
            defer jsonFile.Close()
                //reset DB
            enrich = make(map[string] map[string] map[string] string)

            byteValue, _ := ioutil.ReadAll(jsonFile)
            json.Unmarshal([] byte(byteValue), & enrich)
            p.FileError = false
            p.initialized = true
            p.LastUpdate = time.Now()
        }
    }
    if !p.FileError {
        for _, metric := range metrics {
            CurrentTags := metric.Tags()
            Level1Tag := ""
            Level2Tag := ""
            Level1Tag = CurrentTags[p.Level1TagKey]
            logPrintf("Current L1 Tags value %v", Level1Tag)
            if p.TwoLevels {
                Level2Tag = CurrentTags[p.Level2TagKey]
                logPrintf("Current L2 Tags Value %v", Level2Tag)
            }
            if (Level1Tag != "") {
                // first add the Level 1 tags if present
                for tagKey, tagVal := range enrich[Level1Tag]["LEVEL1TAGS"] {
                        if (tagVal != "") {
                            logPrintf("Add level 1 Tag %s with value %s added", tagKey, tagVal)
                            metric.AddTag(tagKey, addBrackets(tagVal))
                        } else {
                            metric.AddTag(tagKey, "\"\"")
                        }
                    }
                    // if twolevels is set add level 2 tags if present
                if p.TwoLevels {
                    for tagKey, tagVal := range enrich[Level1Tag][Level2Tag] {
                        if (tagVal != "") {
                            logPrintf("Add level 2 Tag %s with value %s added", tagKey, tagVal)
                            metric.AddTag(tagKey, addBrackets(tagVal))
                        } else {
                            metric.AddTag(tagKey, "\"\"")
                        }
                    }
                }
            }
        }
    }
    return metrics
}

func logPrintf(format string, v...interface {}) {
    log.Printf("D! [processors.enrichment] " + format, v...)
}

func init() {
    processors.Add("enrichment", func() telegraf.Processor {
        return &Enrichment {}
    })
}