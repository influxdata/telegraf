package ldap

import (
	"strconv"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

var attrMapOpenLDAP = map[string]string{
	"monitorCounter":     "",
	"monitoredInfo":      "",
	"monitorOpInitiated": "_initiated",
	"monitorOpCompleted": "_completed",
	"olmMDBPagesMax":     "_mdb_pages_max",
	"olmMDBPagesUsed":    "_mdb_pages_used",
	"olmMDBPagesFree":    "_mdb_pages_free",
	"olmMDBReadersMax":   "_mdb_readers_max",
	"olmMDBReadersUsed":  "_mdb_readers_used",
	"olmMDBEntries":      "_mdb_entries",
}

func (l *LDAP) newOpenLDAPConfig() []request {
	req := ldap.NewSearchRequest(
		"cn=Monitor",
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(|(objectClass=monitorCounterObject)(objectClass=monitorOperation)(objectClass=monitoredObject)(objectClass=monitorContainer))",
		[]string{"monitorCounter", "monitorOpInitiated", "monitorOpCompleted", "monitoredInfo"},
		nil,
	)
	return []request{{req, l.convertOpenLDAP}}
}

func (l *LDAP) convertOpenLDAP(result *ldap.SearchResult, ts time.Time) []telegraf.Metric {
	tags := map[string]string{
		"server": l.host,
		"port":   l.port,
	}

	fields := make(map[string]interface{})
	for _, entry := range result.Entries {
		prefix := openLDAPAttrConvertDN(entry.DN, l.ReverseFieldNames)
		for _, attr := range entry.Attributes {
			if len(attr.Values[0]) == 0 {
				continue
			}
			if v, err := strconv.ParseInt(attr.Values[0], 10, 64); err == nil {
				fields[prefix+attrMapOpenLDAP[attr.Name]] = v
			}
		}
	}

	m := metric.New("openldap", tags, fields, ts)
	return []telegraf.Metric{m}
}

// Convert a DN to a field prefix, eg cn=Read,cn=Waiters,cn=Monitor becomes waiters_read
// Assumes the last part of the DN is cn=Monitor and we want to drop it
func openLDAPAttrConvertDN(dn string, reverse bool) string {
	// Normalize DN
	prefix := strings.TrimSpace(dn)
	prefix = strings.ToLower(prefix)
	prefix = strings.ReplaceAll(prefix, " ", "_")
	prefix = strings.ReplaceAll(prefix, "cn=", "")

	// Filter the base
	parts := strings.Split(prefix, ",")
	for i, p := range parts {
		if p == "monitor" {
			parts = append(parts[:i], parts[i+1:]...)
			break
		}
	}

	if reverse {
		for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
			parts[i], parts[j] = parts[j], parts[i]
		}
	}
	return strings.Join(parts, "_")
}
