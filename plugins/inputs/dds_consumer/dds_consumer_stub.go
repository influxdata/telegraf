//go:build !dds

/*****************************************************************************
*   (c) 2005-2015 Copyright, Real-Time Innovations.  All rights reserved.    *
*                                                                            *
* No duplications, whole or partial, manual or electronic, may be made       *
* without express written permission.  Any such copies, or revisions thereof,*
* must display this notice unaltered.                                        *
* This code contains trade secrets of Real-Time Innovations, Inc.            *
*                                                                            *
*****************************************************************************/

package dds_consumer

import (
	"errors"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Stub implementation for testing without RTI DDS dependency
type DDSConsumer struct {
	ConfigFilePath    string   `toml:"config_path"`
	ParticipantConfig string   `toml:"participant_config"`
	ReaderConfig      string   `toml:"reader_config"`
	TagKeys           []string `toml:"tag_keys"`
}

var sampleConfig = `
  ## XML configuration file path
  config_path = "example_configs/ShapeExample.xml"

  ## Configuration name for DDS Participant from a description in XML
  participant_config = "MyParticipantLibrary::Zero"

  ## Configuration name for DDS DataReader from a description in XML
  reader_config = "MySubscriber::MySquareReader"

  # Tag key is an array of keys that should be added as tags.
  tag_keys = ["color"]

  # Override the base name of the measurement
  name_override = "shapes"

  ## Data format to consume.
  data_format = "json"
`

func (d *DDSConsumer) SampleConfig() string {
	return sampleConfig
}

func (d *DDSConsumer) Description() string {
	return "Read metrics from DDS (requires RTI Connext DDS - build with -tags dds)"
}

func (d *DDSConsumer) Start(acc telegraf.Accumulator) error {
	return errors.New("DDS Consumer plugin requires RTI Connext DDS. Build with -tags dds and install RTI dependencies")
}

func (d *DDSConsumer) Stop() {
	// No-op for stub
}

func (d *DDSConsumer) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("dds_consumer", func() telegraf.Input {
		return &DDSConsumer{}
	})
}
