package webhooks

import (
	"reflect"
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/webhooks/artifactory"
	"github.com/influxdata/telegraf/plugins/inputs/webhooks/filestack"
	"github.com/influxdata/telegraf/plugins/inputs/webhooks/github"
	"github.com/influxdata/telegraf/plugins/inputs/webhooks/mandrill"
	"github.com/influxdata/telegraf/plugins/inputs/webhooks/papertrail"
	"github.com/influxdata/telegraf/plugins/inputs/webhooks/particle"
	"github.com/influxdata/telegraf/plugins/inputs/webhooks/rollbar"
)

func TestAvailableWebhooks(t *testing.T) {
	wb := newWebhooks()
	expected := make([]Webhook, 0)
	if !reflect.DeepEqual(wb.availableWebhooks(), expected) {
		t.Errorf("expected to %v.\nGot %v", expected, wb.availableWebhooks())
	}

	wb.Artifactory = &artifactory.Webhook{Path: "/artifactory"}
	expected = append(expected, wb.Artifactory)
	if !reflect.DeepEqual(wb.availableWebhooks(), expected) {
		t.Errorf("expected to be %v.\nGot %v", expected, wb.availableWebhooks())
	}

	wb.Filestack = &filestack.Webhook{Path: "/filestack"}
	expected = append(expected, wb.Filestack)
	if !reflect.DeepEqual(wb.availableWebhooks(), expected) {
		t.Errorf("expected to be %v.\nGot %v", expected, wb.availableWebhooks())
	}

	wb.Github = &github.Webhook{Path: "/github"}
	expected = append(expected, wb.Github)
	if !reflect.DeepEqual(wb.availableWebhooks(), expected) {
		t.Errorf("expected to be %v.\nGot %v", expected, wb.availableWebhooks())
	}

	wb.Mandrill = &mandrill.Webhook{Path: "/mandrill"}
	expected = append(expected, wb.Mandrill)
	if !reflect.DeepEqual(wb.availableWebhooks(), expected) {
		t.Errorf("expected to be %v.\nGot %v", expected, wb.availableWebhooks())
	}

	wb.Papertrail = &papertrail.Webhook{Path: "/papertrail"}
	expected = append(expected, wb.Papertrail)
	if !reflect.DeepEqual(wb.availableWebhooks(), expected) {
		t.Errorf("expected to be %v.\nGot %v", expected, wb.availableWebhooks())
	}

	wb.Particle = &particle.Webhook{Path: "/particle"}
	expected = append(expected, wb.Particle)
	if !reflect.DeepEqual(wb.availableWebhooks(), expected) {
		t.Errorf("expected to be %v.\nGot %v", expected, wb.availableWebhooks())
	}

	wb.Rollbar = &rollbar.Webhook{Path: "/rollbar"}
	expected = append(expected, wb.Rollbar)
	if !reflect.DeepEqual(wb.availableWebhooks(), expected) {
		t.Errorf("expected to be %v.\nGot %v", expected, wb.availableWebhooks())
	}
}
