package twitter

import (
	"reflect"
	"testing"

	"github.com/ChimeraCoder/anaconda"
	"github.com/stretchr/testify/require"
)

func TestGetTags(t *testing.T) {
	id := "1967601206"
	screenName := "InfluxDB"

	user := anaconda.User{
		IdStr:      id,
		ScreenName: screenName,
	}

	getTagsReturn := getTags(user)

	correctTagsReturn := map[string]string{
		"id":          id,
		"screen_name": screenName,
	}

	require.Equal(t, true, reflect.DeepEqual(getTagsReturn, correctTagsReturn))
}

func TestGetFields(t *testing.T) {
	favourites := 1
	followers := 2
	friends := 3
	statuses := int64(4)

	user := anaconda.User{
		FavouritesCount: favourites,
		FollowersCount:  followers,
		FriendsCount:    friends,
		StatusesCount:   statuses,
	}

	getFieldsReturn := getFields(user)

	correctFieldReturn := make(map[string]interface{})

	correctFieldReturn["favourites"] = 1
	correctFieldReturn["followers"] = 2
	correctFieldReturn["friends"] = 3
	correctFieldReturn["statuses"] = int64(4)

	require.Equal(t, true, reflect.DeepEqual(getFieldsReturn, correctFieldReturn))
}
