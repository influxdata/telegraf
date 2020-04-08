package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/amenzhinsky/iothub/cmd/internal"
	"github.com/amenzhinsky/iothub/eventhub"
	"github.com/amenzhinsky/iothub/iotservice"
	"github.com/amenzhinsky/iothub/logger"
)

// globally accessible by command handlers, is it a good idea?
var (
	// common
	formatFlag   string
	logLevelFlag = logger.LevelWarn

	// send
	uidFlag             string
	midFlag             string
	cidFlag             string
	expFlag             time.Duration
	ackFlag             iotservice.AckType
	connectTimeoutFlag  uint
	responseTimeoutFlag uint

	// create/update device
	sasPrimaryFlag    string
	sasSecondaryFlag  string
	x509PrimaryFlag   string
	x509SecondaryFlag string
	caFlag            bool
	statusFlag        iotservice.DeviceStatus
	statusReasonFlag  string
	capabilitiesFlag  map[string]interface{}
	forceFlag         bool
	edgeFlag          bool

	// send
	propsFlag map[string]string

	// sas and connection string
	secondaryFlag bool

	// sas
	uriFlag      string
	durationFlag time.Duration

	// watch events
	ehcsFlag string
	ehcgFlag string

	// query
	pageSizeFlag uint

	// twins
	tagsFlag      map[string]interface{}
	twinPropsFlag map[string]interface{}

	// modules
	managedByFlag string

	// configuration
	schemaVersionFlag   string
	priorityFlag        uint
	labelsFlag          map[string]string
	targetConditionFlag string
	modulesContentFlag  map[string]interface{}
	devicesContentFlag  map[string]interface{}
	metricsFlag         map[string]string

	// export
	excludeKeysFlag bool

	// schedule jobs
	jobIDFlag       string
	queryFlag       string
	startTimeFlag   time.Time
	maxExecTimeFlag uint
	timeoutFlag     uint

	jobTypeFlag   iotservice.JobV2Type
	jobStatusFlag iotservice.JobV2Status

	// deployments
	envFlag map[string]interface{}

	// https://docs.docker.com/engine/api/v1.30/#operation/ContainerCreate
	createOptionsFlag map[string]interface{}
)

func main() {
	if err := run(); err != nil {
		if err != internal.ErrInvalidUsage {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
		}
		os.Exit(1)
	}
}

const help = `Helps with interacting and managing your iothub devices. 
The $IOTHUB_SERVICE_CONNECTION_STRING environment variable is required for authentication.`

