// +build windows

package win_iisproc

import (
	"encoding/csv"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//IWinCommand provides interface to run windows commands
type IWinCommand interface {
	RunCmdCommand(args ...string) ([]byte, error)
	RunPowershellCommand(args ...string) ([]byte, error)
	FileExists(fileName string) bool
}

//WinCommand is IWinCommand implementation
type WinCommand struct {
}

//IISProc is an implementation if telegraf.Input interfac, providing info about iis app pool
type IISProc struct {
	commandMgr IWinCommand
}

//WP is child node in AppCmd xml output
type WP struct {
	XMLName xml.Name `xml:"WP"`
	WpID    string   `xml:"WP.NAME,attr"`
	WPName  string   `xml:"APPPOOL.NAME,attr"`
}

//AppCmd is root node in xml output
type AppCmd struct {
	XMLName xml.Name `xml:"appcmd"`
	WPs     []WP     `xml:"WP"`
}

type AppPool struct {
	pid     int
	mem     float64
	cpu     float64
	appPool string
}

func (c *WinCommand) FileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return true
}
func (c *WinCommand) RunCmdCommand(args ...string) ([]byte, error) {
	cmd := exec.Command("cmd", args...)
	return cmd.Output()
}
func (c *WinCommand) RunPowershellCommand(args ...string) ([]byte, error) {
	cmd := exec.Command("powershell.exe", args...)
	return cmd.CombinedOutput()
}
func (s *IISProc) ListWP() (AppCmd, error) {
	var appCommand AppCmd
	var exeCommand = os.Getenv("windir") + "\\system32\\inetsrv\\appcmd.exe"
	if !s.commandMgr.FileExists(exeCommand) {
		return appCommand, errors.New("appcmd.exe doest not exists.You need to turn on Web Management Tools")
	}
	var command = "/c " + exeCommand + " list wp  /xml"
	output, err := s.commandMgr.RunCmdCommand(command)

	if err != nil {
		return appCommand, err
	}
	err = xml.Unmarshal(output, &appCommand)
	if err != nil {
		return appCommand, err
	}
	return appCommand, nil
}
func (s *IISProc) FindAppPoolName(appCommand AppCmd, pid int) string {
	for i := 0; i < len(appCommand.WPs); i++ {
		var x, _ = strconv.Atoi(appCommand.WPs[i].WpID)
		if x == pid {
			return appCommand.WPs[i].WPName
		}
	}
	return ""
}
func (s *IISProc) ListIISProc() ([]AppPool, error) {
	wp, err := s.ListWP()
	if err != nil {
		return nil, err
	}

	output, err := s.commandMgr.RunPowershellCommand("Get-Process w3wp", "|", "select Id,CPU,PM", "|", "ConvertTo-CSV -NoTypeInformation")
	var result = string(output)
	if err != nil {
		if strings.Contains(result, "Cannot find a process with the name") {
			fmt.Println("w3wp not found")
			//return nill,fmt.Errorf("Couldn not find w3wp process")
		}
		return nil, err
	}
	var appPools []AppPool
	csvReader := csv.NewReader(strings.NewReader(result))
	csvReader.LazyQuotes = true
	for k := 0; ; k++ {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if k == 0 { //skip header
			continue
		}
		var pd, _ = strconv.Atoi(record[0])
		var cpu, _ = strconv.ParseFloat(record[1], 64)
		var mem, _ = strconv.ParseFloat(record[2], 64)
		var appName = s.FindAppPoolName(wp, pd)
		app := AppPool{
			pid:     pd,
			mem:     mem,
			cpu:     cpu,
			appPool: appName,
		}
		appPools = append(appPools, app)
	}
	return appPools, nil
}

func (s *IISProc) SampleConfig() string {
	return ""
}

var description = "Input plugin to report IIS worker processes per app pool."

func (s *IISProc) Description() string {
	return description
}

func (s *IISProc) Gather(acc telegraf.Accumulator) error {

	procList, err := s.ListIISProc()
	if err != nil {
		return err
	}
	for _, procName := range procList {
		tags := map[string]string{
			"appPool": procName.appPool,
		}
		fields := map[string]interface{}{
			"cpu": procName.cpu,
			"mem": procName.mem,
		}
		acc.AddFields("iis_proc", fields, tags)
	}
	return nil
}

func init() {
	inputs.Add("win_iisproc", func() telegraf.Input { return &IISProc{commandMgr: &WinCommand{}} })
}
