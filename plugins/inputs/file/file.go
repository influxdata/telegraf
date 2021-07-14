package file

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dimchansky/utfbom"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/plugins/common/encoding"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type File struct {
	Files             []string `toml:"files"`
	FileTag           string   `toml:"file_tag"`
	CharacterEncoding string   `toml:"character_encoding"`

	parserFunc telegraf.ParserFunc
	filenames  []string
	decoder    *encoding.Decoder

	// Read from stdin, a hidden configuration for isolating processes
	Stdin bool
}

const sampleConfig = `
  ## Files to parse each interval.  Accept standard unix glob matching rules,
  ## as well as ** to match recursive files and directories.
  files = ["/tmp/metrics.out"]


  ## Name a tag containing the name of the file the data was parsed from.  Leave empty
  ## to disable. Cautious when file name variation is high, this can increase the cardinality
  ## significantly. Read more about cardinality here:
  ## https://docs.influxdata.com/influxdb/cloud/reference/glossary/#series-cardinality
  # file_tag = ""
  #

  ## Character encoding to use when interpreting the file contents.  Invalid
  ## characters are replaced using the unicode replacement character.  When set
  ## to the empty string the data is not decoded to text.
  ##   ex: character_encoding = "utf-8"
  ##       character_encoding = "utf-16le"
  ##       character_encoding = "utf-16be"
  ##       character_encoding = ""
  # character_encoding = ""

  ## The dataformat to be read from files
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

// SampleConfig returns the default configuration of the Input
func (f *File) SampleConfig() string {
	return sampleConfig
}

func (f *File) Description() string {
	return "Parse a complete file each interval"
}

func (f *File) Init() error {

	// SECCOMP MAGIC
	var syscalls = []string{"futex", "rt_sigaction", "mmap", "rt_sigprocmask", "pread64", "clone", "mprotect", "openat", "read",
		"epoll_ctl", "close", "rt_sigreturn", "readlinkat", "fcntl", "fstat", "munmap", "sched_yield", "getrandom", "brk", "pipe2",
		"sigaltstack", "epoll_create1", "write", "ioctl", "epoll_pwait", "arch_prctl", "gettid", "sched_getaffinity", "set_tid_address",
		"prlimit64", "connect", "getsockname", "getpeername", "set_robust_list", "access", "getpid", "socket", "execve", "uname", "sysinfo",
		"getuid", "getgid", "newfstatat", "exit_group", "readdir", "getdents64"}
	internal.WhiteList(syscalls)
	// SECCOMP MAGIC

	var err error
	f.decoder, err = encoding.NewDecoder(f.CharacterEncoding)
	return err
}

func (f *File) Gather(acc telegraf.Accumulator) error {
	err := f.refreshFilePaths()
	if err != nil {
		return err
	}
	if f.Stdin {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			metrics, err := f.parser.Parse(scanner.Bytes())
			if err != nil {
				return err
			}
			for _, m := range metrics {
				acc.AddMetric(m)
			}
		}
		if err := scanner.Err(); err != nil {
			return err
		}
	} else {
		for _, k := range f.filenames {
			metrics, err := f.readMetric(k)
			if err != nil {
				return err
			}

			for _, m := range metrics {
				if f.FileTag != "" {
					m.AddTag(f.FileTag, filepath.Base(k))
				}
				acc.AddMetric(m)
			}
		}
	}

	return nil
}

func (f *File) SetParserFunc(fn telegraf.ParserFunc) {
	f.parserFunc = fn
}

func (f *File) refreshFilePaths() error {
	var allFiles []string
	fmt.Println("files", f.Files[0])
	for _, file := range f.Files {
		g, err := globpath.Compile(file)
		if err != nil {
			return fmt.Errorf("could not compile glob %v: %v", file, err)
		}
		files := g.Match()
		if len(files) <= 0 {
			return fmt.Errorf("could not find file: %v", file)
		}
		allFiles = append(allFiles, files...)
	}

	f.filenames = allFiles
	return nil
}

func (f *File) readMetric(filename string) ([]telegraf.Metric, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r, _ := utfbom.Skip(f.decoder.Reader(file))
	fileContents, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("could not read %q: %s", filename, err)
	}
	parser, err := f.parserFunc()
	if err != nil {
		return nil, fmt.Errorf("could not instantiate parser: %s", err)
	}
	return parser.Parse(fileContents)
}

func init() {
	inputs.Add("file", func() telegraf.Input {
		return &File{}
	})
}
