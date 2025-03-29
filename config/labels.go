package config

import (
	"fmt"
	"log"
	"strings"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/validation"
)

var (
	// LabelFlags holds the labels for the running instance of telegraf
	// These are user labels passed via flags
	LabelFlags []string

	label labels.Labels
)

func SetupLabels() error {
	lbls := labels.Set{}

	for _, label := range LabelFlags {
		entry := strings.SplitN(label, ":", 2)
		if len(entry) < 2 {
			return fmt.Errorf("invalid label format %v", entry)
		}
		if errs := validation.IsDNS1123Label(entry[0]); len(errs) > 0 {
			return fmt.Errorf("invalid label key %s, %s", entry[0], strings.Join(errs, " "))
		}
		if errs := validation.IsValidLabelValue(entry[1]); len(errs) > 0 {
			return fmt.Errorf("invalid label value %s, %s", entry[1], strings.Join(errs, " "))
		}
		lbls[entry[0]] = entry[1]
	}
	label = lbls
	log.Print("I! Telegraf configured with labels: ", lbls.String())
	return nil
}
