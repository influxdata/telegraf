package main

import (
	"log"
	"net"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/kentik/libkflow/api"
	"github.com/kentik/libkflow/api/test"
)

type Args struct {
	Port       int          `short:"p"          description:"listen on port "`
	Host       string       `long:"host"        description:"listen on host "`
	TLS        bool         `long:"tls"         description:"require TLS    "`
	Quiet      bool         `long:"quiet"       description:"minimize output"`
	Email      string       `long:"email"       description:"API auth email "`
	Token      string       `long:"token"       description:"API auth token "`
	CompanyID  int          `long:"company-id"  description:"company ID     "`
	DeviceID   int          `long:"device-id"   description:"device ID      "`
	DeviceName string       `long:"device-name" description:"device name    "`
	DeviceIP   string       `long:"device-ip"   description:"device IP addr "`
	MaxFPS     int          `long:"max-fps"     description:"max flows/sec  "`
	Customs    []api.Column `long:"custom"      description:"custom fields  "`
}

func main() {
	args := Args{
		Host:       "127.0.0.1",
		Port:       8999,
		TLS:        false,
		Quiet:      false,
		Email:      "test@example.com",
		Token:      "token",
		CompanyID:  1,
		DeviceID:   1,
		DeviceName: api.NormalizeName("dev1"),
		DeviceIP:   "127.0.0.1",
		MaxFPS:     4000,
		Customs: []api.Column{
			{ID: 1, Type: "uint32", Name: "RETRANSMITTED_IN_PKTS"},
			{ID: 2, Type: "uint32", Name: "RETRANSMITTED_OUT_PKTS"},
			{ID: 3, Type: "uint32", Name: "FRAGMENTS"},
			{ID: 4, Type: "uint32", Name: "CLIENT_NW_LATENCY_MS"},
			{ID: 5, Type: "uint32", Name: "SERVER_NW_LATENCY_MS"},
			{ID: 6, Type: "uint32", Name: "APPL_LATENCY_MS"},
			{ID: 7, Type: "uint32", Name: "OOORDER_IN_PKTS"},
			{ID: 8, Type: "uint32", Name: "OOORDER_OUT_PKTS"},
			{ID: 9, Type: "string", Name: "KFLOW_HTTP_URL"},
			{ID: 10, Type: "uint32", Name: "KFLOW_HTTP_STATUS"},
			{ID: 11, Type: "string", Name: "KFLOW_HTTP_UA"},
			{ID: 12, Type: "string", Name: "KFLOW_HTTP_REFERER"},
			{ID: 13, Type: "string", Name: "KFLOW_HTTP_HOST"},
			{ID: 14, Type: "string", Name: "KFLOW_DNS_QUERY"},
			{ID: 15, Type: "uint32", Name: "KFLOW_DNS_QUERY_TYPE"},
			{ID: 16, Type: "uint32", Name: "KFLOW_DNS_RET_CODE"},
			{ID: 17, Type: "string", Name: "KFLOW_DNS_RESPONSE"},
			{ID: 18, Type: "uint32", Name: "REPEATED_RETRANSMITS"},
			{ID: 19, Type: "uint32", Name: "RECEIVE_WINDOW"},
			{ID: 20, Type: "uint32", Name: "ZERO_WINDOWS"},
			{ID: 21, Type: "uint32", Name: "CONNECTION_ID"},
			{ID: 22, Type: "uint32", Name: "OOORDER_PKTS"},
			{ID: 23, Type: "uint32", Name: "RETRANSMITTED_PKTS"},
			{ID: 24, Type: "uint32", Name: "APP_PROTOCOL"},
			{ID: 25, Type: "uint32", Name: "INT00"},
			{ID: 26, Type: "uint32", Name: "INT01"},
			{ID: 27, Type: "uint32", Name: "INT02"},
			{ID: 28, Type: "uint32", Name: "INT03"},
			{ID: 29, Type: "uint32", Name: "INT04"},
			{ID: 30, Type: "uint32", Name: "INT05"},
			{ID: 31, Type: "string", Name: "STR00"},
			{ID: 32, Type: "string", Name: "STR01"},
			{ID: 33, Type: "string", Name: "STR02"},
			{ID: 34, Type: "string", Name: "STR03"},
			{ID: 35, Type: "string", Name: "STR04"},
			{ID: 36, Type: "string", Name: "STR05"},
		},
	}

	parser := flags.NewParser(&args, flags.PassDoubleDash|flags.HelpFlag)
	if _, err := parser.Parse(); err != nil {
		switch err.(*flags.Error).Type {
		case flags.ErrHelp:
			parser.WriteHelp(os.Stderr)
			os.Exit(1)
		default:
			log.Fatal(err)
		}
	}

	s, err := test.NewServer(args.Host, args.Port, args.TLS, args.Quiet)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("listening on %s:%d", s.Host, s.Port)

	err = s.Serve(args.Email, args.Token, &api.Device{
		ID:          args.DeviceID,
		Name:        args.DeviceName,
		IP:          net.ParseIP(args.DeviceIP),
		MaxFlowRate: args.MaxFPS,
		CompanyID:   args.CompanyID,
		Customs:     args.Customs,
	})

	if err != nil {
		log.Fatal(err)
	}
}
