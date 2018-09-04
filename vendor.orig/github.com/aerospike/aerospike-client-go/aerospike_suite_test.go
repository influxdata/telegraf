package aerospike_test

import (
	"flag"
	"log"
	"math/rand"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	as "github.com/aerospike/aerospike-client-go"
)

var host = flag.String("h", "127.0.0.1", "Aerospike server seed hostnames or IP addresses")
var port = flag.Int("p", 3000, "Aerospike server seed hostname or IP address port number.")
var user = flag.String("U", "", "Username.")
var password = flag.String("P", "", "Password.")
var clientPolicy *as.ClientPolicy
var client *as.Client
var useReplicas = flag.Bool("use-replicas", false, "Aerospike will use replicas as well as master partitions.")

func initTestVars() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()

	clientPolicy = as.NewClientPolicy()
	if *user != "" {
		clientPolicy.User = *user
		clientPolicy.Password = *password
	}

	clientPolicy.RequestProleReplicas = *useReplicas

	if client == nil || !client.IsConnected() {
		client, err = as.NewClientWithPolicy(clientPolicy, *host, *port)
		if err != nil {
			log.Fatal(err.Error())
		}

		// set default policies
		if *useReplicas {
			client.DefaultPolicy.ReplicaPolicy = as.MASTER_PROLES
		}
	}
}

func TestAerospike(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Aerospike Client Library Suite")
}

func featureEnabled(feature string) bool {
	client, err := as.NewClientWithPolicy(clientPolicy, *host, *port)
	if err != nil {
		log.Fatal("Failed to connect to aerospike: err:", err)
	}

	node := client.GetNodes()[0]
	infoMap, err := node.RequestInfo("features")
	if err != nil {
		log.Fatal("Failed to connect to aerospike: err:", err)
	}

	return strings.Contains(infoMap["features"], feature)
}

func nsInfo(ns string, feature string) string {
	client, err := as.NewClientWithPolicy(clientPolicy, *host, *port)
	if err != nil {
		log.Fatal("Failed to connect to aerospike: err:", err)
	}

	node := client.GetNodes()[0]
	infoMap, err := node.RequestInfo("namespace/" + ns)
	if err != nil {
		log.Fatal("Failed to connect to aerospike: err:", err)
	}

	infoStr := infoMap["namespace/"+ns]
	infoPairs := strings.Split(infoStr, ";")
	for _, pairs := range infoPairs {
		pair := strings.Split(pairs, "=")
		if pair[0] == feature {
			return pair[1]
		}
	}

	return ""
}
