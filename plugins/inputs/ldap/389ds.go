package ldap

import (
	"strconv"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

// Empty mappings are identity mappings
var attrMap389ds = map[string]string{
	"addentryops":                    "add_operations",
	"anonymousbinds":                 "anonymous_binds",
	"bindsecurityerrors":             "bind_security_errors",
	"bytesrecv":                      "bytes_received",
	"bytessent":                      "bytes_sent",
	"cacheentries":                   "cache_entries",
	"cachehits":                      "cache_hits",
	"chainings":                      "",
	"compareops":                     "compare_operations",
	"connections":                    "",
	"connectionsinmaxthreads":        "connections_in_max_threads",
	"connectionsmaxthreadscount":     "connections_max_threads",
	"copyentries":                    "copy_entries",
	"currentconnections":             "current_connections",
	"currentconnectionsatmaxthreads": "current_connections_at_max_threads",
	"dtablesize":                     "",
	"entriesreturned":                "entries_returned",
	"entriessent":                    "entries_sent",
	"errors":                         "",
	"inops":                          "in_operations",
	"listops":                        "list_operations",
	"removeentryops":                 "delete_operations",
	"masterentries":                  "master_entries",
	"maxthreadsperconnhits":          "maxthreads_per_conn_hits",
	"modifyentryops":                 "modify_operations",
	"modifyrdnops":                   "modrdn_operations",
	"nbackends":                      "backends",
	"onelevelsearchops":              "onelevel_search_operations",
	"opscompleted":                   "operations_completed",
	"opsinitiated":                   "operations_initiated",
	"readops":                        "read_operations",
	"readwaiters":                    "read_waiters",
	"referrals":                      "referrals",
	"referralsreturned":              "referrals_returned",
	"searchops":                      "search_operations",
	"securityerrors":                 "security_errors",
	"simpleauthbinds":                "simpleauth_binds",
	"slavehits":                      "slave_hits",
	"strongauthbinds":                "strongauth_binds",
	"threads":                        "",
	"totalconnections":               "total_connections",
	"unauthbinds":                    "unauth_binds",
	"wholesubtreesearchops":          "wholesubtree_search_operations",
}

func (l *LDAP) new389dsConfig() []request {
	attributes := make([]string, 0, len(attrMap389ds))
	for k := range attrMap389ds {
		attributes = append(attributes, k)
	}

	req := ldap.NewSearchRequest(
		"cn=Monitor",
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(objectClass=*)",
		attributes,
		nil,
	)
	return []request{{req, l.convert389ds}}
}

func (l *LDAP) convert389ds(result *ldap.SearchResult, ts time.Time) []telegraf.Metric {
	tags := map[string]string{
		"server": l.host,
		"port":   l.port,
	}
	fields := make(map[string]interface{})
	for _, entry := range result.Entries {
		for _, attr := range entry.Attributes {
			if len(attr.Values[0]) == 0 {
				continue
			}
			// Map the attribute-name to the field-name
			name := attrMap389ds[attr.Name]
			if name == "" {
				name = attr.Name
			}
			// Reverse the name if requested
			if l.ReverseFieldNames {
				parts := strings.Split(name, "_")
				for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
					parts[i], parts[j] = parts[j], parts[i]
				}
				name = strings.Join(parts, "_")
			}

			// Convert the number
			if v, err := strconv.ParseInt(attr.Values[0], 10, 64); err == nil {
				fields[name] = v
			}
		}
	}

	m := metric.New("389ds", tags, fields, ts)
	return []telegraf.Metric{m}
}
