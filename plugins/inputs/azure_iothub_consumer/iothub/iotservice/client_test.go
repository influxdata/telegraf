package iotservice

import (
	"context"
	"io"
	"os"
	"testing"
	"time"
)

func TestSendWithNegativeFeedback(t *testing.T) {
	client := newClient(t)
	device := newDevice(t, client)

	mid := genID()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errc := make(chan error, 1)
	go func() {
		errc <- client.SubscribeFeedback(ctx, func(f *Feedback) error {
			if f.OriginalMessageID == mid {
				cancel()
			}
			return nil
		})
	}()

	// send a message to the previously created device that's not connected
	if err := client.SendEvent(
		ctx,
		device.DeviceID,
		nil,
		WithSendAck(AckNegative),
		WithSendMessageID(mid),
		WithSendExpiryTime(time.Now()),
	); err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-errc:
		if err != context.Canceled {
			t.Fatal(err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("timed out")
	}
}

func TestETags(t *testing.T) {
	client := newClient(t)
	device := newDevice(t, client)

	// invalid ETag
	etag := device.ETag
	device.ETag = "fake"
	if _, err := client.UpdateDevice(context.Background(), device); err == nil {
		t.Fatal("expected an error with an invalid etag")
	}

	// valid ETag
	device.ETag = etag
	if _, err := client.UpdateDevice(context.Background(), device); err != nil {
		t.Fatal(err)
	}

	// force update with If-Match = *
	device.ETag = ""
	if _, err := client.UpdateDevice(context.Background(), device); err != nil {
		t.Fatal(err)
	}
}

func TestBulkOperations(t *testing.T) {
	client := newClient(t)
	devices := []*Device{
		{DeviceID: "test-bulk-0"},
		{DeviceID: "test-bulk-1"},
	}
	for _, dev := range devices {
		_ = client.DeleteDevice(context.Background(), dev)
	}

	res, err := client.CreateDevices(context.Background(), devices)
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsSuccessful {
		t.Fatal("create is not successful")
	}

	res, err = client.UpdateDevices(context.Background(), devices, false)
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsSuccessful {
		t.Fatal("update is not successful")
	}

	res, err = client.DeleteDevices(context.Background(), devices, false)
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsSuccessful {
		t.Fatal("delete is not successful")
	}
}

func TestBulkErrors(t *testing.T) {
	client := newClient(t)
	devices := []*Device{
		{DeviceID: "test-bulk-0"},
		{DeviceID: "test-bulk-1"},
	}
	for _, dev := range devices {
		_ = client.DeleteDevice(context.Background(), dev)
	}
	if _, err := client.CreateDevice(context.Background(), devices[0]); err != nil {
		t.Fatal(err)
	}
	res, err := client.CreateDevices(context.Background(), devices)
	if err != nil {
		t.Fatal(err)
	}

	if res.IsSuccessful {
		t.Errorf("IsSuccessful = true, want false")
	}
	if len(res.Errors) != 1 {
		t.Errorf("no errors returned")
	}
}

func TestRegistryError(t *testing.T) {
	client := newClient(t)
	_, err := client.CreateDevice(context.Background(), &Device{DeviceID: "!@#$%^&"})
	re, ok := err.(*BadRequestError)
	if !ok {
		t.Fatalf("expected a registry error, got = %v", err)
	}
	if re.Message == "" || re.ExceptionMessage == "" {
		t.Fatal("message is empty")
	}
}

func TestListDevices(t *testing.T) {
	client := newClient(t)
	device := newDevice(t, client)
	devices, err := client.ListDevices(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, dev := range devices {
		if dev.DeviceID == device.DeviceID {
			return
		}
	}
	t.Fatal("device not found", device)
}

func TestGetDevice(t *testing.T) {
	client := newClient(t)
	device := newDevice(t, client)
	if _, err := client.GetDevice(context.Background(), device.DeviceID); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateDevice(t *testing.T) {
	client := newClient(t)
	device := newDevice(t, client)
	device.Status = Disabled
	dev, err := client.UpdateDevice(context.Background(), device)
	if err != nil {
		t.Fatal(err)
	}
	if dev.Status != device.Status {
		t.Fatal("device is not updated")
	}
}

func TestDeleteDevice(t *testing.T) {
	client := newClient(t)
	device := newDevice(t, client)
	if err := client.DeleteDevice(context.Background(), device); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetDevice(context.Background(), device.DeviceID); err == nil {
		t.Fatal("device found but should be removed")
	}
}

func TestDeviceConnectionString(t *testing.T) {
	client := newClient(t)
	device := newDevice(t, client)
	if _, err := client.DeviceConnectionString(device, false); err != nil {
		t.Fatal(err)
	}
}

func TestDeviceSAS(t *testing.T) {
	client := newClient(t)
	device := newDevice(t, client)
	sas, err := client.DeviceSAS(device, "", time.Hour, false)
	if err != nil {
		t.Fatal(err)
	}
	if sas == "" {
		t.Fatal("empty sas token")
	}
}

func TestGetDeviceTwin(t *testing.T) {
	client := newClient(t)
	device := newDevice(t, client)
	_, err := client.GetDeviceTwin(context.Background(), device.DeviceID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateDeviceTwin(t *testing.T) {
	client := newClient(t)
	device := newDevice(t, client)
	twin, err := client.UpdateDeviceTwin(context.Background(), &Twin{
		DeviceID: device.DeviceID,
		Properties: &Properties{
			Desired: map[string]interface{}{
				"hw": "1.11",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if twin.Properties.Desired["hw"] != "1.11" {
		t.Fatal("twin not updated")
	}
}

func TestModuleConnectionString(t *testing.T) {
	client := newClient(t)
	_, module := newDeviceAndModule(t, client)
	if _, err := client.ModuleConnectionString(module, false); err != nil {
		t.Fatal(err)
	}
}

func TestListModules(t *testing.T) {
	client := newClient(t)
	device, module := newDeviceAndModule(t, client)
	modules, err := client.ListModules(context.Background(), device.DeviceID)
	if err != nil {
		t.Fatal(err)
	}
	for _, mod := range modules {
		if mod.DeviceID == device.DeviceID && mod.ModuleID == module.ModuleID {
			return
		}
	}
	t.Fatal("module not found", device)
}

func TestGetModule(t *testing.T) {
	client := newClient(t)
	device, module := newDeviceAndModule(t, client)
	if _, err := client.GetModule(
		context.Background(), device.DeviceID, module.ModuleID,
	); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteModule(t *testing.T) {
	client := newClient(t)
	device, module := newDeviceAndModule(t, client)
	if err := client.DeleteModule(context.Background(), module); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetModule(
		context.Background(), device.DeviceID, module.ModuleID,
	); err == nil {
		t.Fatal("module is not deleted")
	}
}

func TestGetModuleTwin(t *testing.T) {
	client := newClient(t)
	device, module := newDeviceAndModule(t, client)
	if _, err := client.GetModuleTwin(
		context.Background(), device.DeviceID, module.ModuleID,
	); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateModuleTwin(t *testing.T) {
	client := newClient(t)
	device, module := newDeviceAndModule(t, client)
	twin, err := client.UpdateModuleTwin(context.Background(), &ModuleTwin{
		DeviceID: device.DeviceID,
		ModuleID: module.ModuleID,
		Properties: &Properties{
			Desired: map[string]interface{}{
				"hw": "1.12",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if twin.Properties.Desired["hw"] != "1.12" {
		t.Fatal("twin not updated")
	}
}

func TestListConfigurations(t *testing.T) {
	client := newClient(t)
	config := newConfiguration(t, client)
	configs, err := client.ListConfigurations(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, cfg := range configs {
		if cfg.ID == config.ID {
			return
		}
	}
	t.Fatal("configuration not found in the list")
}

func TestGetConfiguration(t *testing.T) {
	client := newClient(t)
	config := newConfiguration(t, client)
	if _, err := client.GetConfiguration(context.Background(), config.ID); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateConfiguration(t *testing.T) {
	client := newClient(t)
	config := newConfiguration(t, client)
	config.Labels = map[string]string{
		"foo": "bar",
	}
	if _, err := client.UpdateConfiguration(context.Background(), config); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteConfiguration(t *testing.T) {
	client := newClient(t)
	config := newConfiguration(t, client)
	if err := client.DeleteConfiguration(context.Background(), config); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetConfiguration(context.Background(), config.ID); err == nil {
		t.Fatal("configuration is not deleted")
	}
}

func TestStats(t *testing.T) {
	client := newClient(t)
	if _, err := client.Stats(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestQueryDevices(t *testing.T) {
	client := newClient(t)
	device := newDevice(t, client)

	// some delay needed to wait until the device is available
	time.Sleep(500 * time.Millisecond)

	var found bool
	if err := client.QueryDevices(
		context.Background(),
		"select deviceId from devices",
		func(v map[string]interface{}) error {
			if v["deviceId"].(string) == device.DeviceID {
				found = true
			}
			return nil
		},
	); err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("requested device not found")
	}
}

func TestScheduleMethodCall(t *testing.T) {
	client := newClient(t)
	job, err := client.CreateJobV2(context.Background(), &JobV2{
		JobID: genID(),
		Type:  JobTypeDeviceMethod,
		CloudToDeviceMethod: &DeviceMethodParams{
			MethodName:       "dist-upgrade",
			Payload:          map[string]interface{}{"time": "now"},
			TimeoutInSeconds: 0,
		},
		QueryCondition:            "deviceId='nonexisting'",
		StartTime:                 time.Now().Add(time.Minute),
		MaxExecutionTimeInSeconds: 5,
	})
	if err != nil {
		t.Fatal(err)
	}

	// simply test that job can be found
	job, err = client.GetJobV2(context.Background(), job.JobID)
	if err != nil {
		t.Fatal(err)
	}

	// cancel job immediately because free-tier accounts support only one running job
	job, err = client.CancelJobV2(context.Background(), job.JobID)
	if err != nil {
		t.Fatal(err)
	}
	if job.Status != JobStatusCancelled {
		t.Errorf("job status = %q, want %q", job.Status, JobStatusCancelled)
	}

	// sometimes Azure is being slow
	time.Sleep(time.Second)

	// find just cancelled job
	var found bool
	if err = client.QueryJobsV2(context.Background(), &JobV2Query{
		Type:     JobTypeDeviceMethod,
		Status:   JobStatusCancelled,
		PageSize: 10,
	}, func(j *JobV2) error {
		if j.JobID == job.JobID {
			found = true
			return io.EOF
		}
		return nil
	}); err != nil && err != io.EOF {
		t.Fatal(err)
	}
	if !found {
		t.Errorf("QueryJobsV2 hasn't found the job")
	}
}

func newClient(t *testing.T) *Client {
	t.Helper()
	cs := os.Getenv("TEST_IOTHUB_SERVICE_CONNECTION_STRING")
	if cs == "" {
		t.Fatal("$TEST_IOTHUB_SERVICE_CONNECTION_STRING is empty")
	}
	c, err := NewFromConnectionString(cs)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func newDevice(t *testing.T, c *Client) *Device {
	t.Helper()
	device := &Device{
		DeviceID: "golang-iothub-test",
	}
	_ = c.DeleteDevice(context.Background(), device)
	device, err := c.CreateDevice(context.Background(), device)
	if err != nil {
		t.Fatal(err)
	}
	return device
}

func newDeviceAndModule(t *testing.T, c *Client) (*Device, *Module) {
	t.Helper()
	device := newDevice(t, c)
	module := &Module{
		DeviceID:  device.DeviceID,
		ModuleID:  "golang-iothub-test",
		ManagedBy: "admin",
	}
	_ = c.DeleteModule(context.Background(), module)
	module, err := c.CreateModule(context.Background(), module)
	if err != nil {
		t.Fatal(err)
	}
	return device, module
}

func newConfiguration(t *testing.T, c *Client) *Configuration {
	t.Helper()
	config := &Configuration{
		ID:              "golang-iothub-test",
		Priority:        10,
		SchemaVersion:   "1.0",
		TargetCondition: "deviceId='golang-iothub-test'",
		Labels: map[string]string{
			"test": "test",
		},
		Content: &ConfigurationContent{
			DeviceContent: map[string]interface{}{
				"properties.desired.testconf": 1.12,
			},
		},
		Metrics: &ConfigurationMetrics{
			Queries: map[string]string{
				"Total": "select deviceId from devices",
			},
		},
	}
	_ = c.DeleteConfiguration(context.Background(), config)
	config, err := c.CreateConfiguration(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	return config
}
