package docker_log

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"io"
	"log"
)

//Docker client wrapper
type dClient interface {
	ContainerInspect(ctx context.Context, contID string) (types.ContainerJSON, error)
	ContainerLogs(ctx context.Context, contID string, options types.ContainerLogsOptions) (io.ReadCloser, error)
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	Close() error
}

type offsetData struct {
	contID string
	offset int64
}

type logger struct {
	errHeader    string
	warnHeader   string
	infoHeader   string
	debugHeader  string
	headerFormat string
}

var lg = newLogger(inputTitle)

func newLogger(header string) *logger {
	return &logger{
		headerFormat: "%s %s",
		errHeader:    fmt.Sprintf("E! [%s]", header),
		warnHeader:   fmt.Sprintf("W! [%s]", header),
		infoHeader:   fmt.Sprintf("I! [%s]", header),
		debugHeader:  fmt.Sprintf("D! [%s]", header)}
}
func (l logger) logE(format string, v ...interface{}) {
	log.Printf(l.headerFormat, l.errHeader, fmt.Sprintf(format, v...))
}
func (l logger) logW(format string, v ...interface{}) {
	log.Printf(l.headerFormat, l.warnHeader, fmt.Sprintf(format, v...))
}
func (l logger) logI(format string, v ...interface{}) {

	log.Printf(l.headerFormat, l.infoHeader, fmt.Sprintf(format, v...))
}
func (l logger) logD(format string, v ...interface{}) {
	log.Printf(l.headerFormat, l.debugHeader, fmt.Sprintf(format, v...))
}

func trimId(id string) string {
	trimmedId := id
	if len(trimmedId) > 12 {
		trimmedId = trimmedId[0:12]
	}
	return trimmedId
}
