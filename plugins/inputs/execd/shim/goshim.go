package shim

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
)

type empty struct{}

var (
	gatherPromptChans []chan empty
	stdout            io.Writer = os.Stdout
	stdin             io.Reader = os.Stdin
)

const (
	// PollIntervalDisabled is used to indicate that you want to disable polling,
	// as opposed to duration 0 meaning poll constantly.
	PollIntervalDisabled = time.Duration(0)
)

type Shim struct {
	Inputs []telegraf.Input
}

func New() *Shim {
	return &Shim{}
}

// AddInput adds the input to the shim. Later calls to Run() will run this input.
func (s *Shim) AddInput(input telegraf.Input) error {
	if p, ok := input.(telegraf.Initializer); ok {
		err := p.Init()
		if err != nil {
			return fmt.Errorf("failed to init input: %s", err)
		}
	}

	s.Inputs = append(s.Inputs, input)
	return nil
}

// AddInputs adds multiple inputs to the shim. Later calls to Run() will run these.
func (s *Shim) AddInputs(newInputs []telegraf.Input) error {
	for _, inp := range newInputs {
		if err := s.AddInput(inp); err != nil {
			return err
		}
	}
	return nil
}

// Run the input plugins..
func (s *Shim) Run(pollInterval time.Duration) error {
	wg := sync.WaitGroup{}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	collectMetricsPrompt := make(chan os.Signal, 1)
	listenForCollectMetricsSignals(collectMetricsPrompt)

	wg.Add(1) // wait for the metric channel to close
	metricCh := make(chan telegraf.Metric, 1)

	serializer := influx.NewSerializer()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, input := range s.Inputs {
		wrappedInput := inputShim{Input: input}

		acc := agent.NewAccumulator(wrappedInput, metricCh)
		acc.SetPrecision(time.Nanosecond)

		if serviceInput, ok := input.(telegraf.ServiceInput); ok {
			if err := serviceInput.Start(acc); err != nil {
				return fmt.Errorf("failed to start input: %s", err)
			}
		}
		gatherPromptCh := make(chan empty, 1)
		gatherPromptChans = append(gatherPromptChans, gatherPromptCh)
		wg.Add(1)
		go func(input telegraf.Input) {
			startGathering(ctx, input, acc, gatherPromptCh, pollInterval)
			if serviceInput, ok := input.(telegraf.ServiceInput); ok {
				serviceInput.Stop()
			}
			wg.Done()
		}(input)
	}

	go stdinCollectMetricsPrompt(ctx, collectMetricsPrompt)

loop:
	for {
		select {
		case <-quit:
			// cancel, but keep looping until the metric channel closes.
			cancel()
		case <-collectMetricsPrompt:
			collectMetrics(ctx)
		case m, open := <-metricCh:
			if !open {
				wg.Done()
				break loop
			}
			b, err := serializer.Serialize(m)
			if err != nil {
				return fmt.Errorf("failed to serialize metric: %s", err)
			}
			// Write this to stdout
			fmt.Fprint(stdout, string(b))
		}
	}

	wg.Wait()
	return nil
}

func hasQuit(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func stdinCollectMetricsPrompt(ctx context.Context, collectMetricsPrompt chan<- os.Signal) {
	s := bufio.NewScanner(stdin)
	// for every line read from stdin, make sure we're not supposed to quit,
	// then push a message on to the collectMetricsPrompt
	for s.Scan() {
		// first check if we should quit
		if hasQuit(ctx) {
			return
		}

		// now push a non-blocking message to trigger metric collection.
		pushCollectMetricsRequest(collectMetricsPrompt)
	}
}

// pushCollectMetricsRequest pushes a non-blocking (nil) message to the
// collectMetricsPrompt channel to trigger metric collection.
// The channel is defined with a buffer of 1, so if it's full, duplicated
// requests are discarded.
func pushCollectMetricsRequest(collectMetricsPrompt chan<- os.Signal) {
	select {
	case collectMetricsPrompt <- nil:
	default:
	}
}

func collectMetrics(ctx context.Context) {
	if hasQuit(ctx) {
		return
	}
	for i := 0; i < len(gatherPromptChans); i++ {
		// push a message out to each channel to collect metrics. don't block.
		select {
		case gatherPromptChans[i] <- empty{}:
		default:
		}
	}
}

func startGathering(ctx context.Context, input telegraf.Input, acc telegraf.Accumulator, gatherPromptCh <-chan empty, pollInterval time.Duration) {
	if pollInterval == PollIntervalDisabled {
		return // don't poll
	}
	t := time.NewTicker(pollInterval)
	defer t.Stop()
	for {
		// give priority to stopping.
		if hasQuit(ctx) {
			return
		}
		// see what's up
		select {
		case <-ctx.Done():
			return
		case <-gatherPromptCh:
			if err := input.Gather(acc); err != nil {
				fmt.Fprintf(os.Stderr, "failed to gather metrics: %s", err)
			}
		case <-t.C:
			if err := input.Gather(acc); err != nil {
				fmt.Fprintf(os.Stderr, "failed to gather metrics: %s", err)
			}
		}
	}
}

// LoadConfig loads and adds the inputs to the shim
func (s *Shim) LoadConfig(filePath *string) error {
	loadedInputs, err := LoadConfig(filePath)
	if err != nil {
		return err
	}
	return s.AddInputs(loadedInputs)
}

// DefaultImportedPlugins defaults to whatever plugins happen to be loaded and
// have registered themselves with the registry. This makes loading plugins
// without having to define a config dead easy.
func DefaultImportedPlugins() (i []telegraf.Input, e error) {
	for _, inputCreatorFunc := range inputs.Inputs {
		i = append(i, inputCreatorFunc())
	}
	return i, nil
}

// LoadConfig loads the config and returns inputs that later need to be loaded.
func LoadConfig(filePath *string) ([]telegraf.Input, error) {
	if filePath == nil {
		return DefaultImportedPlugins()
	}

	b, err := ioutil.ReadFile(*filePath)
	if err != nil {
		return nil, err
	}

	conf := struct {
		Inputs map[string][]toml.Primitive
	}{}

	md, err := toml.Decode(string(b), &conf)
	if err != nil {
		return nil, err
	}

	loadedInputs, err := loadConfigIntoInputs(md, conf.Inputs)

	if len(md.Undecoded()) > 0 {
		fmt.Fprintf(stdout, "Some plugins were loaded but not used: %q\n", md.Undecoded())
	}
	return loadedInputs, err
}

func loadConfigIntoInputs(md toml.MetaData, inputConfigs map[string][]toml.Primitive) ([]telegraf.Input, error) {
	renderedInputs := []telegraf.Input{}

	for name, primitives := range inputConfigs {
		inputCreator, ok := inputs.Inputs[name]
		if !ok {
			return nil, errors.New("unknown input " + name)
		}

		for _, primitive := range primitives {
			inp := inputCreator()
			// Parse specific configuration
			if err := md.PrimitiveDecode(primitive, inp); err != nil {
				return nil, err
			}

			renderedInputs = append(renderedInputs, inp)
		}
	}
	return renderedInputs, nil
}
