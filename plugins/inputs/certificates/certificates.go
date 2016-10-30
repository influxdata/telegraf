package certificates

import (
	"crypto/tls"
	"net/http"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	// "github.com/yieldbot/telegraf-plugins/telegraf"
	// "github.com/yieldbot/telegraf-plugins/telegraf/plugins/inputs"

	"fmt"
	// "github.com/yieldbot/golang-jenkins"
)

// this is the cert data structure that will hold the values that I need to deal with
type Certificates struct {
	SHA1                string
	SubjectKeyId        string
	Version             int
	SignatureAlgorithm  string
	PublicKeyAlgorithm  string
	Subject             string
	DNSNames            []string
	NotBefore, NotAfter string
	ExpiresIn           string
	Issuer              string
	AuthorityKeyId      string
}

// sample config for the user
var sampleConfig = `
	## specify host for use as an additional tag in influx
	host = jenkins1
	## specify url via a url matching:
	##  [protocol://]address[:port]
	##  e.g.
	##    http://jenkins.service.consul:8080/
	##    http://jenkins.foo.com/
	url = http://jenkins.service.consul:8080
	## specify username and password for logging in to jenkins
	## password may optionally be a jenkins generated API token
	username = admin
	password = password
	## Specify insecure to ignore SSL errors
	insecure = true
`

// return the sample config to te user
func (j *Certificates) SampleConfig() string {
	return sampleConfig
}

// return the description of the telegraf input
func (j *Certificates) Description() string {
	return "Reads metrics from an SSL Certificate"
}

// return the sha hash of the cert field
func SHA1Hash(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return fmt.Sprintf("%X", h.Sum(nil))
}


func checkHost(domainName string, skipVerify bool) ([]SSLCerts, error) {

	//Connect network
	ipConn, err := net.DialTimeout("tcp", domainName, 10000*time.Millisecond)
	if err != nil {
		return nil, err
	}
	defer ipConn.Close()

	// Configure tls to look at domainName
	config := tls.Config{ServerName: domainName,
		InsecureSkipVerify: skipVerify}

	// Connect to tls
	conn := tls.Client(ipConn, &config)
	defer conn.Close()

	// Handshake with TLS to get certs
	hsErr := conn.Handshake()
	if hsErr != nil {
		return nil, hsErr
	}

    // get the certs
	certs := conn.ConnectionState().PeerCertificates

    // check t0 make sure we have certs
	if certs == nil || len(certs) < 1 {
		return nil, errors.New("Could not get server's certificate from the TLS connection.")
	}

    // compile the list of certs
	sslcerts := make([]SSLCerts, len(certs))

    // this will go through each cert in the list and get the details and then compile them into
	// a data structure that can then be iterated over.
	// The qurestion is should I just dump each of these instead of building them or continue w/
	// the data structure and then iterate over that in the end like the commandline tool does.
	for i, cert := range certs {
		s := SSLCerts{SHA1: SHA1Hash(cert.Raw), SubjectKeyId: fmt.Sprintf("%X", cert.SubjectKeyId),
			Version: cert.Version, SignatureAlgorithm: signatureAlgorithm[cert.SignatureAlgorithm],
			PublicKeyAlgorithm: publicKeyAlgorithm[cert.PublicKeyAlgorithm],
			Subject:            cert.Subject.CommonName,
			DNSNames:           cert.DNSNames,
			NotBefore:          cert.NotBefore.Local().String(),
			NotAfter:           cert.NotAfter.Local().String(),
			ExpiresIn:          ExpiresIn(cert.NotAfter.Local()),
			Issuer:             cert.Issuer.CommonName,
			AuthorityKeyId:     fmt.Sprintf("%X", cert.AuthorityKeyId),
		}
		sslcerts[i] = s

	}

	return sslcerts, nil
}

func (j *Certificates) Gather(acc telegraf.Accumulator) error {
	// auth := &gojenkins.Auth{
	// 	Username: j.Username,
	// 	ApiToken: j.Password,
	// }

	// client := gojenkins.NewJenkins(auth, j.URL)

	// if j.Insecure {
	// 	c := newInsecureHTTP()
	// 	client.OverrideHTTPClient(c)
	// }

	// err := j.gatherQueue(acc, client)
	// if err != nil {
	// 	return err
	// }

	// err = j.gatherSlaves(acc, client)
	// if err != nil {
	// 	return err
	// }

	// err = j.gatherSlaveLabels(acc, client)
	// if err != nil {
	// 	return err
	// }

	// return nil
}

