//go:generate ../../../tools/readme_config_includer/generator
package remotefile

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/fspath"
	"github.com/rclone/rclone/vfs"
	"github.com/rclone/rclone/vfs/vfscommon"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

//go:embed sample.conf
var sampleConfig string

type File struct {
	Remote            config.Secret   `toml:"remote"`
	Files             []string        `toml:"files"`
	FinalWriteTimeout config.Duration `toml:"final_write_timeout"`
	WriteBackInterval config.Duration `toml:"cache_write_back"`
	MaxCacheSize      config.Size     `toml:"cache_max_size"`
	UseBatchFormat    bool            `toml:"use_batch_format"`
	Trace             bool            `toml:"trace"`
	Log               telegraf.Logger `toml:"-"`

	root     *vfs.VFS
	fscancel context.CancelFunc
	vfsopts  vfscommon.Options

	templates  []*template.Template
	serializer serializers.Serializer
}

func (*File) SampleConfig() string {
	return sampleConfig
}

func (f *File) SetSerializer(serializer serializers.Serializer) {
	f.serializer = serializer
}

func (f *File) Init() error {
	if len(f.Files) == 0 {
		return errors.New("no files specified")
	}

	// Set defaults
	if f.Remote.Empty() {
		if err := f.Remote.Set([]byte("local")); err != nil {
			return fmt.Errorf("setting default remote failed: %w", err)
		}
	}

	if f.FinalWriteTimeout <= 0 {
		f.FinalWriteTimeout = config.Duration(10 * time.Second)
	}

	// Prepare VFS options
	f.vfsopts = vfscommon.DefaultOpt
	f.vfsopts.CacheMode = vfscommon.CacheModeWrites // required for appends
	if f.WriteBackInterval > 0 {
		f.vfsopts.WriteBack = time.Duration(f.WriteBackInterval)
	}
	if f.MaxCacheSize > 0 {
		f.vfsopts.CacheMaxSize = fs.SizeSuffix(f.MaxCacheSize)
	}

	// Redirect logging
	fs.LogPrint = func(level fs.LogLevel, text string) {
		if !f.Trace {
			return
		}
		f.Log.Debugf("[%s] %s", level.String(), text)
	}

	// Setup custom template functions
	funcs := template.FuncMap{"now": time.Now}

	// Setup filename templates
	f.templates = make([]*template.Template, 0, len(f.Files))
	for _, ftmpl := range f.Files {
		tmpl, err := template.New(ftmpl).Funcs(funcs).Parse(ftmpl)
		if err != nil {
			return fmt.Errorf("parsing file template %q failed: %w", ftmpl, err)
		}
		f.templates = append(f.templates, tmpl)
	}

	return nil
}

func (f *File) Connect() error {
	remoteRaw, err := f.Remote.Get()
	if err != nil {
		return fmt.Errorf("getting remote secret failed: %w", err)
	}
	remote := remoteRaw.String()
	remoteRaw.Destroy()

	// Construct the underlying filesystem config
	parsed, err := fspath.Parse(remote)
	if err != nil {
		return fmt.Errorf("parsing remote failed: %w", err)
	}
	info, err := fs.Find(parsed.Name)
	if err != nil {
		return fmt.Errorf("cannot find remote type %q: %w", parsed.Name, err)
	}

	// Setup the remote virtual filesystem
	ctx, cancel := context.WithCancel(context.Background())
	rootfs, err := info.NewFs(ctx, parsed.Name, parsed.Path, fs.ConfigMap(info, parsed.Name, parsed.Config))
	if err != nil {
		cancel()
		return fmt.Errorf("creating remote failed: %w", err)
	}
	f.fscancel = cancel
	f.root = vfs.New(rootfs, &f.vfsopts)

	// Force connection to make sure we actually can connect
	if _, err := f.root.Fs().List(ctx, "/"); err != nil {
		return err
	}
	total, used, free := f.root.Statfs()
	f.Log.Debugf("Connected to %s with %s total, %s used and %s free!",
		f.root.Fs().String(),
		humanize.Bytes(uint64(total)),
		humanize.Bytes(uint64(used)),
		humanize.Bytes(uint64(free)),
	)

	return nil
}

func (f *File) Close() error {
	// Gracefully shutting down the root VFS
	if f.root != nil {
		f.root.FlushDirCache()
		f.root.WaitForWriters(time.Duration(f.FinalWriteTimeout))
		f.root.Shutdown()
		if err := f.root.CleanUp(); err != nil {
			f.Log.Errorf("Cleaning up vfs failed: %v", err)
		}
		f.root = nil
	}

	if f.fscancel != nil {
		f.fscancel()
		f.fscancel = nil
	}

	return nil
}

func (f *File) Write(metrics []telegraf.Metric) error {
	var buf bytes.Buffer

	// Group the metrics per output file
	groups := make(map[string][]telegraf.Metric)
	for _, m := range metrics {
		for _, tmpl := range f.templates {
			buf.Reset()
			if err := tmpl.Execute(&buf, m); err != nil {
				f.Log.Errorf("Cannot create filename %q for metric %v: %v", tmpl.Name(), m, err)
				continue
			}
			fn := buf.String()
			groups[fn] = append(groups[fn], m)
		}
	}

	// Serialize the metric groups
	groupBuffer := make(map[string][]byte, len(groups))
	for fn, fnMetrics := range groups {
		if f.UseBatchFormat {
			serialized, err := f.serializer.SerializeBatch(fnMetrics)
			if err != nil {
				f.Log.Errorf("Could not serialize metrics: %v", err)
				continue
			}
			groupBuffer[fn] = serialized
		} else {
			for _, m := range fnMetrics {
				serialized, err := f.serializer.Serialize(m)
				if err != nil {
					f.Log.Debugf("Could not serialize metric: %v", err)
					continue
				}
				groupBuffer[fn] = append(groupBuffer[fn], serialized...)
			}
		}
	}

	// Write the files
	for fn, serialized := range groupBuffer {
		// Make sure the directory exists
		dir := filepath.Dir(filepath.ToSlash(fn))
		if dir != "." && dir != "/" {
			// Make sure we keep the original path-separators
			if filepath.ToSlash(fn) != fn {
				dir = filepath.FromSlash(dir)
			}
			if err := f.root.MkdirAll(dir, f.root.Opt.DirPerms); err != nil {
				return fmt.Errorf("creating dir %q failed: %w", dir, err)
			}
		}

		// Open the file for appending or create a new one
		file, err := f.root.OpenFile(fn, os.O_APPEND|os.O_RDWR|os.O_CREATE, f.root.Opt.FilePerms)
		if err != nil {
			return fmt.Errorf("opening file %q: %w", fn, err)
		}

		// Write the data
		if _, err := file.Write(serialized); err != nil {
			file.Close()
			return fmt.Errorf("writing metrics to file %q failed: %w", fn, err)
		}
		file.Close()
	}

	return nil
}

func init() {
	outputs.Add("remotefile", func() telegraf.Output { return &File{} })
}
