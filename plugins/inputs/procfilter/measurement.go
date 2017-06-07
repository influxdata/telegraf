package procfilter

import (
	"fmt"
	"strconv"

	//"github.com/shirou/gopsutil/process"
)

/* A measurement is used to define what tag/values of a set of (filterd) processes will be output.
A measurement contains one filter and thus implements the filter interface by proxy.
*/
type measurement struct {
	name   string
	tags   []string
	fields []string
	f      filter
}

func (m *measurement) Apply() error {
	return m.f.Apply()
}

func (m *measurement) parse(p *Parser) error {
	return p.syntaxError("measurement.Parse() shound not be called")
}

func (m *measurement) Stats() *stats {
	return m.f.Stats()
}

func (m *measurement) getTags(s stat, prefix string) (map[string]string, error) {
	tags := map[string]string{}
	var prefTag string
	for _, tag := range m.tags {
		if prefix == "" {
			prefTag = tag
		} else {
			prefTag = prefix + tag
		}
		switch tag {
		case "user":
			v, err := s.Users()
			if err != nil || v[0] == "" {
				//uids, _ := s.UIDs()
				uid := s.(*packStat).uid
				fmt.Printf("unknown UID %d %s\n", uid, err)
			}
			if err != nil {
				continue
			}
			tags[prefTag] = v[0]
		case "group":
			v, err := s.Groups()
			if err != nil {
				continue
			}
			tags[prefTag] = v[0]
		case "cmd":
			v, err := s.Cmd()
			if err != nil {
				continue
			}
			tags[prefTag] = v
		case "exe":
			v, err := s.Exe()
			if err != nil {
				continue
			}
			tags[prefTag] = v
		case "cmd_pid":
			v, err := s.Cmd()
			if err != nil {
				continue
			}
			pid := s.PID()
			if pid < 0 {
				continue
			}
			tags[prefTag] = fmt.Sprintf("%s-%d", v, pid)
		case "pid":
			pid := s.PID()
			if pid < 0 {
				continue
			}
			tags[prefTag] = strconv.Itoa(int(pid))
		case "uid":
			vs, err := s.UIDs()
			if err != nil {
				continue
			}
			tags[prefTag] = strconv.Itoa(int(vs[0]))
		case "gid":
			vs, err := s.GIDs()
			if err != nil {
				continue
			}
			tags[prefTag] = strconv.Itoa(int(vs[0]))
		default:
			return nil, NYIError(fmt.Sprintf("tag '%s'", tag))
		}
	}
	return tags, nil
}

func (m *measurement) getFields(s stat, prefix string) (map[string]interface{}, error) {
	fields := map[string]interface{}{}
	var prefField string
	for _, field := range m.fields {
		if prefix == "" {
			prefField = field
		} else {
			prefField = prefix + field
		}
		switch field {
		case "user":
			v, err := s.Users()
			if err != nil {
				continue
			}
			fields[prefField] = v[0]
		case "group":
			v, err := s.Groups()
			if err != nil {
				continue
			}
			fields[prefField] = v[0]
		case "cmd":
			v, err := s.Cmd()
			if err != nil || v == "" {
				continue
			}
			fields[prefField] = v
		case "exe":
			v, err := s.Exe()
			if err != nil {
				continue
			}
			fields[prefField] = v
		case "path":
			v, err := s.Path()
			if err != nil {
				continue
			}
			fields[prefField] = v
		case "cmd_line", "cmdline", "commandline":
			v, err := s.CmdLine()
			if err != nil {
				continue
			}
			fields[prefField] = v
		case "pid":
			fields[prefField] = s.PID()
		case "uid":
			vs, err := s.GIDs()
			if err != nil {
				continue
			}
			fields[prefField] = vs[0]
		case "gid":
			vs, err := s.GIDs()
			if err != nil {
				continue
			}
			fields[prefField] = vs[0]
		case "rss":
			v, err := s.RSS()
			if err != nil {
				continue
			}
			fields[prefField] = v
		case "vsz":
			v, err := s.VSZ()
			if err != nil {
				continue
			}
			fields[prefField] = v
		case "swap":
			v, err := s.Swap()
			if err != nil {
				continue
			}
			fields[prefField] = v
		case "cpu", "cpu_percent":
			v, err := s.CPU()
			if stamp == 1 {
				// CPU is not known until 2nd sample
				continue
			}
			if err != nil {
				continue
			}
			fields[prefField] = v
		case "thread_nb":
			v, err := s.ThreadNumber()
			if err != nil {
				continue
			}
			fields[prefField] = v
		case "fd_nb":
			v, err := s.FDNumber()
			if err != nil {
				continue
			}
			fields[prefField] = v
		case "process_nb":
			v := s.ProcessNumber()
			fields[prefField] = v
		default:
			return nil, NYIError(fmt.Sprintf("field '%s'", field))
		}
	}
	return fields, nil
}