func run() error {
	ctx := context.Background()
	return internal.New(help, func(f *flag.FlagSet) {
		f.StringVar(&formatFlag, "format", "json-pretty", "data output format <json|json-pretty>")
		f.Var((*internal.LogLevelFlag)(&logLevelFlag), "log-level", "log `level` <error|warn|info|debug>")
	}, []*internal.Command{
		{
			Name:    "send",
			Args:    []string{"DEVICE", "PAYLOAD"},
			Desc:    "send cloud-to-device message",
			Handler: wrap(ctx, send),
			ParseFunc: func(f *flag.FlagSet) {
				f.StringVar((*string)(&ackFlag), "ack", "", "type of ack feedback <none|positive|negative|full>")
				f.StringVar(&uidFlag, "uid", "golang-iothub", "origin of the message")
				f.StringVar(&midFlag, "mid", "", "identifier for the message")
				f.StringVar(&cidFlag, "cid", "", "message identifier in a request-reply")
				f.DurationVar(&expFlag, "exp", 0, "message lifetime")
				f.Var((*internal.StringsMapFlag)(&propsFlag), "prop", "custom property, key=value")
			},
		},
		{
			Name:    "watch-events",
			Desc:    "subscribe to cloud-to-device messages",
			Handler: wrap(ctx, watchEvents),
			ParseFunc: func(f *flag.FlagSet) {
				f.StringVar(&ehcsFlag, "ehcs", "", "custom eventhub connection string")
				f.StringVar(&ehcgFlag, "ehcg", "$Default", "eventhub consumer group")
			},
		},
		{
			Name:    "watch-feedback",
			Desc:    "subscribe to message delivery feedback",
			Handler: wrap(ctx, watchFeedback),
		},
		{
			Name:    "watch-file-notifications",
			Desc:    "subscribe to file upload notifications",
			Handler: wrap(ctx, watchFileNotifications),
		},
		{
			Name:    "call",
			Args:    []string{"DEVICE", "METHOD", "PAYLOAD"},
			Desc:    "call a direct method on the named device",
			Handler: wrap(ctx, callDevice),
			ParseFunc: func(f *flag.FlagSet) {
				f.UintVar(&connectTimeoutFlag, "connect-timeout", 0, "connect timeout in seconds")
				f.UintVar(&responseTimeoutFlag, "response-timeout", 30, "response timeout in seconds")
			},
		},
		{
			Name:    "device",
			Args:    []string{"DEVICE"},
			Desc:    "get device information",
			Handler: wrap(ctx, getDevice),
		},
		{
			Name:    "devices",
			Desc:    "list all available devices",
			Handler: wrap(ctx, listDevices),
		},
		{
			Name:    "create-device",
			Args:    []string{"DEVICE"},
			Desc:    "request an existing device identity",
			Handler: wrap(ctx, createDevice),
			ParseFunc: func(f *flag.FlagSet) {
				f.StringVar(&sasPrimaryFlag, "primary-key", "", "primary key (base64)")
				f.StringVar(&sasSecondaryFlag, "secondary-key", "", "secondary key (base64)")
				f.StringVar(&x509PrimaryFlag, "primary-thumbprint", "", "x509 primary thumbprint")
				f.StringVar(&x509SecondaryFlag, "secondary-thumbprint", "", "x509 secondary thumbprint")
				f.BoolVar(&caFlag, "ca", false, "use certificate authority authentication")
				f.StringVar((*string)(&statusFlag), "status", "", "device status")
				f.StringVar(&statusReasonFlag, "status-reason", "", "disabled device status reason")
				f.Var((*internal.JSONMapFlag)(&capabilitiesFlag), "capability", "device capability, key=value")
				f.BoolVar(&edgeFlag, "edge", false, "create an IoT Edge device (same as -capability=iotEdge=true)")
			},
		},
		{
			Name:    "update-device",
			Args:    []string{"DEVICE"},
			Desc:    "update the named device",
			Handler: wrap(ctx, updateDevice),
			ParseFunc: func(f *flag.FlagSet) {
				f.StringVar(&sasPrimaryFlag, "sas-primary", "", "SAS primary key (base64)")
				f.StringVar(&sasSecondaryFlag, "sas-secondary-key", "", "SAS secondary key (base64)")
				f.StringVar(&x509PrimaryFlag, "x509-primary", "", "x509 primary thumbprint")
				f.StringVar(&x509SecondaryFlag, "x509-secondary", "", "x509 secondary thumbprint")
				f.BoolVar(&caFlag, "ca", false, "use certificate authority authentication")
				f.StringVar((*string)(&statusFlag), "status", "", "device status")
				f.StringVar(&statusReasonFlag, "status-reason", "", "disabled device status reason")
				f.Var((*internal.JSONMapFlag)(&capabilitiesFlag), "capability", "device capability, key=value")
				f.BoolVar(&forceFlag, "force", false, "force update")
			},
		},
		{
			Name:    "delete-device",
			Args:    []string{"DEVICE"},
			Desc:    "delete the named device from the registry",
			Handler: wrap(ctx, deleteDevice),
			ParseFunc: func(f *flag.FlagSet) {
				f.BoolVar(&forceFlag, "force", false, "force update")
			},
		},
		{
			Name:    "call-module",
			Args:    []string{"DEVICE", "MODULE", "METHOD", "PAYLOAD"},
			Desc:    "call a direct method on the named module",
			Handler: wrap(ctx, callModule),
			ParseFunc: func(f *flag.FlagSet) {
				f.UintVar(&connectTimeoutFlag, "connect-timeout", 0, "connect timeout in seconds")
				f.UintVar(&responseTimeoutFlag, "response-timeout", 30, "response timeout in seconds")
			},
		},
		{
			Name:    "modules",
			Args:    []string{"DEVICE"},
			Desc:    "list the named device's modules",
			Handler: wrap(ctx, listModules),
		},
		{
			Name:    "create-module",
			Args:    []string{"DEVICE", "MODULE"},
			Desc:    "add the given module to the registry",
			Handler: wrap(ctx, createModule),
			ParseFunc: func(f *flag.FlagSet) {
				f.StringVar(&sasPrimaryFlag, "sas-primary", "", "SAS primary key (base64)")
				f.StringVar(&sasSecondaryFlag, "sas-secondary-key", "", "SAS secondary key (base64)")
				f.StringVar(&x509PrimaryFlag, "x509-primary", "", "x509 primary thumbprint")
				f.StringVar(&x509SecondaryFlag, "x509-secondary", "", "x509 secondary thumbprint")
				f.BoolVar(&caFlag, "ca", false, "use certificate authority authentication")
				f.StringVar(&managedByFlag, "managed-by", "", "module's owner")
			},
		},
		{
			Name:    "module",
			Args:    []string{"DEVICE", "MODULE"},
			Desc:    "get info of the named module",
			Handler: wrap(ctx, getModule),
		},
		{
			Name:    "update-module",
			Args:    []string{"DEVICE", "MODULE"},
			Desc:    "update the named module",
			Handler: wrap(ctx, updateModule),
			ParseFunc: func(f *flag.FlagSet) {
				f.StringVar(&sasPrimaryFlag, "sas-primary", "", "SAS primary key (base64)")
				f.StringVar(&sasSecondaryFlag, "sas-secondary-key", "", "SAS secondary key (base64)")
				f.StringVar(&x509PrimaryFlag, "x509-primary", "", "x509 primary thumbprint")
				f.StringVar(&x509SecondaryFlag, "x509-secondary", "", "x509 secondary thumbprint")
				f.BoolVar(&caFlag, "ca", false, "use certificate authority authentication")
				f.BoolVar(&forceFlag, "force", false, "force update")
				f.StringVar(&managedByFlag, "managed-by", "", "module's owner")
			},
		},
		{
			Name:    "delete-module",
			Args:    []string{"DEVICE", "MODULE"},
			Desc:    "remove the named module from the registry",
			Handler: wrap(ctx, deleteModule),
			ParseFunc: func(f *flag.FlagSet) {
				f.BoolVar(&forceFlag, "force", false, "force update")
			},
		},
		{
			Name:    "twin",
			Args:    []string{"DEVICE"},
			Desc:    "inspect the named twin device",
			Handler: wrap(ctx, getDeviceTwin),
		},
		{
			Name:    "module-twin",
			Args:    []string{"DEVICE", "MODULE"},
			Desc:    "get the named module twin",
			Handler: wrap(ctx, getModuleTwin),
		},
		{
			Name:    "update-twin",
			Args:    []string{"DEVICE"},
			Desc:    "update the named twin device",
			Handler: wrap(ctx, updateDeviceTwin),
			ParseFunc: func(f *flag.FlagSet) {
				f.Var((*internal.JSONMapFlag)(&twinPropsFlag), "prop", "property to update, key=value")
				f.Var((*internal.JSONMapFlag)(&tagsFlag), "tag", "custom tag, key=value")
			},
		},
		{
			Name:    "update-module-twin",
			Args:    []string{"DEVICE", "MODULE"},
			Desc:    "update the named module twin",
			Handler: wrap(ctx, updateModuleTwin),
			ParseFunc: func(f *flag.FlagSet) {
				f.Var((*internal.JSONMapFlag)(&twinPropsFlag), "prop", "property to update, key=value")
				f.BoolVar(&forceFlag, "force", false, "force update")
			},
		},
		{
			Name:    "configurations",
			Desc:    "list all configurations",
			Handler: wrap(ctx, listConfigurations),
		},
		{
			Name:    "create-configuration",
			Args:    []string{"CONFIGURATION"},
			Desc:    "add a configuration to the registry",
			Handler: wrap(ctx, createConfiguration),
			ParseFunc: func(f *flag.FlagSet) {
				f.UintVar(&priorityFlag, "priority", 10, "priority to resolve configuration conflicts")
				f.StringVar(&schemaVersionFlag, "schema-version", "1.0", "configuration schema version")
				f.Var((*internal.StringsMapFlag)(&labelsFlag), "label", "specific label, key=value")
				f.StringVar(&targetConditionFlag, "target-condition", "*", "target condition")
				f.Var((*internal.StringsMapFlag)(&metricsFlag), "metric", "metric name and query, key=value")
				f.Var((*internal.JSONMapFlag)(&devicesContentFlag), "device-prop", "device property, key=value")
			},
		},
		{
			Name:    "configuration",
			Args:    []string{"CONFIGURATION"},
			Desc:    "retrieve the named configuration",
			Handler: wrap(ctx, getConfiguration),
		},
		{
			Name:    "update-configuration",
			Args:    []string{"CONFIGURATION"},
			Desc:    "update the named configuration",
			Handler: wrap(ctx, updateConfiguration),
			ParseFunc: func(f *flag.FlagSet) {
				f.UintVar(&priorityFlag, "priority", 0, "priority to resolve configuration conflicts")
				f.StringVar(&schemaVersionFlag, "schema-version", "", "configuration schema version")
				f.Var((*internal.StringsMapFlag)(&labelsFlag), "label", "specific labels in key=value format")
				f.StringVar(&targetConditionFlag, "target-condition", "*", "target condition")
				f.Var((*internal.StringsMapFlag)(&metricsFlag), "metric", "metric name and query, key=value")
				f.Var((*internal.JSONMapFlag)(&devicesContentFlag), "device-prop", "device property, key=value")
				f.BoolVar(&forceFlag, "force", false, "force update")
			},
		},
		{
			Name:    "delete-configuration",
			Args:    []string{"CONFIGURATION"},
			Desc:    "delete the named configuration by id",
			Handler: wrap(ctx, deleteConfiguration),
			ParseFunc: func(f *flag.FlagSet) {
				f.BoolVar(&forceFlag, "force", false, "force update")
			},
		},
		{
			Name:    "apply-configuration",
			Args:    []string{"DEVICE"},
			Desc:    "applies configuration on the named device",
			Handler: wrap(ctx, applyConfiguration),
			ParseFunc: func(f *flag.FlagSet) {
				f.Var((*internal.JSONMapFlag)(&devicesContentFlag), "device-prop", "device property, key=value")
				f.Var((*internal.JSONMapFlag)(&modulesContentFlag), "module-prop", "module property, key=value")
			},
		},
		{
			Name:    "deployments",
			Args:    []string{},
			Desc:    "list all IoT Edge deployments (configurations)",
			Handler: wrap(ctx, listDeployments),
		},
		{
			Name:    "create-deployment",
			Args:    []string{"DEPLOYMENT", "MODULE", "IMAGE"},
			Desc:    "create an IoT Edge deployment",
			Handler: wrap(ctx, createDeployment),
			ParseFunc: func(f *flag.FlagSet) {
				f.UintVar(&priorityFlag, "priority", 10, "priority to resolve configuration conflicts")
				f.StringVar(&schemaVersionFlag, "schema-version", "1.0", "configuration schema version")
				f.Var((*internal.StringsMapFlag)(&labelsFlag), "label", "specific label, key=value")
				f.StringVar(&targetConditionFlag, "target-condition", "*", "target condition")
				f.Var((*internal.StringsMapFlag)(&metricsFlag), "metric", "metric name and query, key=value")
				f.Var((*internal.JSONMapFlag)(&modulesContentFlag), "module-prop", "module property, key=value")
				f.Var((*internal.JSONMapFlag)(&envFlag), "env", "container environment, key=value")
				f.Var((*internal.JSONMapFlag)(&createOptionsFlag), "create-options", "container create options, key=value")
			},
		},
		{
			Name:    "query",
			Args:    []string{"SQL"},
			Desc:    "execute sql query on devices",
			Handler: wrap(ctx, query),
			ParseFunc: func(f *flag.FlagSet) {
				f.UintVar(&pageSizeFlag, "page-size", 0, "number of records per request")
			},
		},
		{
			Name:    "statistics",
			Desc:    "get statistics the registry statistics",
			Handler: wrap(ctx, statistics),
		},
		{
			Name:    "import",
			Desc:    "import devices from a blob",
			Args:    []string{"INPUT", "OUTPUT"},
			Handler: wrap(ctx, importFromBlob),
		},
		{
			Name:    "export",
			Desc:    "export devices to a blob",
			Args:    []string{"OUTPUT"},
			Handler: wrap(ctx, exportToBlob),
			ParseFunc: func(f *flag.FlagSet) {
				f.BoolVar(&excludeKeysFlag, "exclude-keys", false, "exclude keys in the export blob file")
			},
		},
		{
			Name:    "jobs",
			Desc:    "list the last import/export jobs",
			Handler: wrap(ctx, listJobs),
		},
		{
			Name:    "job",
			Args:    []string{"JOB"},
			Desc:    "get the status of a import/export job",
			Handler: wrap(ctx, getJob),
		},
		{
			Name:    "cancel-job",
			Desc:    "cancel a import/export job",
			Handler: wrap(ctx, cancelJob),
		},
		{
			Name:    "schedule-jobs",
			Desc:    "list all scheduled jobs",
			Handler: wrap(ctx, listScheduleJobs),
			ParseFunc: func(f *flag.FlagSet) {
				f.StringVar((*string)(&jobTypeFlag), "type", "",
					"job type <scheduleUpdateTwin|scheduleDeviceMethod>")
				f.StringVar((*string)(&jobStatusFlag), "status", "",
					"job status <queued|scheduled|running|cancelled|completed>")
			},
		},
		{
			Name:    "get-schedule-job",
			Args:    []string{"JOB"},
			Desc:    "retrieve the named job information from the registry",
			Handler: wrap(ctx, getScheduleJob),
		},
		{
			Name:    "cancel-schedule-job",
			Args:    []string{"JOB"},
			Desc:    "cancel the named job",
			Handler: wrap(ctx, cancelScheduleJob),
		},
		{
			Name:    "schedule-method-call",
			Args:    []string{"METHOD", "PAYLOAD"},
			Handler: wrap(ctx, scheduleMethodCall),
			ParseFunc: func(f *flag.FlagSet) {
				f.StringVar(&jobIDFlag, "job-id", "", "unique job id")
				f.StringVar(&queryFlag, "query", "*", "query condition")
				f.Var((*internal.TimeFlag)(&startTimeFlag), "start-time", "start time in RFC3339")
				f.UintVar(&timeoutFlag, "connect-timeout", 0, "connection timeout in seconds")
				f.UintVar(&maxExecTimeFlag, "exec-timeout", 30, "maximal execution time in seconds")
			},
		},
		{
			Name:    "schedule-twin-update",
			Args:    []string{}, // TODO
			Handler: wrap(ctx, scheduleTwinUpdate),
			ParseFunc: func(f *flag.FlagSet) {
				f.StringVar(&jobIDFlag, "job-id", "", "unique job id")
				f.StringVar(&queryFlag, "query", "*", "query condition")
				f.Var((*internal.TimeFlag)(&startTimeFlag), "start-time", "start time in RFC3339")
				f.UintVar(&maxExecTimeFlag, "exec-timeout", 30, "maximal execution time in seconds")
			},
		},
		{
			Name:    "device-connection-string",
			Args:    []string{"DEVICE"},
			Desc:    "get a device's connection string",
			Handler: wrap(ctx, deviceConnectionString),
			ParseFunc: func(f *flag.FlagSet) {
				f.BoolVar(&secondaryFlag, "secondary", false, "use the secondary key instead")
			},
		},
		{
			Name:    "module-connection-string",
			Args:    []string{"DEVICE", "MODULE"},
			Desc:    "get a module's connection string",
			Handler: wrap(ctx, moduleConnectionString),
			ParseFunc: func(f *flag.FlagSet) {
				f.BoolVar(&secondaryFlag, "secondary", false, "use the secondary key instead")
			},
		},
		{
			Name:    "access-signature",
			Args:    []string{"DEVICE"},
			Desc:    "generate a SAS token",
			Handler: wrap(ctx, sas),
			ParseFunc: func(f *flag.FlagSet) {
				f.StringVar(&uriFlag, "uri", "", "storage resource uri")
				f.DurationVar(&durationFlag, "duration", time.Hour, "token validity time")
				f.BoolVar(&secondaryFlag, "secondary", false, "use the secondary key instead")
			},
		},
	}).Run(os.Args)
}

