package spectrum_scale

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type spectrum_scale struct {
	Sensors        []string
	SocketLocation string

	allFields map[string]map[string]interface{}
}

var sampleConfig = `
  # An array of Spectrum Scale (GPFS) sensors
  #
  # These will be monitored through the mmpmon socket

  sensors = ["nsd_ds", "gfis", "fis"]
  socketLocation = "/var/mmfs/mmpmon/mmpmonSocket"

`

func (s *spectrum_scale) SampleConfig() string {
	return sampleConfig
}

func (s *spectrum_scale) Description() string {
	return "Read metrics of Spectrum Scale"
}

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (s *spectrum_scale) Gather(acc telegraf.Accumulator) error {
	var sensors []string
	var socketLocation string

	s.allFields = make(map[string]map[string]interface{})

	sensors = s.Sensors
	socketLocation = s.SocketLocation

	if len(sensors) == 0 {
		sensors = []string{"nsd_ds", "gfis", "fis"}
	}

	if len(socketLocation) == 0 {
		socketLocation = "/var/mmfs/mmpmon/mmpmonSocket"
	}

	// For every sensor specified, connect to the socket and parse the result
	for _, sensor := range sensors {
		err := s.getSensor(socketLocation, sensor, acc)
		if err != nil {
			return err
		}
	}
	return nil
}

func isField(t string) bool {
	if strings.HasPrefix(t, "_") && strings.HasSuffix(t, "_") {
		return true
	}
	return false
}

