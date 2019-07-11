package agent

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/config"
)

type agentMetaData struct {
	/*
		to allow us to use comparison operators without relying on magic, the version string needs to be
		seperated, and numeric in nature. However, to preserve information, we keep the entire version string
	*/
	versionString string
	fields        map[string]interface{}
	tags          map[string]string
}

func runeInList(a rune) bool {
	// return true if a in found within our list of "ok runes"
	ok := []rune{'_', '-', '~', '+'}
	for _, r := range ok {
		if a == r {
			return true
		}

	}
	return false
}

func extractNumeric(digitString string, isLast bool) (int64, error) {
	/*
		basic helper that takes a string, parses it character by character, and expects to find only numerics, unless "isLast" is true
		in which case it will check that the next character is a specific one - otherwise error.

		so, these are ok:
		1
		1-
		1~

		but not these:
		-123
		v1
		1b

		once pafrsing completes, it will return the numerics, and any error that occurred
	*/
	var sb strings.Builder

	valid := true

	for i, digit := range digitString {
		if unicode.IsDigit(digit) {
			sb.WriteRune(digit)
		} else {
			valid = false
			if i > 0 {
				if isLast {
					valid = runeInList(digit)
					break
				}

			}
			break
		}

	}

	if !valid {
		return 0, errors.New("invalid string found within version string: " + digitString)
	}

	d, err := strconv.ParseInt(sb.String(), 0, 64)
	if err != nil {
		// found something that looks numeric, but failed to parse, which is definitely and error
		return 0, err
	}

	return d, nil
}

func getAsSemVer(versionString string) ([]int64, error) {
	/*
		this function verfies that the supplied string represents a semantic version string:

		MAJOR.MINOR.PATCH

		so, the following are considered legal:

		1.2.3
		1.2.3-rc1

		but not:

		1.2
		1.banana.2
		1

		once we have established that the format is correct, we need to extract the numerics. for example:

		v1.2.3 becomes -> 1.2.3
		1.2.3~123bvdbc1 -> 1.2.3
		1.1.9-alpha -> 1.1.9

		this gives us the 3 fields needed for sorting (the appended, optional pre-release version is useless for sorting)
	*/

	numerics := []int64{}
	// verify that the version string contains 2 "." characters as we expect to find
	dotsFound := strings.Count(versionString, ".")
	if dotsFound != 2 {
		return []int64{}, errors.New("version string is not of expected semantic format: " + versionString)
	}

	chunks := strings.Split(versionString, ".")

	for i, chunk := range chunks {
		d, err := extractNumeric(chunk, i == 2)
		if err != nil {
			return []int64{}, err
		}
		numerics = append(numerics, d)
	}
	return numerics, nil
}

func (a *agentMetaData) addMetaData(conf *config.Config) error {
	/*
		simple initial (hence currently useless error return) helper that allows all non-version meta to be added
		now add our basic stats here - can be anything else too high-level for inputs to have visibility of
	*/
	a.fields["number_inputs"] = len(conf.Inputs)
	a.fields["number_outputs"] = len(conf.Outputs)
	return nil
}

func (a *agentMetaData) addVersion(conf *config.Config) error {
	version := conf.Agent.Version
	// handle empty version string
	if version == "" {
		version = "none"
	}
	a.versionString = version
	a.fields["version_string"] = version

	// now that we have set the string version, we need to extract the numerics - and error if this fails
	numericVersionChunks, err := getAsSemVer(version)
	if err != nil {
		if !conf.Agent.IgnoreInvalidVersion {
			return err
		}
	} else {
		// if here, we know that we have the required 3 numeric segments
		a.fields["major_version"] = int64(numericVersionChunks[0])
		a.fields["minor_version"] = int64(numericVersionChunks[1])
		a.fields["patch_version"] = int64(numericVersionChunks[2])
	}
	return nil
}