func wrap(
	ctx context.Context,
	fn func(context.Context, *iotservice.Client, []string) error,
) internal.HandlerFunc {
	return func(args []string) error {
		c, err := iotservice.NewFromConnectionString(
			os.Getenv("IOTHUB_SERVICE_CONNECTION_STRING"),
			iotservice.WithLogger(
				logger.New(logLevelFlag, nil),
			),
		)
		if err != nil {
			return err
		}
		defer c.Close()

		// handle first SIGINT and try to exit gracefully
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, os.Interrupt)
		go func() {
			<-sigc
			signal.Reset(os.Interrupt)
			close(sigc)
			cancel()
		}()
		if err := fn(ctx, c, args); err != nil {
			select {
			case <-sigc:
				if err == context.Canceled {
					return nil
				}
			default:
			}
			return err
		}
		return nil
	}
}

func getDevice(ctx context.Context, c *iotservice.Client, args []string) error {
	return output(c.GetDevice(ctx, args[0]))
}

func listDevices(ctx context.Context, c *iotservice.Client, args []string) error {
	return output(c.ListDevices(ctx))
}

func createDevice(ctx context.Context, c *iotservice.Client, args []string) error {
	if edgeFlag {
		if capabilitiesFlag == nil {
			capabilitiesFlag = map[string]interface{}{}
		}
		capabilitiesFlag["iotEdge"] = true
	}

	device := &iotservice.Device{
		DeviceID:       args[0],
		Authentication: &iotservice.Authentication{},
		Status:         statusFlag,
		StatusReason:   statusReasonFlag,
		Capabilities:   capabilitiesFlag,
	}
	if err := updateAuth(device.Authentication); err != nil {
		return err
	}
	return output(c.CreateDevice(ctx, device))
}

