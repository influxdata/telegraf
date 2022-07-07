//go:generate ../../../tools/readme_config_includer/generator
package classify

// The "classify" plugin is a StreamingProcessor rather than an simple Processor
// because the aggregation side of its functionality may add more metrics to the
// data flow than existed in the input to the plugin.

// This code makes only a nodding attempt to limit itself to an 80-character line
// width.  In an advanced technological age of ubiquitous Unicode and 4K screens,
// slavishly adhering to an ancient, Procrustean punch-card standard like that is
// inane.  And with tab widths set by your program viewer instead of corresponding
// to some fixed number of spaces, who can tell how long a line really is, anyway?

// GLOBAL TODO LIST:
//
// * The standard TOML parser included with Telegraf 1.x is too limited in its
//   functionality to support the complex data that the classify plugin requires.
//   We are therefore forced to hoist the config data for this plugin out of the
//   standard Telegraf config files and into a separate config file that we can
//   analyze ourselves with a different TOML parser.  The canonical Go-language
//   TOML parser appears to be:
//       https://github.com/BurntSushi/toml
//       https://pkg.go.dev/github.com/BurntSushi/toml
//       https://godocs.io/github.com/BurntSushi/toml
//   so that is what we are adopting.  All we leave behind in the standard
//   Telegraf configuration for this plugin is a single option, pointing to
//   the file that we will need to read and parse on our own:
//       classify_config_file = "/etc/telegraf/telegraf.d/classify.toml"
//   Hopefully, Telegraf 2.0 will upgrade to using BurntSushi/toml, at which
//   point we can deprecate use of the classify_config_file option and move all
//   the configuration data back into a standard Telegraf config file.
//
// * The BurntSushi/toml package that is bundled into Telegraf should be updated
//   to the current release (v1.1.0 as of this writing).  That would allow us to
//   produce more detailed error messages when TOML parsing fails.
//
// * Implement proper logging level control, so if the code wants to create a
//   debug-level log message but the user does not want to see that level of
//   detail, that log message will be suppressed.  For a sample implementation,
//   see:  https://pkg.go.dev/github.com/go-kit/log/level
//   This should be a standard feature of a logger provided by Telegraf 2.0.
//   https://pkg.go.dev/github.com/influxdata/telegraf/testutil#Logger does not
//   appear to provide that capability.  All it does is label the messages, with
//   no attempt to possibly filter them.  In the interim, we may need to take
//   extreme measures to hack our own code to provide equivalent functionality,
//   even if we do so clumsily, so detailed logging in development and testing
//   is possible, but logging in production is not excessively verbose.
//
//   For comparison, we see NO documented means to control the logging level
//   during unit testing.  For production deployment, we see here:
//       https://docs.influxdata.com/telegraf/v1.22/configuration/
//   that there are [agent] config options "debug" (which is not well explained)
//   and "quiet", but no other log levels described.  It's not at all clear what
//   the default logging level is set to.  And there ought to be some means to
//   control the log level so it is clearly understood whether or not warning
//   and info level messages appear.
//
// * There should be some documented means to invoke the Telegraf TOML parser
//   on a string containing TOML config data, rather than programmatically
//   setting individual plugin-object-member values.  This would help with
//   unit tests, both generally and specifically to validate that the TOML
//   parsing itself will work as expected by a plugin.  For details on the
//   TOML parser used by Telegraf, see:
//       https://github.com/influxdata/toml
//       https://pkg.go.dev/github.com/influxdata/toml
//
// * Documentation for telegraf.Accumulator needs to be improved so it is
//   clear that synchronization is already built into ALL implementations of
//   AddMetric() and its sibling routines (e.g., both agent/accumulator.go
//   and testutil/accumulator.go), and need not be dealt with explicitly by
//   application code.  That is critical information for developers to know,
//   because the documentation for StreamingProcessor says that you can have a
//   background goroutine ultimately do the final work of processing an input
//   data point, long after Add() itself has returned.  So there could be
//   contention between such a goroutine pushing data into the accumulator and
//   Telegraf code pulling data out of the accumulator.
//
// * The manner in which Telegraf controls its plugin shutdown sequencing
//   should be documented.  Presumably it is in order of data flow through the
//   chains of plugins, but it would be helpful to have that explicitly mentioned.
//
// * The time.Ticker.NewTicker() routine ought to support an optional second
//   argument to specify the initial ticker duration, after which the standard
//   ticker duration would kick in.  That would allow you to more readily
//   control the phase of the ticking with respect to actual wall-clock time.
//   We should find some way to propose this extension and get it accepted in
//   the standard upstream package.

