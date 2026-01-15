package internal

import (
        "fmt"
        "net"
        "strconv"
        "strings"
)

func ResolveLocalTCPAddress(addr string) (*net.TCPAddr, error) {
        // Resolve the local address into IP address and the given port if any
        host, port, err := net.SplitHostPort(addr)
        if err != nil {
                if !strings.Contains(err.Error(), "missing port") {
                        return nil, fmt.Errorf("invalid local address: %w", err)
                }
                host = addr
        }
        local, err := net.ResolveIPAddr("ip", host)
        if err != nil {
                return nil, fmt.Errorf("cannot resolve local address: %w", err)
        }

        var portNo int
        if port != "" {
                p, err := strconv.ParseUint(port, 10, 16)
                if err != nil {
                        return nil, fmt.Errorf("invalid port: %w", err)
                }
                portNo = int(p)
        }

        return &net.TCPAddr{IP: local.IP, Port: portNo, Zone: local.Zone}, nil
}