func updateDevice(ctx context.Context, c *iotservice.Client, args []string) error {
	device, err := c.GetDevice(ctx, args[0])
	if err != nil {
		return err
	}
	if forceFlag {
		device.ETag = ""
	}
	if statusFlag != "" {
		device.Status = statusFlag
	}
	if statusReasonFlag != "" {
		device.StatusReason = statusReasonFlag
	}
	mergeMapJSON(capabilitiesFlag, device.Capabilities)
	if err := updateAuth(device.Authentication); err != nil {
		return err
	}
	return output(c.UpdateDevice(ctx, device))
}

func updateAuth(auth *iotservice.Authentication) error {
	switch {
	case sasPrimaryFlag != "" || sasSecondaryFlag != "":
		if x509PrimaryFlag != "" || x509SecondaryFlag != "" {
			return errors.New("-x509-* options cannot be used along with sas authentication")
		} else if caFlag {
			return errors.New("-ca option cannot be used along with sas authentication")
		}
		auth.Type = iotservice.AuthSAS
		auth.X509Thumbprint = nil
		auth.SymmetricKey = &iotservice.SymmetricKey{
			PrimaryKey:   sasPrimaryFlag,
			SecondaryKey: sasSecondaryFlag,
		}
	case x509PrimaryFlag != "" || x509SecondaryFlag != "":
		if caFlag {
			return errors.New("-ca option cannot be used along with x509 authentication")
		}
		auth.Type = iotservice.AuthSelfSigned
		auth.SymmetricKey = nil
		auth.X509Thumbprint = &iotservice.X509Thumbprint{
			PrimaryThumbprint:   x509PrimaryFlag,
			SecondaryThumbprint: x509SecondaryFlag,
		}
	case caFlag:
		auth.Type = iotservice.AuthCA
		auth.SymmetricKey = nil
		auth.X509Thumbprint = nil
	}
	return nil
}

