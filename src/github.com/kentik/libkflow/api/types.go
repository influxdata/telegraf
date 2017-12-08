package api

import (
	"encoding/json"
	"fmt"
	"net"
)

type DeviceResponse struct {
	Device *Device `json:"device"`
}

type Device struct {
	ID          int      `json:"id,string"`
	Name        string   `json:"device_name"`
	IP          net.IP   `json:"ip"`
	MaxFlowRate int      `json:"max_flow_rate"`
	CompanyID   int      `json:"company_id,string"`
	Customs     []Column `json:"custom_column_data,omitempty"`
}

type Column struct {
	ID   uint64 `json:"field_id,string"`
	Name string `json:"col_name"`
	Type string `json:"col_type"`
}

type Interface struct {
	ID      uint64 `json:"id,string"`
	Index   uint64 `json:"snmp_id,string"`
	Alias   string `json:"snmp_alias"`
	Desc    string `json:"interface_description"`
	Address string `json:"interface_ip"`
	Netmask string `json:"interface_ip_netmask"`
	Addrs   []Addr `json:"secondary_ips"`
}

type InterfaceUpdate struct {
	Index   uint64 `json:"index,string"`
	Alias   string `json:"alias"`
	Desc    string `json:"desc"`
	Speed   uint64 `json:"speed"`
	Type    uint64 `json:"type"`
	Address string `json:"address"`
	Netmask string `json:"netmask"`
	Addrs   []Addr `json:"alias_address"`
}

type Addr struct {
	Address string `json:"address"`
	Netmask string `json:"netmask"`
}

func (d *Device) ClientID() string {
	return fmt.Sprintf("%d:%s:%d", d.CompanyID, d.Name, d.ID)
}

func (c *Column) UnmarshalFlag(value string) error {
	return json.Unmarshal([]byte(value), c)
}

func (c Column) MarshalFlag() (string, error) {
	b, err := json.Marshal(c)
	return string(b), err
}
