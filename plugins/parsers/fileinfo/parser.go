package fileinfo

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type FileInfo struct {
	Dir        string
	Base       string
	Time       time.Time
	Type       string
	Equipment  string
	Site       string
	Extension  string
	Outgoing   string
	Error      string
	Relative   string
	OsFileInfo os.FileInfo
}

type FileInfoParser struct {
	DefaultTags map[string]string
	IncomingDir string
}

func NewFileInfoParser() (*FileInfoParser, error) {
	return &FileInfoParser{}, nil
}

// Provided so that you can accurately calcuate the relative path against
// A specific source directory
func (p *FileInfoParser) SetIncomingDir(dir string) {
	p.IncomingDir = dir
}

func (p *FileInfoParser) GetFileInfo(fileName string) (*FileInfo, error) {
	var baseName = filepath.Base(fileName)
	var dirName = filepath.Dir(fileName)
	var splitName = strings.Split(baseName, "_")
	if len(splitName) < 5 {
		return nil, errors.New("Not a fileinfo parseable file")
	}
	var equipment = splitName[4]
	var site = equipment[0:3]
	var splitExt = strings.Split(splitName[5], ".")
	var relative = fileName
	if len(p.IncomingDir) > 0 {
		relative = strings.TrimPrefix(fileName, p.IncomingDir)
	}

	var fi FileInfo
	var err error
	fi.OsFileInfo, err = os.Stat(fileName)
	if err != nil {
		return nil, err
	}
	fi.Base = baseName
	fi.Dir = dirName
	fi.Equipment = equipment
	fi.Type = splitExt[0]
	fi.Extension = filepath.Ext(fileName)
	fi.Relative = relative
	fi.Site = site
	fi.Time, err = time.ParseInLocation("20060102T150405.000000", splitName[0]+"T"+splitName[1]+"."+splitName[2]+splitName[3], time.Local)
	if err != nil {
		fi.Time = time.Unix(0, 0)
		log.Println("ERROR [time]: ", err)
	}

	return &fi, nil
}

func (p *FileInfoParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	line := string(buf[:len(buf)])
	var metrics []telegraf.Metric
	metric, err := p.ParseLine(line)
	if metric == nil && err == nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	metrics = append(metrics, metric)
	return metrics, nil
}

func (p *FileInfoParser) ParseLine(line string) (telegraf.Metric, error) {
	fi, err := p.GetFileInfo(line)
	if err != nil && err.Error() == "Not a fileinfo parseable file" {
		return nil, nil
	}
	if err != nil && fi != nil {
		log.Println("[ERROR]: Could not get file info for line", line)
		return nil, err
	}
	fields := make(map[string]interface{})
	tags := make(map[string]string)
	fields["filename"] = fi.Base
	fields["relative"] = fi.Relative
	fields["filesize"] = fi.OsFileInfo.Size()
	fields["modtime"] = fi.OsFileInfo.ModTime().String()
	fields["parsetime"] = time.Now().String()
	fields["time"] = fi.Time.String()
	fields["extension"] = fi.Extension
	tags["equipment"] = fi.Equipment
	tags["site"] = fi.Site
	tags["data_format"] = "fileinfo"

	m, err := metric.New("fileinfo", tags, fields, time.Now())

	if err != nil {
		return nil, err
	}

	return m, nil
}

func (p *FileInfoParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}
