package dockerhub

import (
	"fmt"
	"math/rand"
	"time"
)

// See https://docs.docker.com/docker-hub/webhooks/

const dockerid = "somerandomuser"
const hexBytes = "0123456789abcdef"
const imagename = "testimage"
const registry = "https://registry.hub.docker.com"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = hexBytes[rand.Intn(len(hexBytes))]
	}
	return string(b)
}

func NewEventJSONEncoded() string {
	return fmt.Sprintf(`{
"callback_url": %s,
"push_data": {
    "images": [
        %s,
        %s,
    ],
    "pushed_at" %v,
    "pusher": %s
},
"repository": {
    "comment_count": "0",
    "date_created": %v,
    "description: "",
    "dockerfile": "",
    "is_official": false,
    "is_private": true,
    "is_trusted": true,
    "name": "testhook",
    "namespace": "dazwilkin",
    "owner": %s,
    "repo_name": "dazwilkin/testwebhook",
    "repo_url": %s,
    "star_count": 0,
    "status": "Active" 
}
}`,
		registry,
		RandStringBytes(64),
		RandStringBytes(64),
		time.Now().Unix(),
		dockerid,
		time.Now().Unix(),
		dockerid,
		fmt.Sprintf("%s/%s", dockerid, imagename),
		fmt.Sprintf("%s/u/%s/%s", registry, dockerid, imagename))
}