func deleteDevice(ctx context.Context, c *iotservice.Client, args []string) error {
	device, err := c.GetDevice(ctx, args[0])
	if err != nil {
		return err
	}
	if forceFlag {
		device.ETag = ""
	}
	return c.DeleteDevice(ctx, device)
}

func listModules(ctx context.Context, c *iotservice.Client, args []string) error {
	return output(c.ListModules(ctx, args[0]))
}

func getModule(ctx context.Context, c *iotservice.Client, args []string) error {
	return output(c.GetModule(ctx, args[0], args[1]))
}

func createModule(ctx context.Context, c *iotservice.Client, args []string) error {
	module := &iotservice.Module{
		DeviceID:       args[0],
		ModuleID:       args[1],
		Authentication: &iotservice.Authentication{},
		ManagedBy:      managedByFlag,
	}
	if err := updateAuth(module.Authentication); err != nil {
		return err
	}
	return output(c.CreateModule(ctx, module))
}

func updateModule(ctx context.Context, c *iotservice.Client, args []string) error {
	module, err := c.GetModule(ctx, args[0], args[1])
	if err != nil {
		return err
	}
	if forceFlag {
		module.ETag = ""
	}
	if managedByFlag != "" {
		module.ManagedBy = managedByFlag
	}
	if err := updateAuth(module.Authentication); err != nil {
		return err
	}
	return output(c.UpdateModule(ctx, module))
}