func newagentMetaData(conf *config.Config) (*agentMetaData, error) {
	a := new(agentMetaData)
	a.fields = make(map[string]interface{})
	a.tags = conf.Tags
	// add basic metadata
	err := a.addMetaData(conf)
	if err != nil {
		return a, err
	}
	// now add the version
	err = a.addVersion(conf)
	return a, err
}

type agentMonitor struct {
	name     string
	config   *config.Config
	ctx      context.Context
	metaData *agentMetaData
	outgoing chan telegraf.Metric
	signals  chan os.Signal
	jitter   time.Duration
	interval time.Duration
}

/*
	next two functions are to fulfill the MetricMaker interface required by the Accumulator
*/
func (a *agentMonitor) Name() string {
	return a.name
}

func (a *agentMonitor) MakeMetric(metric telegraf.Metric) telegraf.Metric {
	return metric
}

// NewAgentMonitor returns a new AgentMonitor, and an error, if one occured during instantiation of agentMetaData (which handles the version stuff)
func NewAgentMonitor(ctx context.Context, config *config.Config, singals chan os.Signal, outgoing chan telegraf.Metric) (*agentMonitor, error) {
	a := new(agentMonitor)
	a.name = "agent_monitor"
	a.config = config
	a.ctx = ctx
	meta, err := newagentMetaData(config)
	if err != nil {
		return a, err
	}
	a.metaData = meta
	a.outgoing = outgoing
	a.signals = singals
	// the next allows for easier testing as decouples the agent from config so Run() is testable with custom values
	a.jitter = a.config.Agent.CollectionJitter.Duration
	a.interval = a.config.Agent.MetaCollectionInterval.Duration

	return a, nil

}

func (a *agentMonitor) shouldRunCollection() bool {
	// return whatever is set within the config
	return a.config.Agent.EnableMeta
}

func (a *agentMonitor) shouldRunSignals() bool {
	// return whatever is set within the config
	return a.config.Agent.EnableStateChange
}

// Run starts and runs the AgentMonitor IF it has been enabled, and until the context is done.
func (a *agentMonitor) Run() {
	/*
		the agent emits two types of metric:
		1. scheduled periodic measurements. things like agent version
		2. notifications. things such as state changes as and when they occur (in this case, signals received by the agent)

		possible that Run() is called when not needed - simply log and exit gracefully.
	*/

	if !a.shouldRunSignals() && !a.shouldRunCollection() {
		log.Printf("D! [agent monitor] disabled in config, exiting")
		return
	}
	acc := NewAccumulator(a, a.outgoing)
	acc.SetPrecision(time.Second)
	wg := new(sync.WaitGroup)
	if a.shouldRunSignals() {
		log.Printf("D! [agent monitor] starting signals")
		// we need to send a state change on startup - as there is no Signal for this
		fields := make(map[string]interface{})
		fields["state"] = "started"
		acc.AddFields("agent_statechange", fields, a.metaData.tags, time.Now())
		wg.Add(1)
		// now, start listening for signals from the parent as these take 1st priority
		go func() {
			defer wg.Done()
			for {
				select {
				case data := <-a.signals:
					signalReceived := data.String()
					fields := make(map[string]interface{})
					fields["state"] = signalReceived
					acc.AddFields("agent_statechange", fields, a.metaData.tags, time.Now())
				case <-a.ctx.Done():
					return
				}
			}
		}()
	}
	if a.shouldRunCollection() {
		log.Printf("D! [agent monitor] starting collector")
		// now initiate the periodic collector
		ticker := NewTicker(a.interval, a.jitter)
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer ticker.Stop()
			for {
				acc.AddFields("agent_meta_data", a.metaData.fields, a.metaData.tags, time.Now())
				select {
				case <-ticker.C:
					//nothing to do - just allow next iteration
				case <-a.ctx.Done():
					return
				}

			}
		}()
	}
	wg.Wait()
	log.Printf("D! [agent monitor] exiting")
}
