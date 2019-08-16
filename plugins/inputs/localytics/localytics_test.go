package localytics

import (
	"net/http"
	"reflect"
	"testing"

	loc "github.com/Onefootball/go-localytics"
	"github.com/stretchr/testify/require"
)

func TestNewLocalyticsClient(t *testing.T) {
	httpClient := &http.Client{}
	l := &Localytics{}

	client := l.newLocalyticsClient(httpClient)
	require.NotNil(t, client)
}

func TestGetFields(t *testing.T) {
	sessions := 1
	closes := 1
	users := 1
	events := 1

	app := &loc.App{
		Stats: loc.Stats{
			Sessions: sessions,
			Closes:   closes,
			Users:    users,
			Events:   events,
		},
	}

	getFieldsReturn := getFields(app)

	correctFieldReturn := make(map[string]interface{})

	correctFieldReturn["sessions"] = sessions
	correctFieldReturn["closes"] = closes
	correctFieldReturn["users"] = users
	correctFieldReturn["events"] = events

	require.Equal(t, true, reflect.DeepEqual(getFieldsReturn, correctFieldReturn))
}

func TestGetTags(t *testing.T) {
	name := "fooBar"
	appID := "1"

	app := &loc.App{
		Name:  name,
		AppID: appID,
	}

	getTagsReturn := getTags(app)

	correctTagsReturn := make(map[string]string)

	correctTagsReturn["id"] = appID
	correctTagsReturn["name"] = name

	require.Equal(t, true, reflect.DeepEqual(getTagsReturn, correctTagsReturn))
}
