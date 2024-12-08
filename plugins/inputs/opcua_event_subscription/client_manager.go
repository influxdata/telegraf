package opcua_event_subscription

import (
    "context"
    "crypto/rsa"
    "crypto/x509"
    "encoding/pem"
    "fmt"
    "io/ioutil"
    "github.com/gopcua/opcua"
    "github.com/gopcua/opcua/ua"
    "github.com/influxdata/telegraf"
)

type ClientManager struct {
    Endpoint       string
    SecurityMode   string
    SecurityPolicy string
    Certificate    string
    PrivateKey     string
    Client         *opcua.Client
    Log            telegraf.Logger
    cancel         context.CancelFunc
}

func (cm *ClientManager) InitClient() error {
    cm.Log.Info("Create Client")
    opts := []opcua.Option{
        opcua.SecurityMode(ua.MessageSecurityModeFromString(cm.SecurityMode)),
        opcua.SecurityPolicy(cm.SecurityPolicy),
    }

    if cm.Certificate != "" && cm.PrivateKey != "" {
        cert, err := loadCertificate(cm.Certificate)
        if err != nil {
            cm.Log.Errorf("failed to load certificate: %v", err)
            return fmt.Errorf("failed to load certificate: %v", err)
        }
        key, err := loadPrivateKey(cm.PrivateKey)
        if err != nil {
            cm.Log.Errorf("failed to load private key: %v", err)
            return fmt.Errorf("failed to load private key: %v", err)
        }
        opts = append(opts, opcua.Certificate(cert), opcua.PrivateKey(key))
    }

    client, err := opcua.NewClient(cm.Endpoint, opts...)
    if err != nil {
        return fmt.Errorf("failed to create OPC UA client: %v", err)
    }

    err = client.Connect(context.Background())
    if err != nil {
        return fmt.Errorf("failed to connect to OPC UA server: %v", err)
    }
    cm.Client = client
    cm.Log.Info("Client connected")
    return nil
}

func loadCertificate(certPath string) ([]byte, error) {
    data, err := ioutil.ReadFile(certPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read certificate file: %v", err)
    }
    block, _ := pem.Decode(data)
    if block == nil || block.Type != "CERTIFICATE" {
        return nil, fmt.Errorf("failed to decode certificate PEM block")
    }
    return block.Bytes, nil
}

func loadPrivateKey(keyPath string) (*rsa.PrivateKey, error) {
    data, err := ioutil.ReadFile(keyPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read private key file: %v", err)
    }
    block, _ := pem.Decode(data)
    if block == nil || block.Type != "PRIVATE KEY" {
        return nil, fmt.Errorf("failed to decode private key PEM block")
    }
    privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
    if err != nil {
        return nil, fmt.Errorf("failed to parse private key: %v", err)
    }
    switch key := privKey.(type) {
    case *rsa.PrivateKey:
        return key, nil
    default:
        return nil, fmt.Errorf("unsupported private key type: %T", key)
    }
}