func (s *spectrum_scale) getSensor(socketPath string, sensor string, acc telegraf.Accumulator) error {
	conn, err := net.Dial("unix", socketPath)

	if err != nil {
		return fmt.Errorf("Could not connect to socket '%s': %s", socketPath, err)
	}

	defer conn.Close()

	_, errw := conn.Write([]byte("mmpmon " + sensor + "\n"))

	if errw != nil {
		return fmt.Errorf("Could not write to socket '%s': %s", socketPath, errw)
	}

	now := time.Now()

	scanner := bufio.NewScanner(bufio.NewReader(conn))

	scanner.Split(bufio.ScanWords)

	// Generic tags
	var cluster, filesystem, ipaddr string = "", "", ""
	// Specific tags
	var dnn, rg, da, vdisk, location, device, disk string = "", "", "", "", "", "", ""

	// Collected information
	var field, value, mode string = "", "", ""

	var twoFields bool = false
	var fields map[string]interface{}
	fields = make(map[string]interface{})
	var tags map[string]string
	tags = make(map[string]string)

	var sensorField string = fmt.Sprintf("_mmpmon::%s_", sensor)

	for scanner.Scan() {

		var token string = scanner.Text()
		field = ""
		value = ""

		if token == "" || (len(scanner.Bytes()) == 0) {
			// Exit on empty token
			break
		}

		// If the token is a field (in the shape of _x_), read in the next field as a value
		if isField(token) {
			field = token
			scanner.Scan()
			value = scanner.Text()

			if (field == "_event_") || (field == "_node_") {
				scanner.Scan()
				continue
			}

			if (field == "_response_") || (field == sensorField) {

				if len(fields) >= 3 {
					acc.AddFields("spectrum_scale", fields, tags, now)
					if (field == "_response_") && (value == "end") {
						// Exit on end of stream
						break
					}

					fields = make(map[string]interface{})
					tags = make(map[string]string)

					// Clear tags
					cluster, filesystem, ipaddr = "", "", ""
					dnn, rg, da, vdisk, location, device, disk = "", "", "", "", "", "", ""
					mode = ""
				}

			}

			if isField(value) {
				switch field {
				case "_r_":
					mode = "read"
				case "_w_":
					mode = "write"
				}
				field = value
				scanner.Scan()
				value = scanner.Text()
			}

			twoFields = false

			switch sensor {
			case "vfss":
				// Most of the 'vfss' sensor fields consists of two values, but only capture the integer
				switch field {
				case "_response_", "_event_", "mmpmon", "_mmpmon::vfss_", "_n_", "_node_", "_nn_", "_rc_", "_t_", "_tu_":
					break
				default:
					twoFields = true
				}

			}

			if twoFields {
				scanner.Scan()
			}

			fieldname := ""

			// Translate the field into human-readable and more descriptive, if possible
			switch field {
			case "_response_", "_event_":
				break
			case "_n_":
				ipaddr = value
			case "_nn_":
				dnn = value // daemon node name
			case "_cl_":
				cluster = value
			case "_fs_":
				filesystem = value
			case "_rg_":
				rg = value
			case "_da_":
				da = value
			case "_v_":
				vdisk = value
			case "_locn_":
				location = value
			case "_dev_":
				device = value
			case "_d_":
				disk = value
			case "_ops_":
				fieldname = fmt.Sprintf("%s_calls", mode)
			case "_b_":
				fieldname = fmt.Sprintf("%s_bytes", mode)
			case "_br_":
				fieldname = "read_bytes_disk"
			case "_bw_":
				fieldname = "write_bytes" //Total number of bytes written, to both disk and cache
			case "_c_":
				fieldname = "read_calls_cache" //read ops from cache
			case "_r_":
				switch sensor {
				case "vios":
					fieldname = "client_reads"
				default:
					fieldname = "read_calls_disk" //read ops from disk
				}
			case "_w_":
				fieldname = "write_calls" //write ops to both disk & cache
			case "_oc_":
				fieldname = "open" //open() calls
			case "_cc_":
				fieldname = "close" //close() calls
			case "_rdc_":
				fieldname = "read_req" //app read requests serviced by gpfs
			case "_wc_":
				fieldname = "write_req" //app write requests serviced by gpfs
			case "_dir_":
				fieldname = "readdir" //readdir()
			case "_iu_":
				fieldname = "inode_update" //inode updates to disk
			case "_irc_":
				fieldname = "inode_read" //inode reads
			case "_idc_":
				fieldname = "inode_del" //inode dels
			case "_icc_":
				fieldname = "inode_create" //inode creations
			case "_bc_":
				fieldname = "read_bytes_cache" //bytes read from cache
			case "_sch_":
				fieldname = "stat_cache_hit" //stat cache hits
			case "_scm_":
				fieldname = "stat_cache_miss" //stat cache misses
			case "_tw_":
				fieldname = "total_wait_time" // Total time waiting for disk operations, in seconds
			case "_qt_":
				fieldname = "total_queued_time" // Total time spent between being queued for a disk operation and the completion of that operation
			case "_stw_":
				fieldname = "shortest_wait_time" // Shortest time spent waiting for a disk operation
			case "_sqt_":
				fieldname = "shortest_queued_time" // Shortest time between being queued for a disk operation and the completion of that operation
			case "_ltw_":
				fieldname = "longest_wait_time" // Longest spent waiting for a disk oper
			case "_lqt_":
				fieldname = "longest_queued_time" // Longest time between being queued for a disk operation and the completion of that oper
			case "_t_":
				fieldname = "time"
			case "_tu_":
				fieldname = "time_microseconds" // Microseconds part of the sample time
			case "_seq_":
				fieldname = "sequence_num"
			case "_noncri_":
				fieldname = "threads_noncritical"
			case "_daestr_":
				fieldname = "threads_daemonstartup"
			case "_mbhan_":
				fieldname = "threads_mbhandler"
			case "_rcvwor_":
				fieldname = "threads_rcvworkers"
			case "_revwor_":
				fieldname = "threads_revokeworkers"
			case "_rngrvk_":
				fieldname = "threads_rangerevokeworkers"
			case "_recrvk_":
				fieldname = "threads_reclockrevokeworkers"
			case "_prefth_":
				fieldname = "threads_prefetchworkers"
			case "_sgexpn_":
				fieldname = "threads_sgexception"
			case "_recv_":
				fieldname = "threads_receivers"
			case "_pcache_":
				fieldname = "threads_pcache"
			case "_multh_":
				fieldname = "multithreadwork"
			case "_sw_":
				fieldname = "client_short_writes"
			case "_mw_":
				fieldname = "client_medium_writes"
			case "_pfw_":
				fieldname = "client_promoted_full_track_writes"
			case "_ftw_":
				fieldname = "client_full_track_writes"
			case "_fuw_":
				fieldname = "flushed_update_writes"
			case "_fpw_":
				fieldname = "flushed_promoted_full_track_writes"
			case "_m_":
				fieldname = "migrate_operations"
			case "_s_":
				fieldname = "scrub_operations"
			case "_l_":
				fieldname = "log_writes"
			case "_fc_":
				fieldname = "force_consistency_operations"
			case "_fix_":
				fieldname = "fixit_operations"
			case "_ltr_":
				fieldname = "log_tip_read_operations"
			case "_lhr_":
				fieldname = "log_home_read_operations"
			case "_rgd_":
				fieldname = "rgdesc_writes"
			case "_meta_":
				fieldname = "metadata_writes"
			default:
				// just remove leading and trailing underscores
				fieldname = strings.Trim(field, "_")
			}

			if fieldname != "" {

				tags["sensor"] = sensor
				if cluster != "" {
					tags["cluster"] = cluster
				}
				if filesystem != "" {
					tags["filesystem"] = filesystem
				}
				if ipaddr != "" {
					tags["ipaddr"] = ipaddr
				}
				if dnn != "" {
					tags["daemon_node_name"] = dnn
				}
				if rg != "" {
					tags["recovery_group"] = rg
				}
				if da != "" {
					tags["declustered_array"] = da
				}
				if vdisk != "" {
					tags["vdisk"] = vdisk
				}
				if location != "" {
					tags["location"] = location
				}
				if device != "" {
					tags["device"] = device
				}
				if disk != "" {
					tags["disk"] = disk
				}

				switch field {
				case "_noncri_", "_daestr_", "_mbhan_", "_rcvwor_", "_revwor_", "_rngrvk_", "_recrvk_", "_prefth_", "_sgexpn_", "_recv_", "_pcache_", "_multh_":

					threadinfo := strings.Split(value, "/")
					threadcounts := []int{}

					for _, tinfo := range threadinfo {
						tcount, err := strconv.Atoi(tinfo)
						if err != nil {
							fmt.Println(err)
							os.Exit(2)
						}
						threadcounts = append(threadcounts, tcount)
					}

					fields[fmt.Sprintf("%s_current", fieldname)] = threadcounts[0]
					fields[fmt.Sprintf("%s_highest", fieldname)] = threadcounts[1]
					fields[fmt.Sprintf("%s_maximum", fieldname)] = threadcounts[2]
				default:
					if strings.Contains(value, ".") {
						// These values should be floats, while others are integers
						parsedvalue, err := strconv.ParseFloat(value, 64)
						if err != nil {
							fmt.Println(err)
							os.Exit(2)
						}
						fields[fieldname] = parsedvalue
					} else {
						parsedvalue, err := strconv.Atoi(value)
						if err != nil {
							// handle error
							fmt.Println(err)
							os.Exit(2)
						}
						fields[fieldname] = parsedvalue
					}
				}
			}

		}

	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading input:", err)
	}
	return nil

}

func init() {
	inputs.Add("spectrum_scale", func() telegraf.Input {
		return &spectrum_scale{}
	})
}
