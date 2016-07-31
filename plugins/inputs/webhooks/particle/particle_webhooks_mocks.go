package particle

import (
    "fmt"
    "math/rand"
    "net/url"
    "time"
)

const hexBytes = "0123456789abcdef"

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

func NewEventURLEncoded() string {
    rand.Seed(time.Now().UnixNano())
    return fmt.Sprintf("event=%v&data=%v&published_at=%v&coreid=%v",
        "event",
        rand.Intn(1000),
        url.QueryEscape(time.Now().Format(time.RFC3339)),
        RandStringBytes(24))
}