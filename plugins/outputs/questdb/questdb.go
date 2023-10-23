//go:generate ../../../tools/readme_config_includer/generator
package questdb

import (
	"bufio"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"math/big"
	"net"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

type QuestDB struct {
	User            string        `toml:"user"`
	Token           config.Secret `toml:"token"`
	Address         string
	KeepAlivePeriod *config.Duration
	tlsint.ClientConfig
	Log  telegraf.Logger `toml:"-"`
	Conn net.Conn

	serializer influx.Serializer
	encoder    internal.ContentEncoder
}

func (*QuestDB) SampleConfig() string {
	return sampleConfig
}

func (questdb *QuestDB) Connect() error {
	spl := strings.SplitN(questdb.Address, "://", 2)
	if len(spl) != 2 {
		return fmt.Errorf("invalid address: %s", questdb.Address)
	}
	if spl[0] != "tcp" && spl[0] != "tcp4" {
		return fmt.Errorf("unsupported protocol: %s, only tcp or tcp4 are supported", questdb.Address)
	}

	tlsCfg, err := questdb.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	// Decode the key if provided
	var key *ecdsa.PrivateKey
	token, err := questdb.Token.Get()
	if err != nil {
		return err
	}
	if questdb.User != "" && token.TemporaryString() != "" {
		keyRaw, err := base64.RawURLEncoding.DecodeString(token.TemporaryString())
		if err != nil {
			return err
		}
		key = new(ecdsa.PrivateKey)
		key.PublicKey.Curve = elliptic.P256()
		key.PublicKey.X, key.PublicKey.Y = key.PublicKey.Curve.ScalarBaseMult(keyRaw)
		key.D = new(big.Int).SetBytes(keyRaw)
		token.Destroy()
	}

	var c net.Conn
	if tlsCfg == nil {
		c, err = net.Dial(spl[0], spl[1])
	} else {
		c, err = tls.Dial(spl[0], spl[1], tlsCfg)
	}
	if err != nil {
		return err
	}

	if err := questdb.setKeepAlive(c); err != nil {
		questdb.Log.Debugf("Unable to configure keep alive (%s): %s", questdb.Address, err)
	}

	if key != nil {
		_, err = c.Write([]byte(questdb.User + "\n"))
		if err != nil {
			c.Close()
			return err
		}

		reader := bufio.NewReader(c)
		raw, err := reader.ReadBytes('\n')
		if len(raw) < 2 {
			c.Close()
			return err
		}
		// Remove the `\n` in the last position.
		raw = raw[:len(raw)-1]
		if err != nil {
			c.Close()
			return err
		}

		// Hash the challenge with sha256.
		hash := crypto.SHA256.New()
		hash.Write(raw)
		hashed := hash.Sum(nil)

		stdSig, err := ecdsa.SignASN1(rand.Reader, key, hashed)
		if err != nil {
			c.Close()
			return err
		}
		_, err = c.Write([]byte(base64.StdEncoding.EncodeToString(stdSig) + "\n"))
		if err != nil {
			c.Close()
			return err
		}
	}

	questdb.encoder, err = internal.NewIdentityEncoder()
	questdb.serializer = influx.Serializer{UintSupport: false}
	if err := questdb.serializer.Init(); err != nil {
		c.Close()
		return err
	}

	if err != nil {
		return err
	}

	questdb.Conn = c
	return nil
}

func (questdb *QuestDB) setKeepAlive(c net.Conn) error {
	if questdb.KeepAlivePeriod == nil {
		return nil
	}
	tcpc, ok := c.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("cannot set keep alive on a %s socket", strings.SplitN(questdb.Address, "://", 2)[0])
	}
	if *questdb.KeepAlivePeriod == 0 {
		return tcpc.SetKeepAlive(false)
	}
	if err := tcpc.SetKeepAlive(true); err != nil {
		return err
	}
	return tcpc.SetKeepAlivePeriod(time.Duration(*questdb.KeepAlivePeriod))
}

// Write writes the given metrics to the destination.
// If an error is encountered, it is up to the caller to retry the same write again later.
// Not parallel safe.
func (questdb *QuestDB) Write(metrics []telegraf.Metric) error {
	if questdb.Conn == nil {
		// previous write failed with permanent error and socket was closed.
		if err := questdb.Connect(); err != nil {
			return err
		}
	}

	for _, m := range metrics {
		bs, err := questdb.serializer.Serialize(m)
		if err != nil {
			questdb.Log.Debugf("Could not serialize metric: %v", err)
			continue
		}

		bs, err = questdb.encoder.Encode(bs)
		if err != nil {
			questdb.Log.Debugf("Could not encode metric: %v", err)
			continue
		}

		if _, err := questdb.Conn.Write(bs); err != nil {
			//TODO log & keep going with remaining strings
			var netErr net.Error
			if errors.As(err, &netErr) {
				// permanent error. close the connection
				questdb.Close()
				questdb.Conn = nil
				return fmt.Errorf("closing connection: %w", netErr)
			}
			return err
		}
	}

	return nil
}

// Close closes the connection. Noop if already closed.
func (questdb *QuestDB) Close() error {
	if questdb.Conn == nil {
		return nil
	}
	err := questdb.Conn.Close()
	questdb.Conn = nil
	return err
}

func init() {
	outputs.Add("questdb", func() telegraf.Output {
		return &QuestDB{}
	})
}