func deleteModule(ctx context.Context, c *iotservice.Client, args []string) error {
	module, err := c.GetModule(ctx, args[0], args[1])
	if err != nil {
		return err
	}
	if forceFlag {
		module.ETag = ""
	}
	return c.DeleteModule(ctx, module)
}

func listConfigurations(ctx context.Context, c *iotservice.Client, args []string) error {
	return listConfigurationsFiltered(ctx, c, func(cfg *iotservice.Configuration) bool {
		return cfg.Content.DeviceContent != nil
	})
}

func listDeployments(ctx context.Context, c *iotservice.Client, args []string) error {
	return listConfigurationsFiltered(ctx, c, func(cfg *iotservice.Configuration) bool {
		return cfg.Content.ModulesContent != nil
	})
}

func listConfigurationsFiltered(
	ctx context.Context,
	c *iotservice.Client,
	matches func(configuration *iotservice.Configuration) bool,
) error {
	configurations, err := c.ListConfigurations(ctx)
	if err != nil {
		return err
	}
	filtered := make([]*iotservice.Configuration, 0, len(configurations))
	for _, configuration := range configurations {
		if matches(configuration) {
			filtered = append(filtered, configuration)
		}
	}
	return output(filtered, nil)
}

func getConfiguration(ctx context.Context, c *iotservice.Client, args []string) error {
	return output(c.GetConfiguration(ctx, args[0]))
}

func createConfiguration(ctx context.Context, c *iotservice.Client, args []string) error {
	return output(c.CreateConfiguration(ctx, &iotservice.Configuration{
		ID:              args[0],
		SchemaVersion:   schemaVersionFlag,
		Priority:        priorityFlag,
		Labels:          labelsFlag,
		TargetCondition: targetConditionFlag,
		Content: &iotservice.ConfigurationContent{
			DeviceContent: devicesContentFlag,
		},
		Metrics: &iotservice.ConfigurationMetrics{
			Queries: metricsFlag,
		},
	}))
}

// https://github.com/Azure/azure-iot-cli-extension/blob/v0.8.7/azext_iot/assets/edge-deploy-2.0.schema.json
func createDeployment(ctx context.Context, c *iotservice.Client, args []string) error {
	env := make(map[string]interface{}, len(envFlag))
	for k, v := range envFlag {
		env[k] = map[string]interface{}{
			"value": v,
		}
	}
	createOptions, err := json.Marshal(createOptionsFlag)
	if err != nil {
		return err
	}

	return output(c.CreateConfiguration(ctx, &iotservice.Configuration{
		ID:              args[0],
		SchemaVersion:   schemaVersionFlag,
		Priority:        priorityFlag,
		Labels:          labelsFlag,
		TargetCondition: targetConditionFlag,
		Content: &iotservice.ConfigurationContent{
			ModulesContent: map[string]interface{}{
				"$edgeAgent": map[string]interface{}{
					"properties.desired": map[string]interface{}{
						"modules": map[string]interface{}{
							args[1]: map[string]interface{}{
								"type": "docker",
								"settings": map[string]interface{}{
									"image":         args[2],
									"createOptions": string(createOptions),
								},
								"env":           env,
								"status":        "running",
								"restartPolicy": "always",
								"version":       "1.0",
							},
						},

						"runtime": map[string]interface{}{
							"type": "docker",
							"settings": map[string]interface{}{
								"minDockerVersion":    "v1.25",
								"registryCredentials": map[string]interface{}{
									// TODO: "REGISTRYNAME": map[string]interface{}{
									// TODO: 	"address":  "docker.com",
									// TODO: 	"password": "pwd",
									// TODO: 	"username": "test",
									// TODO: },
								},
							},
						},

						"schemaVersion": "1.0",
						"systemModules": map[string]interface{}{
							"edgeAgent": map[string]interface{}{
								"settings": map[string]interface{}{
									"image":         "mcr.microsoft.com/azureiotedge-agent:1.0",
									"createOptions": "",
								},
								"type": "docker",
							},
							"edgeHub": map[string]interface{}{
								"settings": map[string]interface{}{
									"image":         "mcr.microsoft.com/azureiotedge-hub:1.0",
									"createOptions": "{\"HostConfig\":{\"PortBindings\":{\"8883/tcp\":[{\"HostPort\":\"8883\"}],\"5671/tcp\":[{\"HostPort\":\"5671\"}],\"443/tcp\":[{\"HostPort\":\"443\"}]}}}",
								},
								"type":          "docker",
								"status":        "running",
								"restartPolicy": "always",
							},
						},
					},
				},

				"$edgeHub": map[string]interface{}{
					"properties.desired": map[string]interface{}{
						"routes":        map[string]interface{}{},
						"schemaVersion": "1.0",
						"storeAndForwardConfiguration": map[string]interface{}{
							"timeToLiveSecs": 7200,
						},
					},
				},

				// TODO: "testmodulename": map[string]interface{}{
				// TODO: 	"properties.desired.test": map[string]interface{}{
				// TODO: 		"foo": "bar",
				// TODO: 	},
				// TODO: },
			},
		},
		Metrics: &iotservice.ConfigurationMetrics{
			Queries: metricsFlag,
		},
	}))
}

