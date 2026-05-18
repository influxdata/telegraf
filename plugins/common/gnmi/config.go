package gnmi

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/common/yangmodel"
)

type Option func(*Handler)

func WithDefaultName(name string) Option {
	return func(h *Handler) {
		h.defaultName = name
	}
}

func WithEnforceFirstNamespaceAsOrigin() Option {
	return func(h *Handler) {
		h.EnforceFirstNamespaceAsOrigin = true
	}
}

type HandlerConfig struct {
	EmitDeleteMetrics   bool     `toml:"emit_delete_metrics"`
	CanonicalFieldNames bool     `toml:"canonical_field_names"`
	TrimSlash           bool     `toml:"trim_field_names"`
	TagPathPrefix       bool     `toml:"prefix_tag_key_with_path"`
	GuessPathStrategy   string   `toml:"path_guessing_strategy"`
	VendorExt           []string `toml:"vendor_specific"`
	YangModelPaths      []string `toml:"yang_model_paths"`
}

func (cfg *HandlerConfig) Handler(log telegraf.Logger, options ...Option) (*Handler, error) {
	// Check vendor_specific options configured by user
	if err := choice.CheckSlice(cfg.VendorExt, supportedExtensions); err != nil {
		return nil, fmt.Errorf("unsupported vendor_specific option: %w", err)
	}

	// Check path guessing and handle deprecated option
	switch cfg.GuessPathStrategy {
	case "", "none", "common path", "subscription":
	default:
		return nil, fmt.Errorf("invalid 'path_guessing_strategy' %q", cfg.GuessPathStrategy)
	}

	// Load the YANG models if specified by the user and check for errors early.
	// We redo this later to actually use the decode.
	if len(cfg.YangModelPaths) > 0 {
		if _, err := yangmodel.NewDecoder(cfg.YangModelPaths...); err != nil {
			return nil, fmt.Errorf("invalid YANG model decoder: %w", err)
		}
	}

	h := &Handler{
		HandlerConfig: cfg,
		aliases:       make(map[*pathInfo]string),
		tagStore:      newTagStore(),
		log:           log,
	}

	for _, o := range options {
		o(h)
	}

	// Load the YANG models if specified by the user
	if len(cfg.YangModelPaths) > 0 {
		decoder, err := yangmodel.NewDecoder(cfg.YangModelPaths...)
		if err != nil {
			return nil, fmt.Errorf("creating YANG model decoder failed: %w", err)
		}
		h.decoder = decoder
	}
	return h, nil
}
