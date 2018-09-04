// Copyright 2013-2017 Aerospike, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aerospike_test

import (
	"time"

	as "github.com/aerospike/aerospike-client-go"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// ALL tests are isolated by SetName and Key, which are 50 random characters
var _ = Describe("Aerospike Node Tests", func() {
	initTestVars()

	Describe("Node Connection Pool", func() {
		// connection data
		var err error
		var client *as.Client

		BeforeEach(func() {
			// use the same client for all
			client, err = as.NewClientWithPolicy(clientPolicy, *host, *port)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("When Authentication is Used", func() {

			if *user != "" {

				It("must return error if it fails to authenticate", func() {
					clientPolicy := as.NewClientPolicy()
					clientPolicy.User = "non_existent_user"
					clientPolicy.Password = "non_existent_user"

					client, err = as.NewClientWithPolicy(clientPolicy, *host, *port)
					Expect(err).To(HaveOccurred())
				})

			}

		})

		Context("When No Connection Count Limit Is Set", func() {

			It("must return a new connection on every poll", func() {
				clientPolicy := as.NewClientPolicy()
				clientPolicy.LimitConnectionsToQueueSize = false
				clientPolicy.ConnectionQueueSize = 4
				clientPolicy.User = *user
				clientPolicy.Password = *password

				client, err = as.NewClientWithPolicy(clientPolicy, *host, *port)
				Expect(err).ToNot(HaveOccurred())
				defer client.Close()

				node := client.GetNodes()[0]

				for i := 0; i < 20; i++ {
					c, err := node.GetConnection(0)
					Expect(err).NotTo(HaveOccurred())
					Expect(c).NotTo(BeNil())
					Expect(c.IsConnected()).To(BeTrue())

					node.InvalidateConnection(c)
				}

			})

		})

		Context("When A Connection Count Limit Is Set", func() {

			It("must return an error when maximum number of connections are polled", func() {
				clientPolicy := as.NewClientPolicy()
				clientPolicy.LimitConnectionsToQueueSize = true
				clientPolicy.ConnectionQueueSize = 4
				clientPolicy.User = *user
				clientPolicy.Password = *password

				client, err = as.NewClientWithPolicy(clientPolicy, *host, *port)
				Expect(err).ToNot(HaveOccurred())
				defer client.Close()

				node := client.GetNodes()[0]

				// 4-1 is because we reserve a connection for tend
				for i := 0; i < 4-1; i++ {
					c, err := node.GetConnection(0)
					Expect(err).NotTo(HaveOccurred())
					Expect(c).NotTo(BeNil())
					Expect(c.IsConnected()).To(BeTrue())

					// don't call invalidate here; we are testing node's connection queue behaviour
					// if there are connections which are not invalidated.
					// Don't call close as well, since it automatically reduces the total conn count.
					// c.Close()
				}

				// 4-1 is because we reserve a connection for tend
				for i := 0; i < 4-1; i++ {
					_, err := node.GetConnection(0)
					Expect(err).To(HaveOccurred())
				}

			})

		})

		Context("When Idle Timeout Is Used", func() {

			It("must reuse connections before they become idle", func() {
				clientPolicy := as.NewClientPolicy()
				clientPolicy.IdleTimeout = 1000 * time.Millisecond
				// clientPolicy.TendInterval = time.Hour
				clientPolicy.User = *user
				clientPolicy.Password = *password

				client, err = as.NewClientWithPolicy(clientPolicy, *host, *port)
				Expect(err).ToNot(HaveOccurred())
				defer client.Close()

				node := client.GetNodes()[0]

				// get a few connections at once
				var conns []*as.Connection
				for i := 0; i < 4; i++ {
					// By(fmt.Sprintf("Retrieving conns i=%d", i))
					c, err := node.GetConnection(0)
					Expect(err).NotTo(HaveOccurred())
					Expect(c).NotTo(BeNil())
					Expect(c.IsConnected()).To(BeTrue())

					conns = append(conns, c)
				}

				// return them to the pool
				for _, c := range conns {
					node.PutConnection(c)
				}

				start := time.Now()
				estimatedDeadline := start.Add(clientPolicy.IdleTimeout)
				deadlineThreshold := clientPolicy.IdleTimeout / 10

				// make sure the same connections are all retrieved again
				checkCount := 0
				for estimatedDeadline.Sub(time.Now()) > deadlineThreshold {
					checkCount++
					// By(fmt.Sprintf("Retrieving conns2 checkCount=%d", checkCount))
					var conns2 []*as.Connection
					for i := 0; i < len(conns); i++ {
						c, err := node.GetConnection(0)
						Expect(err).NotTo(HaveOccurred())
						Expect(c).NotTo(BeNil())
						Expect(c.IsConnected()).To(BeTrue())
						Expect(conns).To(ContainElement(c))
						Expect(conns2).NotTo(ContainElement(c))

						conns2 = append(conns2, c)
					}

					// just put them in the pool
					for _, c := range conns2 {
						node.PutConnection(c)
					}

					time.Sleep(time.Millisecond)
				}

				// we should be called lots of times
				Expect(checkCount).To(BeNumerically(">", 500))

				// sleep again until all connections are all idle
				<-time.After(2 * clientPolicy.IdleTimeout)

				// get connections again, making sure they are all new
				var conns3 []*as.Connection
				for i := 0; i < len(conns); i++ {
					// By(fmt.Sprintf("Retrieving conns3 i=%d", i))
					c, err := node.GetConnection(0)
					Expect(err).NotTo(HaveOccurred())
					Expect(c).NotTo(BeNil())
					Expect(c.IsConnected()).To(BeTrue())

					Expect(conns).NotTo(ContainElement(c))
					Expect(conns3).NotTo(ContainElement(c))

					conns3 = append(conns3, c)
				}

				// refresh and return them to the pool
				for _, c := range conns {
					Expect(c.IsConnected()).To(BeFalse())
				}

				// don't forget to close connections
				for _, c := range conns3 {
					c.Close()
				}
			})

			It("must delay the connection from becoming idle if it is put back in the queue", func() {
				clientPolicy := as.NewClientPolicy()
				clientPolicy.IdleTimeout = 1000 * time.Millisecond
				clientPolicy.User = *user
				clientPolicy.Password = *password

				client, err = as.NewClientWithPolicy(clientPolicy, *host, *port)
				Expect(err).ToNot(HaveOccurred())
				defer client.Close()

				node := client.GetNodes()[0]

				deadlineThreshold := clientPolicy.IdleTimeout / 10

				// By("Retrieving c")
				c, err := node.GetConnection(0)
				Expect(err).NotTo(HaveOccurred())
				Expect(c).NotTo(BeNil())
				Expect(c.IsConnected()).To(BeTrue())
				node.PutConnection(c)

				// continuously refresh the connection just before it goes idle
				for i := 0; i < 5; i++ {
					time.Sleep(clientPolicy.IdleTimeout - deadlineThreshold)
					// By(fmt.Sprintf("Retrieving c2 i=%d", i))

					c2, err := node.GetConnection(0)
					Expect(err).NotTo(HaveOccurred())
					Expect(c2).NotTo(BeNil())
					Expect(c2).To(Equal(c))
					Expect(c2.IsConnected()).To(BeTrue())

					node.PutConnection(c2)
				}

				// wait about the required time to become idle
				<-time.After(2 * clientPolicy.IdleTimeout)

				// we should get a new connection
				c3, err := node.GetConnection(0)
				Expect(err).NotTo(HaveOccurred())
				Expect(c3).NotTo(BeNil())
				defer node.InvalidateConnection(c3)
				Expect(c3).ToNot(Equal(c))
				Expect(c3.IsConnected()).To(BeTrue())

				// the original connection should be closed
				Expect(c.IsConnected()).To(BeFalse())
			})

		})
	})
})