func updateConfiguration(ctx context.Context, c *iotservice.Client, args []string) error {
	config, err := c.GetConfiguration(ctx, args[0])
	if err != nil {
		return err
	}
	if forceFlag {
		config.ETag = ""
	}
	if schemaVersionFlag != "" {
		config.SchemaVersion = schemaVersionFlag
	}
	if priorityFlag != 0 {
		config.Priority = priorityFlag
	}
	mergeMapStrings(config.Labels, labelsFlag)
	mergeMapJSON(config.Content.ModulesContent, modulesContentFlag)
	mergeMapJSON(config.Content.DeviceContent, devicesContentFlag)
	mergeMapStrings(config.Metrics.Queries, metricsFlag)
	if targetConditionFlag != "" {
		config.TargetCondition = targetConditionFlag
	}
	return output(c.UpdateConfiguration(ctx, config))
}

func deleteConfiguration(ctx context.Context, c *iotservice.Client, args []string) error {
	config, err := c.GetConfiguration(ctx, args[0])
	if err != nil {
		return err
	}
	if forceFlag {
		config.ETag = ""
	}
	return c.DeleteConfiguration(ctx, config)
}

func applyConfiguration(ctx context.Context, c *iotservice.Client, args []string) error {
	return c.ApplyConfigurationContentOnDevice(
		ctx,
		args[0],
		&iotservice.ConfigurationContent{
			ModulesContent: modulesContentFlag,
			DeviceContent:  devicesContentFlag,
		},
	)
}

func query(ctx context.Context, c *iotservice.Client, args []string) error {
	return c.QueryDevices(ctx, args[0], func(v map[string]interface{}) error {
		return output(v, nil)
	})
}

func statistics(ctx context.Context, c *iotservice.Client, args []string) error {
	return output(c.Stats(ctx))
}

func importFromBlob(ctx context.Context, c *iotservice.Client, args []string) error {
	return output(c.CreateJob(ctx, &iotservice.Job{
		Type:                   iotservice.JobImport,
		InputBlobContainerURI:  args[0],
		OutputBlobContainerURI: args[1],
	}))
}

func exportToBlob(ctx context.Context, c *iotservice.Client, args []string) error {
	return output(c.CreateJob(ctx, &iotservice.Job{
		Type:                   iotservice.JobExport,
		OutputBlobContainerURI: args[0],
		ExcludeKeysInExport:    excludeKeysFlag,
	}))
}

func getDeviceTwin(ctx context.Context, c *iotservice.Client, args []string) error {
	return output(c.GetDeviceTwin(ctx, args[0]))
}

func getModuleTwin(ctx context.Context, c *iotservice.Client, args []string) error {
	return output(c.GetModuleTwin(ctx, args[0], args[1]))
}

func updateDeviceTwin(ctx context.Context, c *iotservice.Client, args []string) error {
	twin, err := c.GetDeviceTwin(ctx, args[0])
	if err != nil {
		return err
	}
	if forceFlag {
		twin.ETag = ""
	}
	mergeMapJSON(twin.Tags, tagsFlag)
	mergeMapJSON(twin.Properties.Desired, twinPropsFlag)
	return output(c.UpdateDeviceTwin(ctx, twin))
}

func updateModuleTwin(ctx context.Context, c *iotservice.Client, args []string) error {
	twin, err := c.GetModuleTwin(ctx, args[0], args[1])
	if err != nil {
		return err
	}
	if forceFlag {
		twin.ETag = ""
	}
	mergeMapJSON(twin.Properties.Desired, twinPropsFlag)
	return output(c.UpdateModuleTwin(ctx, twin))
}

func callDevice(ctx context.Context, c *iotservice.Client, args []string) error {
	call, err := mkcall(args[1], args[2])
	if err != nil {
		return err
	}
	return output(c.CallDeviceMethod(ctx, args[0], call))
}

func callModule(ctx context.Context, c *iotservice.Client, args []string) error {
	call, err := mkcall(args[2], args[3])
	if err != nil {
		return err
	}
	return output(c.CallModuleMethod(ctx, args[0], args[1], call))
}

func mkcall(method, payload string) (*iotservice.MethodCall, error) {
	var p map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return nil, err
	}
	return &iotservice.MethodCall{
		MethodName:      method,
		ConnectTimeout:  connectTimeoutFlag,
		ResponseTimeout: responseTimeoutFlag,
		Payload:         p,
	}, nil
}

func send(ctx context.Context, c *iotservice.Client, args []string) error {
	expiryTime := time.Time{}
	if expFlag != 0 {
		expiryTime = time.Now().Add(expFlag)
	}
	return c.SendEvent(ctx, args[0], []byte(args[1]),
		iotservice.WithSendMessageID(midFlag),
		iotservice.WithSendAck(ackFlag),
		iotservice.WithSendProperties(propsFlag),
		iotservice.WithSendUserID(uidFlag),
		iotservice.WithSendCorrelationID(cidFlag),
		iotservice.WithSendExpiryTime(expiryTime),
	)
}