import (
	_ "embed"
	"fmt"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

// A StreamingProcessor plugin implements PluginDescriber and so must
// meet its requirements, and itself requires the following methods:
//     Start(acc Accumulator) error
//     Add(metric Metric, acc Accumulator) error
//     Stop() error
//
// A PluginDescriber requires the following methods:
//     SampleConfig() string
//     Description() string
// Somewhat bizarrely, NO TELEGRAF PLUGIN SUPPLIES THE Description()
// FUNCTION!  I don't understand how that can be the case, but for
// the time being we follow along in that tradition.
//
// Other than those things, everything else is a supporting-cast member.

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

var newlineRegex = regexp.MustCompile(`\n`)

type Classify struct {
	// The place to find a value used to determine which group of
	// per-category regexes should be used for pattern matching.
	// These options are mutually exclusive.  If neither of these
	// options is provided, the default_regex_group option value
	// will be used instead as that determination.
	SelectorTag   string `toml:"selector_tag"`
	SelectorField string `toml:"selector_field"`

	// We wanted to allow selector_mapping to be supplied in the config file
	// in one of two possible formats.
	//
	//     ## Examples of Random Order mapping elements:
	//     selector_mapping.'fire\d{3}' = 'firewall'
	//     selector_mapping.'ora456'    = 'database'
	//
	//     ## Equivalent mapping elements in Listed Order format:
	//     selector_mapping = [
	//         { 'fire\d{3}' = 'firewall' },
	//         { 'ora456'    = 'database' },
	//     ]
	//
	// Those formats would have been supplied to the plugin as SelectorMapping
	// when the config file is parsed, as one of these three types:
	//
	//     a nil value          // no selector_mapping was supplied
	//     map[string]string    // Random Order format
	//     []map[string]string  // Listed Order format
	//
	// In the Random Order form, the one map contains as many keys as needed.
	// At run time, the order of evaluation would be unspecified.  Conversely,
	// in the Listed Order form, each map is only allowed to contain one key,
	// and the order of evaluation would be exactly as mappings are presented
	// in the config file.
	//
	// In pursuit of that flexibility, we thought we could use this declaration
	// for SelectorMapping:
	//     SelectorMapping interface{} `toml:"selector_mapping"`
	// And then we thought, the TOML parser would figure out on the basis of
	// the incoming config-file syntax whether the selector_mapping option was
	// being provided as a TOML table (i.e., a hash), or as an array of small
	// hashes.  And it would turn that interface{} into the corresponding kind
	// of Go object.
	//
	// In practice, the BurntSushi/toml parser is not able to process the
	// syntax in the config file in that manner, when we declare the target
	// variable as a simple interface{}.  So we have to choose between the
	// possible config-file formats.  Since we want to allow full control
	// over the order of evaluation when that can affect the result of the
	// mapping, we choose to support the Listed Order format and abandon
	// trying to support the Random Order format as well.  That's too bad,
	// because it forces the use of extra punctuation in the config file
	// even when it would not otherwise be necessary.  In any case, with
	// that restriction in mind, we declare SelectorMapping in accord with
	// just that one format.
	//
	SelectorMapping []map[string]string `toml:"selector_mapping"`

	// Like any Go string, this will default to an empty string if not defined,
	// which in this case is exactly what we want at the application level as
	// a default value for this option.
	DefaultRegexGroup string `toml:"default_regex_group"`

	// One or the other of these two options must be defined, so we have
	// some value to match the mapped_selector_regexes against.  These
	// options are mutually exclusive.
	MatchTag   string `toml:"match_tag"`
	MatchField string `toml:"match_field"`

	// The category to use if category-regex pattern matching did not succeed.
	DefaultCategory string `toml:"default_category"`

	// drop_categories may be supplied in the config file in one of
	// two possible formats.
	//
	//     drop_categories = 'some_category'
	//     drop_categories = [ 'cat', 'dog', 'fish' ]
	//
	// These formats will be supplied to the plugin as DropCategories by
	// Telegraf once it parses the config file, as one of these three types:
	//
	//     a nil value  // no drop_categories was supplied
	//     string       // the user supplied one string
	//     []string     // the user supplied several strings
	//
	DropCategories interface{} `toml:"drop_categories"`

	// One or the other of these two options must be defined, so we have
	// a well-defined place to record the classification result in each
	// output data point.  These options are mutually exclusive.
	ResultTag   string `toml:"result_tag"`
	ResultField string `toml:"result_field"`

	// This is a rather complex data structure.  See the documentation
	// for the levels of mapping involved and for variations in how
	// the regexes at the leaf-node part of this data structure may be
	// written in the config file.
	MappedSelectorRegexes map[string][]map[string]interface{} `toml:"mapped_selector_regexes"`

	// Everything you could ever want to specify about how aggregation
	// counters should be handled and reported.
	AggregationPeriod         string   `toml:"aggregation_period"`
	AggregationMeasurement    string   `toml:"aggregation_measurement"`
	AggregationDroppedField   string   `toml:"aggregation_dropped_field"`
	AggregationTotalField     string   `toml:"aggregation_total_field"`
	AggregationSummaryTag     string   `toml:"aggregation_summary_tag"`
	AggregationSummaryValue   string   `toml:"aggregation_summary_value"`
	AggregationSummaryFields  []string `toml:"aggregation_summary_fields"`
	AggregationGroupTag       string   `toml:"aggregation_group_tag"`
	AggregationGroupFields    []string `toml:"aggregation_group_fields"`
	AggregationSelectorTag    string   `toml:"aggregation_selector_tag"`
	AggregationSelectorFields []string `toml:"aggregation_selector_fields"`
	AggregationIncludesZeroes bool     `toml:"aggregation_includes_zeroes"`

	// The Accumulator is passed to this plugin by Telegraf.  We populate it
	// with metrics that this plugin either takes from its input, perhaps
	// modifies, and copies to its output (this Accumulator), or generates
	// internally on a periodic basis as aggregation metrics.  Either
	// way, we never get to see what happens with the Accumulator after
	// that.  Apparently, Telegraf implements some mechanism to pull stuff
	// out of the Accumulator, but we are never told about that in the
	// https://pkg.go.dev/github.com/influxdata/telegraf documentation,
	// where such an explantion ought to reside.
	acc telegraf.Accumulator

	// The following internal data structure normalizes the external
	// "Random Order" (if we still supported that) and "Listed Order"
	// forms of selector_mapping into one form that we can calculate
	// with regardless of how the user specified the configuration data.
	// It also carries the precomputed regexes, so we don't have to
	// compile those again for every incoming data point.  Elements in
	// this array will be matched in order against the match item value.
	//
	//                 regex         map_value (regex group name)
	selectorMap []map[*regexp.Regexp]string

	// A convenience hash, used to look up a given matched category
	// for each input data point rather than walking the entire array
	// of such categories.
	dropThisCategory map[string]bool

	// The MappedSelectorRegexes data structure is so complex that what
	// we really need is a multi-index container, similar to what the
	// C++ Boost library provides.  Lacking that in Go, it is easiest
	// to use an ancillary internal data structure, not used during the
	// matching process, as we incrementally build up the mappedRegexes
	// data structure with which we will access the MappedSelectorRegexes
	// config data while processing.  Then we can properly validate what
	// the user provides, detecting a duplicate category in the same group.
	//
	//                      group    category seen?
	groupCategoriesSeen map[string]map[string]bool

	// We need a well-controlled data structure holding the regexes to
	// use during matching, and this is it.
	//
	// Logically, we don't want the first array in the middle of this
	// data structure, but the second-level map here must be iterated
	// over in as-listed-in-the-config-file order.  Since we do not
	// have a multi-index container in Go, we have to insert the array
	// into the middle of this datatype, and limit each secondary map
	// to just one category per map.  With that construction, we end
	// up with a well-defined ordering of categories.
	//
	//                group  seq  category array_of_regexes
	mappedRegexes map[string][]map[string][]*regexp.Regexp

	// All distinct categories found in the MappedSelectorRegexes, put
	// together in a hash for quick lookup.  This does not include the
	// value of DefaultCategory if that is a distinct category.
	allRegexCategories map[string]bool

	// How often the aggregation statistics should be emitted, translated
	// into an internal form that we can use for clocking that activity.
	aggregationTimePeriod time.Duration

	// Some of these fields are shared betwixt threads, but only after
	// their values have been set permanently, and that happens before
	// the extra threads have even been spawned.  So we don't need any
	// mutex protection for them.
	doAggregation         bool
	doSummaryAggregation  bool
	doGroupAggregation    bool
	doSelectorAggregation bool

	// The guarded fields here are shared dynamically between the main
	// thread, which increments counters, and the aggregation thread,
	// which reports and resets those counters.  So in this case we need
	// mutex protection.
	sharedData struct {
		// A guard to synchronize classification and aggregation threads
		// so they do not concurrently access shared data.
		aggregationMutex sync.Mutex

		// Where to collect summary-level counts.
		//                    category
		aggregationSummary map[string]int

		// Where to collect regex-group-level counts.
		//                     group     category
		aggregationByGroup map[string]map[string]int

		// Where to collect selector-level counts.
		//                       selector   category
		aggregationBySelector map[string]map[string]int
	}

	// A synchronization point provided so shutdown of the plugin
	// can be managed cleanly, by waiting for the aggregation
	// thread to exit before the plugin as a whole exits.
	syncWaitGroup sync.WaitGroup

	// A channel used to signal the aggregation thread that it should stop.
	stopRequested chan bool

	// A copy of the Log member of the parent wrapper, for reference
	// when all we have in hand is the Classify data structure and we
	// can't necessarily cast that to its parent ClassifyWrapper to
	// find its copy of this field.
	logger telegraf.Logger
}

type ClassifyWrapper struct {
	// The path to the .toml file that contains the detailed configuration
	// data for this plugin, so the full expressiveness of TOML v1.0.0 is
	// available.  Once Telegraf itself can parse full TOML v1.0.0, this
	// plugin should evolve to get rid of this wrapper and deprecate use
	// of the classify_config_file option.
	ClassifyConfigFile string `toml:"classify_config_file"`

	// Where all the detailed config data lives, in memory.
	Classify

	// The Log element is placed at the wrapper level so Telegraf can
	// understand that it exists, and initialize it according to its
	// usual conventions.  But since all of our routines
	//
	// This config option is not defined in the TOML; presumably, Telegraf
	// itself always defines this in an operational context, because we
	// get SIGSEGV upon a cl.Log.XXX() call if this is not defined in a
	// unit-test context, and other plugins (such as reverse_dns) contain
	// this element via the same construction but do not check for a nil
	// value before using this option to make a call.
	//
	// In terms of where log data ends up when this plugin is operating in
	// normal (not test) context, we see here:
	//     https://docs.influxdata.com/telegraf/v1.22/configuration/
	// that there is some [agent] setup that controls logging.  And we are
	// told separately that:
	//
	//     If your plugin struct has a Log telegraf.Logger then Telegraf
	//     will init it.  The output of it is defined in the [agent] config
	//     logtarget (either stdout or file or eventlog for Windows).
	//
	Log telegraf.Logger `toml:"-"`
}

// Reset() may be called to destroy all derivative (internal) state in the
// Classify object, leaving only the settings that were externally defined.
// This can be helpful when running multiple tests where you want to re-use
// the same object but with tweaks in between tests.
func (cl *Classify) Reset() error {
	cl.acc = nil
	cl.selectorMap = nil
	cl.dropThisCategory = nil
	cl.groupCategoriesSeen = nil
	cl.mappedRegexes = nil
	cl.allRegexCategories = nil

	cl.aggregationTimePeriod = 0

	cl.doAggregation = false
	cl.doSummaryAggregation = false
	cl.doGroupAggregation = false
	cl.doSelectorAggregation = false

	// The zero value for a sync.Mutex is an unlocked mutex.
	// Given that we have declared cl.sharedData.aggregationMutex to
	// be a fixed copy of such an object instead of a nillable pointer
	// to such an object, we cannot reset this variable directly.  So
	// we use indirection here to establish that we have the desired
	// state in hand when we finish resetting.
	if !cl.sharedData.aggregationMutex.TryLock() {
		// It should be impossible to get here.  If we do, some part of
		// the code has left the mutex locked, so we could not obtain
		// the lock.  If and when companion code later tries to unlock
		// the mutex, that will cause a run-time error.  Hence we raise
		// an error immediately instead of unlocking the mutex, so the
		// root cause of the potential later problem is identified.
		return fmt.Errorf("the cl.sharedData.aggregationMutex was found locked")
	}
	cl.sharedData.aggregationMutex.Unlock()

	cl.sharedData.aggregationSummary = nil
	cl.sharedData.aggregationByGroup = nil
	cl.sharedData.aggregationBySelector = nil

	// The zero value for a sync.WaitGroup is obviously such an object
	// with a current wait count of zero.  However, the sync package
	// does not provide any means to know the present state of the
	// wait count, other than waiting for it to become zero.  There is,
	// for instance, no TimedWait() call that would return if the wait
	// count does not become zero within the specified timeout (which
	// would be zero if you just wanted to check and return immediately).
	// Regardless of whether or not the timeout expired, the return value
	// of TimedWait() would be the present value of the wait count at the
	// time that routine returned, perhaps along with the time remaining
	// in the timeout period when the routine returned (which would be
	// zero if the timeout expired).  (Of course, one might consider that
	// count to be somewhat volatile, since other threads might alter it
	// before you try to use it.  One might want to warn against TOCTOU
	// errors in the documentation.)  Perhaps adding a TimedWait() routine
	// is something that we ought to propose be added to the sync package.
	//
	// In the alternative, for purposes of just wanting to check the
	// current value of the wait count, one might propose that the Add()
	// routine be equipped to return the new wait count after the delta
	// has been added, and that it be permissible to add a zero delta.
	//
	// With no clear way to reset cl.syncWaitGroup (except perhaps to define
	// it as a nillable pointer to a sync.WaitGroup instead of a fixed
	// copy of such an object), we must leave it untouched in this Reset()
	// action, and depend on the rest of this package to never leave it in
	// an undesirable state.

	cl.stopRequested = nil

	return nil
}

// SampleConfig() is called by Telegraf when this plugin is requested to
// supply the outline of a useful configuration, to be locally modified.
func (*Classify) SampleConfig() string {
	return sampleConfig
}

// ParseDetailedConfig() handles full TOML v1.0.0 parsing of the classify.toml
// file that is specified by the classify_config_file plugin option known to
// Telegraf.
func (clw *ClassifyWrapper) ParseDetailedConfig() error {
	if clw.ClassifyConfigFile == "" {
		return fmt.Errorf("the classify_config_file option is not defined")
	}

	// We use BurntSushi/toml for this parsing, since it supports the full
	// range of TOML v1.0.0, it seems to be the canonical package for that,
	// and it is dead-simple to use.
	//
	// As long as we are handling parsing ourselves, we take some trouble to
	// make the error message be as descriptive as possible. to simplify the
	// process of debugging problems in the user-supplied classify.toml file.
	//
	switch md, err := toml.DecodeFile(clw.ClassifyConfigFile, &clw.Classify); err {
	default:
		if parseError, ok := err.(toml.ParseError); ok {
			// As of this writing, Telegraf seems to be bundling in only the
			// BurntSushi/toml v0.4.1 release, which does not yet include
			// support for the ErrorWithPosition() and ErrorWithUsage()
			// routines.  That comes in with the v1.0.0 release.  So for the
			// time being, we make do with what is readily available.  This
			// call should be switched to ErrorWithUsage() as soon as possible.
			return fmt.Errorf("parsing of the %q file failed:\n%s",
				clw.ClassifyConfigFile, parseError.Error())
		}
		return fmt.Errorf("parsing of the %q file failed:\n%v", clw.ClassifyConfigFile, err)
	case nil:
		// It's possible that the TOML decoding "succeeded" but some of the
		// items in the config file never made it into our structure.  That
		// could be evidence of a typo or other mistake in the config file,
		// so we check that here and flag the issue to the user.  We would
		// like to catch problems of that sort early on, independent of the
		// individual checks later on for items that were in fact decoded.
		extraKeys := ""
		undecodedKeys := md.Undecoded()
		for _, key := range undecodedKeys {
			extraKeys += fmt.Sprintf("\nfound undecoded key %q", key.String())
		}
		if extraKeys != "" {
			return fmt.Errorf("parsing of the %q file failed:%s",
				clw.ClassifyConfigFile, extraKeys)
		}
	}

	// Copy the logger to where we can access it when we have only a Classify
	// object in hand, not the wrapper.
	clw.Classify.logger = clw.Log

	return nil
}

// Start() is called once when the plugin starts; it is only called once per
// plugin instance, and never in parallel.
//
// Start should return once it is ready to receive metrics.
//
// The passed in accumulator is the same as the one passed to Add(), so you
// can choose to save it in the plugin, or use the one received from Add().
//
func (clw *ClassifyWrapper) Start(acc telegraf.Accumulator) error {
	var err error = nil //nolint:revive // sometimes, explicit initialization is clearer
	if err == nil {
		err = clw.ParseDetailedConfig()
	}
	if err == nil {
		err = clw.Classify.Start(acc)
	}
	return err
}

// The internal form of Start(), to be called once the Classify
// structure is filled in.  This call is used directly in unit
// testing, wherein we do not parse an external config file to
// obtain all the detailed config data.
func (cl *Classify) Start(acc telegraf.Accumulator) error {
	var err error = nil //nolint:revive // sometimes, explicit initialization is clearer
	if err == nil {
		err = cl.InitClassification(acc)
	}
	if err == nil {
		err = cl.InitAggregation()
	}
	if err == nil {
		err = cl.InitSynchronization()
	}
	if err == nil {
		err = cl.StartAggregation()
	}
	if err == nil {
		err = cl.StartClassification()
	}
	return err
}

// processor.go says in part:
//
//     Add is called for each metric to be processed. The Add() function does not
//     need to wait for the metric to be processed before returning, and it may
//     be acceptable to let background goroutine(s) handle the processing if you
//     have slow processing you need to do in parallel.
//
//     Metrics you don't want to pass downstream should have metric.Drop() called,
//     rather than simply omitting the acc.AddMetric() call.
//
// However, NO JUSTIFICATION IS GIVEN FOR CALLING EITHER metric.Drop() OR
// acc.addMetric().  In particular, it is not explained that unless you
// add the metric to the accumulator, it won't be seen by downstream
// plugins, and that if you want to drop the metric, you MUST NOT add it
// to the accumulator.  Nor is any explanation at all given for what a
// call to metric.Drop() is supposed to accomplish.
//
func (cl *Classify) Add(metric telegraf.Metric, _ telegraf.Accumulator) error {
	dropPoint := false

	selectorItemValue := ""
	regexGroup := ""
	haveRegexGroup := false
	switch {
	case cl.SelectorTag != "":
		if value, ok := metric.GetTag(cl.SelectorTag); ok {
			selectorItemValue = value
		} else {
			if cl.logger != nil {
				cl.logger.Infof("dropping point (selector tag %q is missing)", cl.SelectorTag)
			}
			dropPoint = true
		}
	case cl.SelectorField != "":
		if value, ok := metric.GetField(cl.SelectorField); ok {
			if v, ok := value.(string); ok {
				selectorItemValue = v
			} else {
				// We don't have any conversions from other data types to strings in place,
				// nor any plans to put conversions in place.
				if cl.logger != nil {
					cl.logger.Infof("dropping point (selector field %q is not a string)", cl.SelectorField)
				}
				dropPoint = true
			}
		} else {
			if cl.logger != nil {
				cl.logger.Infof("dropping point (selector field %q is missing)", cl.SelectorField)
			}
			dropPoint = true
		}
	default:
		// This is a legitimate case.  Neither cl.SelectorTag nor cl.SelectorField was
		// defined in the configuration, so we default to using the cl.DefaultRegexGroup
		// value as the result of the selector mapping.  But we must also make sure that
		// the rest of the logic for this data point correctly skips trying to do anything
		// more with the selector mapping, even if cl.DefaultRegexGroup is an empty string
		// and it doesn't look like assignment to regexGroup took place here, and that
		// all later validation of the chosen regexGroup still happens.
		regexGroup = cl.DefaultRegexGroup
		haveRegexGroup = true
	}

	if !dropPoint && !haveRegexGroup {
		// Iterate over the selector mapping and attempt to map the selector to a regex group name.
		var matchedRegex *regexp.Regexp = nil //nolint:revive // sometimes, explicit initialization is clearer
		for _, mapping := range cl.selectorMap {
			for regex, group := range mapping {
				if regex.MatchString(selectorItemValue) {
					matchedRegex = regex
					if group == "*" {
						regexGroup = selectorItemValue
					} else {
						regexGroup = group
					}
					break
				}
			}
		}
		if matchedRegex == nil {
			// The selector did not match any mapping regex.  In this case, we use cl.DefaultRegexGroup
			// as the desired regex group, though that might itself be an empty string or not be the
			// name of any configured regex group.  We will check those conditions just below.
			if cl.logger != nil {
				cl.logger.Infof("selector item value %q does not match anything in the selector_mapping", selectorItemValue)
				// cl.logger.Debugf("selectorMap = %v", cl.selectorMap);
			}
			regexGroup = cl.DefaultRegexGroup
		}
	}

	if !dropPoint {
		if regexGroup == "" {
			// The selector effectively maps to an empty string, however that was calculated.
			// There is no further recourse in thie case; count this as a dropped data point.
			if cl.logger != nil {
				cl.logger.Infof("dropping point (selector item value %q effectively maps to an empty string)", selectorItemValue)
			}
			dropPoint = true
		} else if _, ok := cl.mappedRegexes[regexGroup]; !ok {
			// The selector effectively mapped to a non-empty string, but that is not the name of any configured regex group.
			if cl.DefaultRegexGroup == "" {
				if cl.logger != nil {
					cl.logger.Infof("dropping point (selector item value %q maps to %q, which does not match any mapped_selector_regexes group)",
						selectorItemValue, regexGroup)
				}
				dropPoint = true
			} else if _, ok := cl.mappedRegexes[cl.DefaultRegexGroup]; !ok {
				if cl.logger != nil {
					cl.logger.Infof("dropping point (selector item value %q maps to %q, and neither that nor the default regex group %q matchs any mapped_selector_regexes group)",
						selectorItemValue, regexGroup, cl.DefaultRegexGroup)
				}
				dropPoint = true
			} else {
				regexGroup = cl.DefaultRegexGroup
			}
		}
	}

	matchItemValue := ""
	if !dropPoint {
		switch {
		case cl.MatchTag != "":
			if value, ok := metric.GetTag(cl.MatchTag); ok {
				matchItemValue = value
			} else {
				if cl.logger != nil {
					cl.logger.Infof("dropping point (match tag %q is missing)", cl.MatchTag)
				}
				dropPoint = true
			}
		case cl.MatchField != "":
			if value, ok := metric.GetField(cl.MatchField); ok {
				if v, ok := value.(string); ok {
					matchItemValue = v
				} else {
					// We don't have any conversions from other data types to strings in place,
					// nor any plans to put conversions in place.
					if cl.logger != nil {
						cl.logger.Infof("dropping point (match field %q is not a string)", cl.MatchField)
					}
					dropPoint = true
				}
			} else {
				if cl.logger != nil {
					cl.logger.Infof("dropping point (match field %q is missing)", cl.MatchField)
				}
				dropPoint = true
			}
		default:
			// PROGRAMMING ERROR:  THIS SHOULD NEVER HAPPEN.
			// Neither cl.MatchTag nor cl.MatchField was defined in the configuration,
			// but this should have been caught during initialization and processing
			// should never have been enabled.
			if cl.logger != nil {
				cl.logger.Error("dropping point (internal programming error when both match_tag and match_field are missing)")
			}
			dropPoint = true //nolint:ineffassign // dead assignment, but here for consistency when logging "dropping point"
			// We are not told what will happen if Add() returns an error,
			// but if ever there is a time to do that, it is now.
			return fmt.Errorf("internal programming error when both match_tag and match_field are missing")
		}
	}

	// Note that we do not do any checking to see if matchItemValue is an empty string.
	// There is no special-case handling in that situation; it is up to the user's regexes
	// to match or not match that value as the user sees fit.
	matchedCategory := ""
	if !dropPoint {
		// Iterate through all the categories in the chosen regex group.
		if cl.logger != nil {
			cl.logger.Debug("attempting category regex matches")
		}
		// These loops for regex matching are complicated to decipher.  The intent is to perform
		// this iteration through the group's categories in order as defined in the config file.
		//
		//                   group  seq  category array_of_regexes
		// mappedRegexes map[string][]map[string][]*regexp.Regexp
		//
		switch categoryList, ok := cl.mappedRegexes[regexGroup]; ok {
		default:
			// There is no such regex group; count this as a dropped data point.
			// PROGRAMMING ERROR:  THIS SHOULD NEVER HAPPEN.
			// Earlier checks in processing this data point should have ruled out
			// our reaching this branch.
			if cl.logger != nil {
				cl.logger.Errorf("dropping point (internal programming error:  selector item mapped value %q does not match any mapped_selector_regexes group)", regexGroup)
			}
			dropPoint = true //nolint:ineffassign // dead assignment, but here for consistency when logging "dropping point"
			// We are not told what will happen if Add() returns an error,
			// but if ever there is a time to do that, it is now.
			return fmt.Errorf("internal programming error:  selector item mapped value %q does not match any mapped_selector_regexes group", regexGroup)
		case true:
			if cl.logger != nil {
				cl.logger.Debugf("selector item value %q mapped to regex group %q, which has %d categories",
					selectorItemValue, regexGroup, len(categoryList))
			}
		regex_match_loop:
			// Each categoryDefinition only contains one category => array_of_regexes mapping,
			// and each such category buried within the regexGroup is only allowed to belong
			// to one categoryDefinition within that regexGroup.
			for _, categoryDefinition := range categoryList {
				for category, categoryRegexes := range categoryDefinition {
					if cl.logger != nil {
						cl.logger.Debugf("matching category %q, which has %d regexes", category, len(categoryRegexes))
					}
					// iterate through all the regexes in each category
					for _, regex := range categoryRegexes {
						if cl.logger != nil {
							cl.logger.Debugf("matching against regex %q", regex)
						}
						if regex.MatchString(matchItemValue) {
							if cl.logger != nil {
								cl.logger.Debug("found match")
							}
							// upon finding a match of the matchItemValue to a regex, set the result item
							// into the output data point and exit all match-search looping
							matchedCategory = category
							break regex_match_loop
						}
					}
				}
			}
		}
	}
	if !dropPoint {
		if matchedCategory == "" {
			if cl.logger != nil {
				cl.logger.Debugf("no match was found; default category %q is to be applied", cl.DefaultCategory)
			}
			// No match was found.  Apply the default category.
			matchedCategory = cl.DefaultCategory
		} else {
			if cl.logger != nil {
				cl.logger.Debugf("matched category %q", matchedCategory)
			}
		}
		if matchedCategory == "" {
			// Applying the default category didn't help us determine a valid (non-empty-string)
			// result item value.  The only sensible thing we can do is to drop this data point.
			if cl.logger != nil {
				cl.logger.Debug("dropping point (no match category was found, and default_category is not defined)")
			}
			dropPoint = true
		} else if cl.dropThisCategory[matchedCategory] {
			if cl.logger != nil {
				cl.logger.Debugf("dropping point (category %q is configured in drop_categories)", matchedCategory)
			}
			// This case is special.  We have gone all the way to the point where we found a match
			// category (or used the default category), but then we decided to drop the data point
			// anyway.  In this case, we will want to count this data point in the aggregation
			// statistics both as an effectively matched point (in that respective category) and
			// as a dropped point as well.
			dropPoint = true
		}
	}
	if !dropPoint {
		if cl.ResultTag != "" {
			if cl.logger != nil {
				cl.logger.Debugf("setting result tag %q to %q", cl.ResultTag, matchedCategory)
			}
			metric.AddTag(cl.ResultTag, matchedCategory)
		} else if cl.ResultField != "" {
			if cl.logger != nil {
				cl.logger.Debugf("setting result field %q to %q", cl.ResultField, matchedCategory)
			}
			metric.AddField(cl.ResultField, matchedCategory)
		}
		// Add this data point to the accumulator so it will be seen by downstream plugins.
		cl.acc.AddMetric(metric)
	} else {
		// We are told that the proper action here is to call this function,
		// even though there is no description of what it will do.
		metric.Drop()
	}

	if cl.doAggregation {
		// I would prefer to bracket the code in this block with the Lock/Unlock actions,
		// as that would be clearer and there are no early exits from this block that might
		// bypass the Unlock.  But this block is at the end of the function anyway, and
		// deferring the Unlock makes the code immune to future changes that might somehow
		// invoke an early exit before the Unlock takes place.  The alternative is to move
		// the code in this block into a separate subroutine and do the Lock/Unlock actions
		// there, but it seems like a pointless exercise to invoke the extra overhead of
		// such a subroutine call.
		//
		cl.sharedData.aggregationMutex.Lock()
		defer cl.sharedData.aggregationMutex.Unlock()

		if cl.doSummaryAggregation {
			//                       category
			// aggregationSummary map[string]int
			//
			if matchedCategory != "" {
				cl.sharedData.aggregationSummary[matchedCategory]++
			}
			if dropPoint && cl.AggregationDroppedField != "" {
				cl.sharedData.aggregationSummary[cl.AggregationDroppedField]++
			}
			if cl.AggregationTotalField != "" {
				cl.sharedData.aggregationSummary[cl.AggregationTotalField]++
			}
		}

		// For regex-group-level and selector-level statistics, we don't create any counters
		// in advance, because we would prefer that they be dynamically allocated only
		// for those aggregation species (regex groups and selectors, respectively) that
		// actually show up as a consequence of processing input data points.  That dynamic
		// allocation saves us from extra work in walking the aggregation counters to see
		// which sets contain some non-zero values, when outputting the values of those
		// counters.  But it does mean that we must do some allocation here, when counting.

		if cl.doGroupAggregation {
			//                        group     category
			// aggregationByGroup map[string]map[string]int
			//
			if regexGroup != "" {
				if matchedCategory != "" {
					if cl.sharedData.aggregationByGroup[regexGroup] == nil {
						cl.sharedData.aggregationByGroup[regexGroup] = make(map[string]int)
					}
					cl.sharedData.aggregationByGroup[regexGroup][matchedCategory]++
				}
				if dropPoint && cl.AggregationDroppedField != "" {
					if cl.sharedData.aggregationByGroup[regexGroup] == nil {
						cl.sharedData.aggregationByGroup[regexGroup] = make(map[string]int)
					}
					cl.sharedData.aggregationByGroup[regexGroup][cl.AggregationDroppedField]++
				}
				if cl.AggregationTotalField != "" {
					if cl.sharedData.aggregationByGroup[regexGroup] == nil {
						cl.sharedData.aggregationByGroup[regexGroup] = make(map[string]int)
					}
					cl.sharedData.aggregationByGroup[regexGroup][cl.AggregationTotalField]++
				}
			}
		}

		if cl.doSelectorAggregation {
			//                          selector   category
			// aggregationBySelector map[string]map[string]int
			//
			if selectorItemValue != "" {
				if matchedCategory != "" {
					if cl.sharedData.aggregationBySelector[selectorItemValue] == nil {
						cl.sharedData.aggregationBySelector[selectorItemValue] = make(map[string]int)
					}
					cl.sharedData.aggregationBySelector[selectorItemValue][matchedCategory]++
				}
				if dropPoint && cl.AggregationDroppedField != "" {
					if cl.sharedData.aggregationBySelector[selectorItemValue] == nil {
						cl.sharedData.aggregationBySelector[selectorItemValue] = make(map[string]int)
					}
					cl.sharedData.aggregationBySelector[selectorItemValue][cl.AggregationDroppedField]++
				}
				if cl.AggregationTotalField != "" {
					if cl.sharedData.aggregationBySelector[selectorItemValue] == nil {
						cl.sharedData.aggregationBySelector[selectorItemValue] = make(map[string]int)
					}
					cl.sharedData.aggregationBySelector[selectorItemValue][cl.AggregationTotalField]++
				}
			}
		}
	}

	return nil
}

// Stop() gives you an opportunity to gracefully shut down the processor.
//
// Once Stop() is called, Add() will not be called any more. If you are using
// goroutines, you should wait for any in-progress metrics to be processed
// before returning from Stop().
//
// When Stop returns, you should no longer be writing metrics to the
// accumulator.
//
func (cl *Classify) Stop() error {
	// We must stop the classifier before we stop the aggregator, so the classifier
	// does not keep counting data that will never be output by the aggregator.
	//
	// We stop both sides regardless, but prefer reporting a problem with stopping
	// the classifier, if any.
	err := cl.StopClassification()
	if aggErr := cl.StopAggregation(); err == nil {
		err = aggErr
	}
	return err
}

// We factor out this helper function partly because the code is otherwise common
// between several places later on, and partly because unlike Perl, Go does not have
// auto-vivification of hashes and arrays, so that means that the common code would
// otherwise be even longer at each place it would occur.
//
// We design SaveRegexes() to accept all the regexes for a given {group, category}
// pair in one call.  That makes it possible to detect if the user has listed the
// same category multiple times under the same group, which is something we must
// rule out because it would be overall too confusing.
//
func (cl *Classify) SaveRegexes(group string, category string, allRegexes []string) error {
	if cl.groupCategoriesSeen == nil {
		cl.groupCategoriesSeen = make(map[string]map[string]bool)
	}
	if cl.groupCategoriesSeen[group] == nil {
		cl.groupCategoriesSeen[group] = make(map[string]bool)
	}
	if cl.groupCategoriesSeen[group][category] {
		return fmt.Errorf("found a second instance of category %q for mapped_selector_regexes group %q", category, group)
	}
	cl.groupCategoriesSeen[group][category] = true
	allRegexPtrs := make([]*regexp.Regexp, 0)
	for _, regex := range allRegexes {
		if regex == "" {
			return fmt.Errorf("found an empty mapped_selector_regexes regex for group %q category %q", group, category)
		}
		switch regexPtr, err := regexp.Compile(regex); err {
		default:
			return fmt.Errorf("invalid mapped_selector_regexes regular expression %q for group %q, category %q:\n%v",
				regex, group, category, err)
		case nil:
			allRegexPtrs = append(allRegexPtrs, regexPtr)
		}
	}
	if len(allRegexPtrs) > 0 {
		// If we never call SaveRegexes(), this map never gets initialized.  But we
		// check for that condition after running all loops that call SaveRegexes(),
		// to ensure that if that happens, the configuration is rejected.
		if cl.mappedRegexes == nil {
			cl.mappedRegexes = make(map[string][]map[string][]*regexp.Regexp)
		}
		if cl.mappedRegexes[group] == nil {
			cl.mappedRegexes[group] = make([]map[string][]*regexp.Regexp, 0)
		}
		cl.mappedRegexes[group] = append(cl.mappedRegexes[group], map[string][]*regexp.Regexp{category: allRegexPtrs})
	}

	return nil
}

// This is where we do the work of translating the exernally-specified config data
// into an internal form used for the processor side of this plugin.
func (cl *Classify) InitClassification(acc telegraf.Accumulator) error {
	cl.acc = acc

	// It's okay not to define either selector_tag or selector_field.  In that case,
	// default_regex_group will come into play.  (If default_regex_group is also not
	// defined, or it is defined as an empty string, all input data points will be
	// dropped.  But at least for now, we consider that to be a valid [if unuseful]
	// configuration, and we don't err out on that here.  Perhaps in the future we
	// might decide to reject such a configuration.)
	if cl.SelectorTag != "" && cl.SelectorField != "" {
		return fmt.Errorf("selector_tag and selector_field cannot both be defined")
	}
	if cl.MatchTag == "" && cl.MatchField == "" {
		return fmt.Errorf("either match_tag or match_field must be defined")
	}
	if cl.MatchTag != "" && cl.MatchField != "" {
		return fmt.Errorf("match_tag and match_field cannot both be defined")
	}
	if cl.ResultTag == "" && cl.ResultField == "" {
		return fmt.Errorf("either result_tag or result_field must be defined")
	}
	if cl.ResultTag != "" && cl.ResultField != "" {
		return fmt.Errorf("result_tag and result_field cannot both be defined")
	}

	haveSeenSelectorRegex := make(map[string]bool)
	cl.selectorMap = make([]map[*regexp.Regexp]string, 0)
	if cl.SelectorMapping != nil {
		for _, mapping := range cl.SelectorMapping {
			if len(mapping) > 1 {
				return fmt.Errorf("the selector_mapping includes more than one regex-group mapping in one of the elements")
			}
			for regex, group := range mapping {
				if regex == "" {
					return fmt.Errorf("found an empty selector_mapping regex for selector group %q", group)
				}
				if haveSeenSelectorRegex[regex] {
					return fmt.Errorf("found duplicate selector_mapping regex %q", regex)
				}
				haveSeenSelectorRegex[regex] = true

				// In theory, we would want to validate the selector mapping "group" as best we can.
				// It should be one of:
				// * the special string "*", in which case we cannot validate against keys in the
				//   cl.MappedSelectorRegexes map since we don't know what selector item values
				//   will appear at run time
				// * a non-empty literal string, in which case it should generally match some key
				//   in the cl.MappedSelectorRegexes map; but if it does not, that is not an error,
				//   and we will later need to substitute the cl.DefaultRegexGroup value for the
				//   group that the user specified in the configuration, perhaps with some logged
				//   warning about that
				// * an empty string (meaning an input data point with a matching selector item
				//   value will just be dropped)
				// In practice, that covers the waterfront, so there is not much we can due to
				// perform validation here.

				switch regexPtr, err := regexp.Compile(regex); err {
				default:
					return fmt.Errorf("invalid selector_mapping regular expression %q for group %q:\n%v",
						regex, group, err)
				case nil:
					cl.selectorMap = append(cl.selectorMap, map[*regexp.Regexp]string{regexPtr: group})
				}
			}
		}
	}

	// When making calls to SaveRegexes(), we purposely do not check whether the set of regexes we are passing
	// is empty, before we make that call, in a possible attempt to skip that call.  That allows us to centralize
	// the logic that validates that we do not see the same category multiple times within a given regex group.
	for group, categoryHashes := range cl.MappedSelectorRegexes {
		for _, categoryRegexes := range categoryHashes {
			for category, regexes := range categoryRegexes {
				if regexString, ok := regexes.(string); ok {
					// This test is not quite accurate for what we want to accomplish, but there might be no good
					// way to make it totally accurate.  Suppose we just have a single regex, on a line terminated
					// with the closing three single-quote characters of a multi-line literal instead of having
					// the close quotes on a separate line.  In that case, we want to treat that as a multi-line
					// literal, and trim the leading whitespace.  But in that one special corner case, we probably
					// have no information on whether the string we see inside the code was presented as a simple
					// literal string or as a multi-line literal string in the config file.
					// if (regexString is a multi-line string) ...
					isMultiline := newlineRegex.MatchString(regexString)
					if isMultiline {
						// split regexString into individual lines, to be processed separately
						allRegexes := make([]string, 0)
						for _, regex := range strings.Split(regexString, "\n") {
							regex = strings.TrimSpace(regex)
							//
							// We adopt a convention here that blank lines within the multi-line format will be ignored.
							// That helps the user by allowing them to visually separate clusters of related regexes
							// within a given category.  But it also helps us here in the code, in two ways:
							//
							// (*) We need to gracefully handle the situation if the user provides invisible whitespace
							//     immediately after the opening delimiter and before the newline at the end of that
							//     line.  Such whitespace will not be trimmed away by TOML, so we must deal with it here.
							//
							// (*) We need to handle the trailing newline at the end of the last regex in a multi-line
							//     string, and extra whitespace on the next line, without objecting that we found an
							//     empty regex because of that next line.
							//
							if regex != "" {
								allRegexes = append(allRegexes, regex)
							}
						}
						if err := cl.SaveRegexes(group, category, allRegexes); err != nil {
							return err
						}
					} else {
						if err := cl.SaveRegexes(group, category, []string{regexString}); err != nil {
							return err
						}
					}
				} else if regexArray, ok := regexes.([]string); ok {
					// We might be provided an array of strings, typically during a unit test which forces
					// the value to be in that precise format.
					if err := cl.SaveRegexes(group, category, regexArray); err != nil {
						return err
					}
				} else if interfaceArray, ok := regexes.([]interface{}); ok { //nolint:revive // linter is simply wrong about its analysis and advice
					// We might be provided an array of interface{}s, typically by the BurntSushi/toml parser
					// when it encounters an array of strings.  (Exactly why it does not declare the elements
					// to be strings, when it ought to know that from the value format in the config file, is
					// unknown to us.  Perhaps it is because we don't have a declaration in our Go structure
					// of the datatypes down to this level, because we wanted flexibility at just above this
					// level, and the parser is taking that as more of a clue [since TOML array elements can
					// be of mixed types within the same array].)  Regardless of the reason, we must adapt to
					// what we see from the parser as compared to what we are expecting the user to provide.
					// We do at least check our assumptions before accepting the parser output.
					regexArray := make([]string, 0)
					for _, interfaceElement := range interfaceArray {
						switch regex, ok := interfaceElement.(string); ok {
						default:
							return fmt.Errorf("invalid mapped_selector_regexes regular expression construction for group %q, category %q:\n"+
								"one of the array values is not a string: %v", group, category, interfaceElement)
						case true:
							regexArray = append(regexArray, regex)
						}
					}
					if err := cl.SaveRegexes(group, category, regexArray); err != nil {
						return err
					}
				} else {
					return fmt.Errorf("invalid mapped_selector_regexes regular expression construction for group %q, category %q:\n"+
						"the value is not a string or an array of strings (it is a %T instead)", group, category, regexes)
				}
			}
		}
	}

	// We had to wait until now to do the following work, because only now will
	// cl.mappedRegexes be defined and populated, in calls to cl.SaveRegexes() in
	// the loop just above.

	if cl.mappedRegexes == nil {
		// There are no regex groups configured (cl.mappedRegexes is an empty map).
		// Return an error because this is an invalid configuration.
		return fmt.Errorf("invalid configuration:  there are no mapped_selector_regexes groups defined that contain category regexes")
	}

	cl.allRegexCategories = make(map[string]bool)
	for _, categoryList := range cl.mappedRegexes {
		for _, categoryDefinition := range categoryList {
			for category := range categoryDefinition {
				cl.allRegexCategories[category] = true
			}
		}
	}

	// There is no validation of the cl.DefaultCategory value.  We might want to
	// tell the difference between default_category not being defined at all in the
	// config file (which would of course be allowed), and it being defined by the
	// user as an empty string (which doesn't make a lot of sense, and perhaps should
	// be disallowed).  I suppose we could have declared DefaultCategory to be of
	// type "interface{}" instead of type "string", and that would have provided
	// the necessary distinction, by our checking the type of that variable.  But
	// it doesn't seem to be important enough to do that, so we just treat an empty
	// string the same as an undefined option, meaning there is no default category
	// in play.
	//
	// We also might have checked to see whether cl.DefaultCategory matches one of the
	// categories listed in the mappedRegexes, and declare the configuration invalid
	// if that is not the case.  We choose not to do that, because it seems like the
	// user ought to be able to declare an "unknown" or "out-of-bounds" category via
	// the default_category mechanism, not associating it with any regexes.  After
	// all, that might be a reasonable interpretation of "default".

	cl.dropThisCategory = make(map[string]bool)
	if cl.DropCategories != nil {
		// Before we copy the data to an internal data structure for validation and use in
		// operation, we normalize its structure, which can be polymorphic in the config file.
		if category, ok := cl.DropCategories.(string); ok {
			cl.DropCategories = []string{category}
		} else if _, ok := cl.DropCategories.([]string); !ok {
			return fmt.Errorf("drop_categories is neither a string nor an array of strings")
		}
		for _, category := range cl.DropCategories.([]string) {
			if category == "" {
				return fmt.Errorf("found an empty string for a category in drop_categories")
			}
			if !cl.allRegexCategories[category] && category != cl.DefaultCategory {
				return fmt.Errorf("%q is named in drop_categories but is not\n"+
					"used as a category in any mapped_selector_regexes group\n"+
					" or as the value of default_category", category)
			}
			cl.dropThisCategory[category] = true
		}
	}

	return nil
}

// This is where we do the work of translating the exernally-specified config data
// into an internal form used for the aggregator side of this plugin.
func (cl *Classify) InitAggregation() error {
	var err error
	if cl.AggregationPeriod != "" {
		if cl.aggregationTimePeriod, err = time.ParseDuration(cl.AggregationPeriod); err != nil {
			return fmt.Errorf("invalid aggregation_period: %v", err)
		}
		if cl.aggregationTimePeriod < time.Second {
			return fmt.Errorf("you cannot specify an aggregation_period shorter than a second")
		}
	} else {
		cl.aggregationTimePeriod = 0
	}

	// All distinct categories found in the MappedSelectorRegexes, plus AggregationDroppedField
	// and AggregationTotalField, put together in a hash for quick lookup.
	allLegalCategories := make(map[string]bool)

	// We make a separate copy of the allLegalCategories map so as not to
	// change the original copy of that map when we add more entries to it.
	for category, flag := range cl.allRegexCategories {
		allLegalCategories[category] = flag
	}

	// We validate that the cl.AggregationDroppedField and cl.AggregationTotalField do not match
	// any other configured category names, so we are never in danger of double-counting points
	// in aggregation statistics.
	//
	if cl.AggregationDroppedField != "" {
		if _, ok := cl.allRegexCategories[cl.AggregationDroppedField]; ok {
			return fmt.Errorf("invalid aggregation_dropped_field %q (matches a category in mapped_selector_regexes)", cl.AggregationDroppedField)
		}
		if cl.AggregationDroppedField == cl.DefaultCategory {
			return fmt.Errorf("invalid aggregation_dropped_field %q (matches the default_category)", cl.AggregationDroppedField)
		}
		allLegalCategories[cl.AggregationDroppedField] = true
	}
	if cl.AggregationTotalField != "" {
		if _, ok := cl.allRegexCategories[cl.AggregationTotalField]; ok {
			return fmt.Errorf("invalid aggregation_total_field %q (matches a category in mapped_selector_regexes)", cl.AggregationTotalField)
		}
		if cl.AggregationTotalField == cl.AggregationDroppedField {
			return fmt.Errorf("invalid aggregation_total_field %q (matches the aggregation_dropped_field)", cl.AggregationTotalField)
		}
		if cl.AggregationTotalField == cl.DefaultCategory {
			return fmt.Errorf("invalid aggregation_total_field %q (matches the default_category)", cl.AggregationTotalField)
		}
		allLegalCategories[cl.AggregationTotalField] = true
	}

	if (cl.AggregationSummaryTag == "") != (cl.AggregationSummaryValue == "") {
		return fmt.Errorf("aggregation_summary_tag and aggregation_summary_value must both be either defined or undefined")
	}

	for _, field := range cl.AggregationSummaryFields {
		if field == "" {
			return fmt.Errorf("aggregation_summary_fields cannot contain any empty strings")
		}
		if !allLegalCategories[field] {
			return fmt.Errorf("%q is listed in aggregation_summary_fields\n"+
				"    but is not either a defined regex category or equal to the value\n"+
				"    of aggregation_dropped_field or aggregation_total_field", field)
		}
	}
	for _, field := range cl.AggregationGroupFields {
		if field == "" {
			return fmt.Errorf("aggregation_group_fields cannot contain any empty strings")
		}
		if !allLegalCategories[field] {
			return fmt.Errorf("%q is listed in aggregation_group_fields\n"+
				"    but is not either a defined regex category or equal to the value\n"+
				"    of aggregation_dropped_field or aggregation_total_field", field)
		}
	}
	for _, field := range cl.AggregationSelectorFields {
		if field == "" {
			return fmt.Errorf("aggregation_selector_fields cannot contain any empty strings")
		}
		if !allLegalCategories[field] {
			return fmt.Errorf("%q is listed in aggregation_selector_fields\n"+
				"    but is not either a defined regex category or equal to the value\n"+
				"    of aggregation_dropped_field or aggregation_total_field", field)
		}
	}

	cl.doAggregation = false
	cl.doSummaryAggregation = false
	cl.doGroupAggregation = false
	cl.doSelectorAggregation = false
	if cl.aggregationTimePeriod != 0 && cl.AggregationMeasurement != "" {
		if cl.AggregationSummaryTag != "" && cl.AggregationSummaryValue != "" && len(cl.AggregationSummaryFields) != 0 {
			cl.doSummaryAggregation = true
		}
		if cl.AggregationGroupTag != "" && len(cl.AggregationGroupFields) != 0 {
			cl.doGroupAggregation = true
		}
		if cl.AggregationSelectorTag != "" && len(cl.AggregationSelectorFields) != 0 {
			cl.doSelectorAggregation = true
		}
		cl.doAggregation = cl.doSummaryAggregation || cl.doGroupAggregation || cl.doSelectorAggregation
	}

	if cl.doAggregation {
		// * create some corresponding aggregation counters (others will depend on details
		//   of the input data points, and be dynamically created as data is processed)
		// * set the created counters to zero

		// For summary-level statistics, we create all the necessary counters in advance,
		// since none of them depend dynamically on the input data points seen during the
		// aggregation period.  This saves us from the extra work of auto-vivifying these
		// counters later on while processing input data points.
		//
		// For regex-group-level and selector-level statistics, we don't create any counters
		// in advance, because we would prefer that they be dynamically allocated only
		// for those aggegation species (regex groups and selectors, respectively) that
		// actually show up as a consequence of processing input data points.  That dynamic
		// allocation saves us from extra work in walking the aggregation counters to see
		// which sets contain some non-zero values, when outputting the values of those
		// counters.  So the only allocation we need to do here is for the base maps, not
		// the second-level maps.

		if cl.doSummaryAggregation {
			cl.sharedData.aggregationSummary = make(map[string]int)
			for _, category := range cl.AggregationSummaryFields {
				cl.sharedData.aggregationSummary[category] = 0
			}
		}
		if cl.doGroupAggregation {
			cl.sharedData.aggregationByGroup = make(map[string]map[string]int)
		}
		if cl.doSelectorAggregation {
			cl.sharedData.aggregationBySelector = make(map[string]map[string]int)
		}
	}

	return nil
}

func (cl *Classify) InitSynchronization() error {
	// if cl.doAggregation {
	// // There is no need to initialize the required thread-synchronization object
	// // (aggregationMutex).  Its initial zero value is already an unlocked mutex,
	// // and we have no way to force that state if it is not already in place (see
	// // the Reset() function).
	// }
	return nil
}

func (cl *Classify) StartClassification() error {
	// This routine is a placeholder for any work that needs to be done
	// specifically for the classification side of the calculations.
	return nil
}

func (cl *Classify) StopClassification() error {
	// This routine is a placeholder for any work that needs to be done
	// specifically for the classification side of the calculations.
	return nil
}

func (cl *Classify) StartAggregation() error {
	if cl.doAggregation {
		cl.stopRequested = make(chan bool)
		cl.syncWaitGroup.Add(1)
		go cl.RunAggregation()
	}
	return nil
}

func (cl *Classify) StopAggregation() error {
	if cl.doAggregation {
		cl.stopRequested <- true
		// Block, waiting for cl.RunAggregation() to stop.
		cl.syncWaitGroup.Wait()
	}
	return nil
}

// This function is designed to run as a goroutine, so it doesn't return any value.
func (cl *Classify) RunAggregation() {
	// This should never be invoked in practice, because we should not have any
	// bugs that panic the code.  It is here only because the ordinary data that
	// is printed out when an actual panic occurs seems to be wanting.  We need
	// to know exactly where the panic occurred so we can track it down.  And if
	// a panic occurs in this aggregation thread, we need to have the recover()
	// routine called from within that thread, not from within the main thread,
	// for it to do any good.
	defer func() {
		if p := recover(); p != nil {
			if cl.logger != nil {
				cl.logger.Infof("internal panic: %v", p)
			}
			fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
		}
	}()

	// When RunAggregation() exits, flag that fact so the cl.syncWaitGroup.Wait()
	// call in StopAggregation() can return.
	defer cl.syncWaitGroup.Done()

	// We prefer the aggregation-cycle ticker to not have an uncontrolled phase within
	// its period.  Instead, we generally wish the initial wait period to be shorter
	// than a full cl.aggregationTimePeriod, so ticker period endpoints align naturally
	// with the obvious nominal clock points (e.g.,  00:05:00, 00:10:00, and so forth,
	// for a 300-second cl.aggregationTimePeriod).
	//
	// It would be more convenient if time.NewTicker() supported an optional second
	// argument to specify the initial ticker duration, after which the standard
	// ticker duration would kick in.  That would allow you to more readily control
	// the phase of the ticking with respect to wall-clock time.  Without that, we
	// have to complexify the code here to achieve the same effect.  The initial
	// ticker expiration is placed into the far future so in fact it will never fire;
	// it is only present so we can get the datatype of that variable defined up
	// front and so the first reference to the ticker in the select{} does not panic.
	// That ticker will be replaced with the operational ticker as soon as we are
	// synchronized with wall-clock time.
	//
	now := time.Now()
	timer := time.NewTimer(now.Add(cl.aggregationTimePeriod).Truncate(cl.aggregationTimePeriod).Sub(now))
	ticker := time.NewTicker(1_000_000 * time.Hour)
	var t time.Time
	for quit := false; !quit; {
		// Block until either the timer expires (telling us we are at the end
		// of the first (likely foreshortened) aggregation period), or we get
		// the next tick, or we are requested to stop.  In each case, capture
		// the end-of-wait timestamp, which we will use as the timestamp of
		// any aggregated data we put out in this cycle.
		select {
		case t = <-timer.C:
			ticker.Stop()
			ticker = time.NewTicker(cl.aggregationTimePeriod)
		case t = <-ticker.C:
		case quit = <-cl.stopRequested:
			t = time.Now()
		}
		// Carry out one iteration of aggregation, whether it be the initial cycle
		// (that will typically be shorter than usual, to sync up with the desired
		// clock phase), an ordinary cycle, or the last cycle (also probably short,
		// run as soon as possible after shutdown is requested) that flushes out
		// all remaining data before stopping.
		cl.OutputAggregationData(t)
	}
	timer.Stop()
	ticker.Stop()
}

// Helper function for OutputAggregationData(), to factor out common code.
func (cl *Classify) GenerateMetric(tagName string, tagValue string, counters map[string]int, timestamp time.Time) {
	fields := make(map[string]interface{})
	haveNonzeroCounter := false
	for category, count := range counters {
		// If all the counters for this aggregation data point are zero, the entire data
		// point will be suppressed, to reduce noise in the system.  If at least one counter
		// is non-zero, the data point will of course be generated and sent downstream.
		// By default, all the individual item-category counters in that data point that
		// have zero values will be omitted from that data point, so as not to use extra
		// processing power and take up more space in downstream storage.  If you do want
		// those zero values to be present in the aggregated-data output data point, you can
		// use the aggregation_includes_zeroes config option to enable that behavior
		if count > 0 {
			haveNonzeroCounter = true
			fields[category] = count
		} else if cl.AggregationIncludesZeroes {
			fields[category] = count
		}
	}
	if haveNonzeroCounter {
		// We could have added the aggregated statistics as untyped metrics, but
		// somewhat arbitrarily we chose to instead label them as counters.
		//
		// The thread that calls this routine will be running asynchronously with respect to
		// the main thread that is processing input data points.  We depend on synchronization
		// which is built into the implementation of telegraf.Accumulator itself (whether that
		// be the testing version or the production version of the accumulator) to be sure
		// that access to the accumulator for output from this routine (adding to the cl.acc
		// telegraf.Accumulator) is cleanly interleaved with both access to the accumulator for
		// output from the main thread and access to the accumulator from Telegraf itself to
		// send data points downstream.  That is needed so downstream plugins do not see mangled
		// data from mashed-together data points, or other forms of data-structure corruption.
		//
		// Notice the odd placement of fields before tags in this call.  That just seems wrong,
		// given that it does not match the order of elements in the InfluxDB Line Protocol.
		// But that's the way this call is documented:
		// https://pkg.go.dev/github.com/influxdata/telegraf#Accumulator
		//
		cl.acc.AddCounter(cl.AggregationMeasurement, fields,
			map[string]string{tagName: tagValue}, timestamp)
	}
}

func (cl *Classify) OutputAggregationData(timestamp time.Time) {
	// Run one cycle of outputting accumulated statistics-aggregation counts
	// as data points separate from the input data points being processed by
	// the main thread of this plugin.  Synchronize access to the relevant
	// counters so no corruption can occur.  After outputting the data
	// points for the aggregated data, either reset the relevant counters
	// (for static counters) or destroy the counters (for counters which are
	// dynamically auto-vivified during each aggregation cycle).

	cl.sharedData.aggregationMutex.Lock()
	defer cl.sharedData.aggregationMutex.Unlock()

	if cl.doSummaryAggregation {
		cl.GenerateMetric(cl.AggregationSummaryTag, cl.AggregationSummaryValue,
			cl.sharedData.aggregationSummary, timestamp)

		// Now that we have possibly output the single summary line that is
		// appropriate in this context, we reset all the associated counters
		// so they will be freshly initialized for the next aggregation period.
		for category := range cl.sharedData.aggregationSummary {
			cl.sharedData.aggregationSummary[category] = 0
		}
	}
	if cl.doGroupAggregation {
		for group, categoryCounts := range cl.sharedData.aggregationByGroup {
			cl.GenerateMetric(cl.AggregationGroupTag, group, categoryCounts, timestamp)
		}

		// Destroy all the group-level counters.
		cl.sharedData.aggregationByGroup = make(map[string]map[string]int)
	}
	if cl.doSelectorAggregation {
		for selector, categoryCounts := range cl.sharedData.aggregationBySelector {
			cl.GenerateMetric(cl.AggregationSelectorTag, selector, categoryCounts, timestamp)
		}

		// Destroy all the selector-level counters.
		cl.sharedData.aggregationBySelector = make(map[string]map[string]int)
	}
}

// This function is called by Telegraf to add this plugin into the running configuration.
func init() {
	processors.AddStreaming("classify", func() telegraf.StreamingProcessor {
		// There are no default values for any of the config parameters,
		// so we don't specify any here when we create this object.
		return &ClassifyWrapper{}
	})
}