func init() {
	// inputs.Add("jenkins", func() telegraf.Input {
	// 	return &Jenkins{}
	// })
}

// func (j *Jenkins) gatherQueue(acc telegraf.Accumulator, client *gojenkins.Jenkins) error {

// 	fields := make(map[string]interface{})
// 	tags := make(map[string]string)

// 	qSize := 0
// 	qMap := make(map[string]int)

// 	queue, err := client.GetQueue()
// 	if err != nil {
// 		return err
// 	}

// 	for _, item := range queue.Items {
// 		if item.Buildable {
// 			qSize++
// 			j, err := client.GetJobProperties(item.Task.Name)
// 			if err == nil {
// 				//split label and filter out operators
// 				labels := strings.Split(j.AssignedNode, " ")
// 				for _, label := range labels {
// 					if label != "&&" || label != "||" || label != "->" || label != "<->" {
// 						if val, ok := qMap[j.AssignedNode]; ok {
// 							qMap[j.AssignedNode] = val + 1
// 						} else {
// 							qMap[j.AssignedNode] = 1
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}

// 	fields["queue_size"] = qSize
// 	if j.Host != "" {
// 		tags["host"] = j.Host
// 	}
// 	if len(qMap) > 0 {
// 		for key := range qMap {
// 			fields[fmt.Sprintf("label_%s", key)] = qMap[key]
// 		}
// 	}
// 	tags["url"] = j.URL

// 	acc.AddFields("jenkins_queue", fields, tags)

// 	return nil
// }

// func (j *Jenkins) gatherSlaves(acc telegraf.Accumulator, client *gojenkins.Jenkins) error {
// 	fields := make(map[string]interface{})
// 	tags := make(map[string]string)

// 	var slaveCount = 0
// 	var busyCount = 0

// 	slaves, err := client.GetComputers()
// 	if err != nil {
// 		return err
// 	}

// 	for _, slave := range slaves {
// 		if slave.JnlpAgent {
// 			slaveCount++
// 			if !slave.Idle {
// 				busyCount++
// 			}
// 		}
// 	}

// 	fields["slave_count"] = slaveCount
// 	fields["slaves_busy"] = busyCount
// 	if j.Host != "" {
// 		tags["host"] = j.Host
// 	}
// 	tags["url"] = j.URL

// 	acc.AddFields("jenkins_slaves", fields, tags)

// 	return nil
// }

// func (j *Jenkins) gatherSlaveLabels(acc telegraf.Accumulator, client *gojenkins.Jenkins) error {
// 	fields := make(map[string]interface{})
// 	tags := make(map[string]string)

// 	slaves, err := client.GetComputers()
// 	if err != nil {
// 		return err
// 	}

// 	for _, slave := range slaves {
// 		if slave.JnlpAgent {
// 			conf, err := client.GetComputerConfig(slave.DisplayName)
// 			if err != nil {
// 				return err
// 			}
// 			labels := strings.Split(conf.Label, " ")
// 			for _, label := range labels {
// 				if val, ok := fields[label]; ok {
// 					fields[label] = val.(int) + 1
// 					if !slave.Idle {
// 						v := fields[fmt.Sprintf("%s_busy", label)]
// 						if v != nil {
// 							fields[fmt.Sprintf("%s_busy", label)] = v.(int) + 1
// 						} else {
// 							fields[fmt.Sprintf("%s_busy", label)] = 1
// 						}
// 					}
// 				} else {
// 					fields[label] = 1
// 					if !slave.Idle {
// 						fields[fmt.Sprintf("%s_busy", label)] = 1
// 					}
// 				}
// 			}
// 		}
// 	}

// 	if j.Host != "" {
// 		tags["host"] = j.Host
// 	}
// 	tags["url"] = j.URL

// 	acc.AddFields("jenkins_labels", fields, tags)

// 	return nil
// }

// func newInsecureHTTP() *http.Client {
// 	ntls := &http.Transport{
// 		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
// 	}
// 	return &http.Client{Transport: ntls}
// }