func watchEvents(ctx context.Context, c *iotservice.Client, args []string) error {
	if ehcsFlag != "" {
		return watchEventHubEvents(ctx, ehcsFlag, ehcgFlag)
	}
	return c.SubscribeEvents(ctx, func(msg *iotservice.Event) error {
		return output(msg, nil)
	})
}

func watchEventHubEvents(ctx context.Context, cs, group string) error {
	c, err := eventhub.DialConnectionString(cs)
	if err != nil {
		return err
	}
	return c.Subscribe(ctx, func(m *eventhub.Event) error {
		return output(iotservice.FromAMQPMessage(m.Message), nil)
	},
		eventhub.WithSubscribeConsumerGroup(group),
		eventhub.WithSubscribeSince(time.Now()),
	)
}

func watchFeedback(ctx context.Context, c *iotservice.Client, args []string) error {
	return c.SubscribeFeedback(ctx, func(f *iotservice.Feedback) error {
		return output(f, nil)
	})
}

func watchFileNotifications(ctx context.Context, c *iotservice.Client, args []string) error {
	return c.SubscribeFileNotifications(ctx, func(f *iotservice.FileNotification) error {
		return output(f, nil)
	})
}

func listJobs(ctx context.Context, c *iotservice.Client, args []string) error {
	return output(c.ListJobs(ctx))
}

func getJob(ctx context.Context, c *iotservice.Client, args []string) error {
	return output(c.GetJob(ctx, args[0]))
}

func cancelJob(ctx context.Context, c *iotservice.Client, args []string) error {
	return output(c.CancelJob(ctx, args[0]))
}

func listScheduleJobs(ctx context.Context, c *iotservice.Client, args []string) error {
	return c.QueryJobsV2(ctx, &iotservice.JobV2Query{
		Type:     jobTypeFlag,
		Status:   jobStatusFlag,
		PageSize: pageSizeFlag,
	}, func(job *iotservice.JobV2) error {
		return output(job, nil)
	})
}

func getScheduleJob(ctx context.Context, c *iotservice.Client, args []string) error {
	return output(c.GetJobV2(ctx, args[0]))
}

func cancelScheduleJob(ctx context.Context, c *iotservice.Client, args []string) error {
	return output(c.CancelJobV2(ctx, args[0]))
}

func scheduleMethodCall(ctx context.Context, c *iotservice.Client, args []string) error {
	if jobIDFlag == "" {
		jobIDFlag = genID()
	}
	var payload interface{}
	if err := json.Unmarshal([]byte(args[1]), &payload); err != nil {
		return err
	}
	return output(c.CreateJobV2(ctx, &iotservice.JobV2{
		JobID: jobIDFlag,
		Type:  iotservice.JobTypeDeviceMethod,
		CloudToDeviceMethod: &iotservice.DeviceMethodParams{
			MethodName:       args[0],
			Payload:          payload,
			TimeoutInSeconds: timeoutFlag,
		},
		QueryCondition:            queryFlag,
		StartTime:                 startTimeFlag,
		MaxExecutionTimeInSeconds: maxExecTimeFlag,
	}))
}

func scheduleTwinUpdate(ctx context.Context, c *iotservice.Client, args []string) error {
	if jobIDFlag == "" {
		jobIDFlag = genID()
	}
	return output(c.CreateJobV2(ctx, &iotservice.JobV2{
		JobID: jobIDFlag,
		Type:  iotservice.JobTypeUpdateTwin,

		// TODO: should be the same as updateTwin action, query is not applied here
		UpdateTwin: map[string]interface{}{
			"etag":     "*",
			"deviceId": "",
			"tags":     map[string]interface{}{},
			"properties": map[string]interface{}{
				"desired": map[string]interface{}{
					"scheduled": 1,
				},
			},
		},
		QueryCondition:            queryFlag,
		StartTime:                 startTimeFlag,
		MaxExecutionTimeInSeconds: maxExecTimeFlag,
	}))
}

func genID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func deviceConnectionString(ctx context.Context, c *iotservice.Client, args []string) error {
	device, err := c.GetDevice(ctx, args[0])
	if err != nil {
		return err
	}
	return output(c.DeviceConnectionString(device, secondaryFlag))
}

func moduleConnectionString(ctx context.Context, c *iotservice.Client, args []string) error {
	module, err := c.GetModule(ctx, args[0], args[1])
	if err != nil {
		return err
	}
	return output(c.ModuleConnectionString(module, secondaryFlag))
}

func sas(ctx context.Context, c *iotservice.Client, args []string) error {
	device, err := c.GetDevice(ctx, args[0])
	if err != nil {
		return err
	}
	return output(c.DeviceSAS(device, "", durationFlag, secondaryFlag))
}

func output(v interface{}, err error) error {
	if err != nil {
		return err
	}
	if s, ok := v.(string); ok {
		return internal.OutputLine(s)
	}
	return internal.Output(v, formatFlag)
}

func mergeMapStrings(src, changes map[string]string) {
	for k, v := range changes {
		src[k] = v
	}
}

func mergeMapJSON(src, changes map[string]interface{}) {
	for k, v := range changes {
		src[k] = v
	}
}
