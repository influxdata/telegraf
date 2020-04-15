// Copyright 2012-2019 The NATS Authors
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

package server

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	// Allow dynamic profiling.
	_ "net/http/pprof"

	"github.com/nats-io/jwt"
	"github.com/nats-io/nats-server/v2/logger"
	"github.com/nats-io/nkeys"
)

const (
	// Time to wait before starting closing clients when in LD mode.
	lameDuckModeDefaultInitialDelay = int64(10 * time.Second)

	// Interval for the first PING for non client connections.
	firstPingInterval = time.Second

	// This is for the first ping for client connections.
	firstClientPingInterval = 2 * time.Second
)

// Make this a variable so that we can change during tests
var lameDuckModeInitialDelay = int64(lameDuckModeDefaultInitialDelay)

// Info is the information sent to clients, routes, gateways, and leaf nodes,
// to help them understand information about this server.
type Info struct {
	ID                string   `json:"server_id"`
	Name              string   `json:"server_name"`
	Version           string   `json:"version"`
	Proto             int      `json:"proto"`
	GitCommit         string   `json:"git_commit,omitempty"`
	GoVersion         string   `json:"go"`
	Host              string   `json:"host"`
	Port              int      `json:"port"`
	AuthRequired      bool     `json:"auth_required,omitempty"`
	TLSRequired       bool     `json:"tls_required,omitempty"`
	TLSVerify         bool     `json:"tls_verify,omitempty"`
	MaxPayload        int32    `json:"max_payload"`
	IP                string   `json:"ip,omitempty"`
	CID               uint64   `json:"client_id,omitempty"`
	Nonce             string   `json:"nonce,omitempty"`
	Cluster           string   `json:"cluster,omitempty"`
	ClientConnectURLs []string `json:"connect_urls,omitempty"` // Contains URLs a client can connect to.

	// Route Specific
	Import *SubjectPermission `json:"import,omitempty"`
	Export *SubjectPermission `json:"export,omitempty"`

	// Gateways Specific
	Gateway           string   `json:"gateway,omitempty"`             // Name of the origin Gateway (sent by gateway's INFO)
	GatewayURLs       []string `json:"gateway_urls,omitempty"`        // Gateway URLs in the originating cluster (sent by gateway's INFO)
	GatewayURL        string   `json:"gateway_url,omitempty"`         // Gateway URL on that server (sent by route's INFO)
	GatewayCmd        byte     `json:"gateway_cmd,omitempty"`         // Command code for the receiving server to know what to do
	GatewayCmdPayload []byte   `json:"gateway_cmd_payload,omitempty"` // Command payload when needed
	GatewayNRP        bool     `json:"gateway_nrp,omitempty"`         // Uses new $GNR. prefix for mapped replies

	// LeafNode Specific
	LeafNodeURLs []string `json:"leafnode_urls,omitempty"` // LeafNode URLs that the server can reconnect to.
}

// Server is our main struct.
type Server struct {
	gcid uint64
	stats
	mu               sync.Mutex
	kp               nkeys.KeyPair
	prand            *rand.Rand
	info             Info
	configFile       string
	optsMu           sync.RWMutex
	opts             *Options
	running          bool
	shutdown         bool
	listener         net.Listener
	gacc             *Account
	sys              *internal
	accounts         sync.Map
	tmpAccounts      sync.Map // Temporarily stores accounts that are being built
	activeAccounts   int32
	accResolver      AccountResolver
	clients          map[uint64]*client
	routes           map[uint64]*client
	routesByHash     sync.Map
	hash             []byte
	remotes          map[string]*client
	leafs            map[uint64]*client
	users            map[string]*User
	nkeys            map[string]*NkeyUser
	totalClients     uint64
	closed           *closedRingBuffer
	done             chan bool
	start            time.Time
	http             net.Listener
	httpHandler      http.Handler
	profiler         net.Listener
	httpReqStats     map[string]uint64
	routeListener    net.Listener
	routeInfo        Info
	routeInfoJSON    []byte
	leafNodeListener net.Listener
	leafNodeInfo     Info
	leafNodeInfoJSON []byte
	leafNodeOpts     struct {
		resolver    netResolver
		dialTimeout time.Duration
	}

	quitCh chan struct{}

	// Tracking Go routines
	grMu         sync.Mutex
	grTmpClients map[uint64]*client
	grRunning    bool
	grWG         sync.WaitGroup // to wait on various go routines

	cproto     int64     // number of clients supporting async INFO
	configTime time.Time // last time config was loaded

	logging struct {
		sync.RWMutex
		logger Logger
		trace  int32
		debug  int32
	}

	clientConnectURLs []string

	// Used internally for quick look-ups.
	clientConnectURLsMap map[string]struct{}

	lastCURLsUpdate int64

	// For Gateways
	gatewayListener net.Listener // Accept listener
	gateway         *srvGateway

	// Used by tests to check that http.Servers do
	// not set any timeout.
	monitoringServer *http.Server
	profilingServer  *http.Server

	// LameDuck mode
	ldm   bool
	ldmCh chan bool

	// Trusted public operator keys.
	trustedKeys []string

	// We use this to minimize mem copies for request to monitoring
	// endpoint /varz (when it comes from http).
	varzMu sync.Mutex
	varz   *Varz
	// This is set during a config reload if we detect that we have
	// added/removed routes. The monitoring code then check that
	// to know if it should update the cluster's URLs array.
	varzUpdateRouteURLs bool

	// Keeps a sublist of of subscriptions attached to leafnode connections
	// for the $GNR.*.*.*.> subject so that a server can send back a mapped
	// gateway reply.
	gwLeafSubs *Sublist

	// Used for expiration of mapped GW replies
	gwrm struct {
		w  int32
		ch chan time.Duration
		m  sync.Map
	}
}

// Make sure all are 64bits for atomic use
type stats struct {
	inMsgs        int64
	outMsgs       int64
	inBytes       int64
	outBytes      int64
	slowConsumers int64
}

// New will setup a new server struct after parsing the options.
// DEPRECATED: Use NewServer(opts)
func New(opts *Options) *Server {
	s, _ := NewServer(opts)
	return s
}

// NewServer will setup a new server struct after parsing the options.
// Could return an error if options can not be validated.
func NewServer(opts *Options) (*Server, error) {
	setBaselineOptions(opts)

	// Process TLS options, including whether we require client certificates.
	tlsReq := opts.TLSConfig != nil
	verify := (tlsReq && opts.TLSConfig.ClientAuth == tls.RequireAndVerifyClientCert)

	// Created server's nkey identity.
	kp, _ := nkeys.CreateServer()
	pub, _ := kp.PublicKey()

	serverName := pub
	if opts.ServerName != "" {
		serverName = opts.ServerName
	}

	// Validate some options. This is here because we cannot assume that
	// server will always be started with configuration parsing (that could
	// report issues). Its options can be (incorrectly) set by hand when
	// server is embedded. If there is an error, return nil.
	if err := validateOptions(opts); err != nil {
		return nil, err
	}

	// If there is an URL account resolver, do basic test to see if anyone is home
	if ar := opts.AccountResolver; ar != nil {
		if ur, ok := ar.(*URLAccResolver); ok {
			if _, err := ur.Fetch(""); err != nil {
				return nil, err
			}
		}
	}

	info := Info{
		ID:           pub,
		Version:      VERSION,
		Proto:        PROTO,
		GitCommit:    gitCommit,
		GoVersion:    runtime.Version(),
		Name:         serverName,
		Host:         opts.Host,
		Port:         opts.Port,
		AuthRequired: false,
		TLSRequired:  tlsReq,
		TLSVerify:    verify,
		MaxPayload:   opts.MaxPayload,
	}

	now := time.Now()
	s := &Server{
		kp:         kp,
		configFile: opts.ConfigFile,
		info:       info,
		prand:      rand.New(rand.NewSource(time.Now().UnixNano())),
		opts:       opts,
		done:       make(chan bool, 1),
		start:      now,
		configTime: now,
		gwLeafSubs: NewSublistWithCache(),
	}

	// Trusted root operator keys.
	if !s.processTrustedKeys() {
		return nil, fmt.Errorf("Error processing trusted operator keys")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure that non-exported options (used in tests) are properly set.
	s.setLeafNodeNonExportedOptions()

	// Used internally for quick look-ups.
	s.clientConnectURLsMap = make(map[string]struct{})

	// Call this even if there is no gateway defined. It will
	// initialize the structure so we don't have to check for
	// it to be nil or not in various places in the code.
	if err := s.newGateway(opts); err != nil {
		return nil, err
	}

	if s.gateway.enabled {
		s.info.Cluster = s.getGatewayName()
	}

	// This is normally done in the AcceptLoop, once the
	// listener has been created (possibly with random port),
	// but since some tests may expect the INFO to be properly
	// set after New(), let's do it now.
	s.setInfoHostPortAndGenerateJSON()

	// For tracking clients
	s.clients = make(map[uint64]*client)

	// For tracking closed clients.
	s.closed = newClosedRingBuffer(opts.MaxClosedClients)

	// For tracking connections that are not yet registered
	// in s.routes, but for which readLoop has started.
	s.grTmpClients = make(map[uint64]*client)

	// For tracking routes and their remote ids
	s.routes = make(map[uint64]*client)
	s.remotes = make(map[string]*client)

	// For tracking leaf nodes.
	s.leafs = make(map[uint64]*client)

	// Used to kick out all go routines possibly waiting on server
	// to shutdown.
	s.quitCh = make(chan struct{})

	// For tracking accounts
	if err := s.configureAccounts(); err != nil {
		return nil, err
	}

	// In local config mode, check that leafnode configuration
	// refers to account that exist.
	if len(opts.TrustedOperators) == 0 {
		checkAccountExists := func(accName string) error {
			if accName == _EMPTY_ {
				return nil
			}
			if _, ok := s.accounts.Load(accName); !ok {
				return fmt.Errorf("cannot find account %q specified in leafnode authorization", accName)
			}
			return nil
		}
		if err := checkAccountExists(opts.LeafNode.Account); err != nil {
			return nil, err
		}
		for _, lu := range opts.LeafNode.Users {
			if lu.Account == nil {
				continue
			}
			if err := checkAccountExists(lu.Account.Name); err != nil {
				return nil, err
			}
		}
		for _, r := range opts.LeafNode.Remotes {
			if r.LocalAccount == _EMPTY_ {
				continue
			}
			if _, ok := s.accounts.Load(r.LocalAccount); !ok {
				return nil, fmt.Errorf("no local account %q for remote leafnode", r.LocalAccount)
			}
		}
	}

	// Used to setup Authorization.
	s.configureAuthorization()

	// Start signal handler
	s.handleSignals()

	return s, nil
}

// ClientURL returns the URL used to connect clients. Helpful in testing
// when we designate a random client port (-1).
func (s *Server) ClientURL() string {
	// FIXME(dlc) - should we add in user and pass if defined single?
	opts := s.getOpts()
	scheme := "nats://"
	if opts.TLSConfig != nil {
		scheme = "tls://"
	}
	return fmt.Sprintf("%s%s:%d", scheme, opts.Host, opts.Port)
}

func validateOptions(o *Options) error {
	// Check that the trust configuration is correct.
	if err := validateTrustedOperators(o); err != nil {
		return err
	}
	// Check on leaf nodes which will require a system
	// account when gateways are also configured.
	if err := validateLeafNode(o); err != nil {
		return err
	}
	// Check that gateway is properly configured. Returns no error
	// if there is no gateway defined.
	return validateGatewayOptions(o)
}

func (s *Server) getOpts() *Options {
	s.optsMu.RLock()
	opts := s.opts
	s.optsMu.RUnlock()
	return opts
}

func (s *Server) setOpts(opts *Options) {
	s.optsMu.Lock()
	s.opts = opts
	s.optsMu.Unlock()
}

func (s *Server) globalAccount() *Account {
	s.mu.Lock()
	gacc := s.gacc
	s.mu.Unlock()
	return gacc
}

// Used to setup Accounts.
// Lock is held upon entry.
func (s *Server) configureAccounts() error {
	// Create global account.
	if s.gacc == nil {
		s.gacc = NewAccount(globalAccountName)
		s.registerAccountNoLock(s.gacc)
	}

	opts := s.opts

	// Check opts and walk through them. We need to copy them here
	// so that we do not keep a real one sitting in the options.
	for _, acc := range s.opts.Accounts {
		a := acc.shallowCopy()
		acc.sl = nil
		acc.clients = nil
		s.registerAccountNoLock(a)
	}

	// Now that we have this we need to remap any referenced accounts in
	// import or export maps to the new ones.
	swapApproved := func(ea *exportAuth) {
		for sub, a := range ea.approved {
			var acc *Account
			if v, ok := s.accounts.Load(a.Name); ok {
				acc = v.(*Account)
			}
			ea.approved[sub] = acc
		}
	}

	s.accounts.Range(func(k, v interface{}) bool {
		acc := v.(*Account)
		// Exports
		for _, ea := range acc.exports.streams {
			if ea != nil {
				swapApproved(&ea.exportAuth)
			}
		}
		for _, ea := range acc.exports.services {
			if ea != nil {
				swapApproved(&ea.exportAuth)
			}
		}
		// Imports
		for _, si := range acc.imports.streams {
			if v, ok := s.accounts.Load(si.acc.Name); ok {
				si.acc = v.(*Account)
			}
		}
		for _, si := range acc.imports.services {
			if v, ok := s.accounts.Load(si.acc.Name); ok {
				si.acc = v.(*Account)
			}
		}
		return true
	})

	// Check for configured account resolvers.
	if err := s.configureResolver(); err != nil {
		return err
	}

	// Set the system account if it was configured.
	if opts.SystemAccount != _EMPTY_ {
		// Lock may be acquired in lookupAccount, so release to call lookupAccount.
		s.mu.Unlock()
		_, err := s.lookupAccount(opts.SystemAccount)
		s.mu.Lock()
		if err != nil {
			return fmt.Errorf("error resolving system account: %v", err)
		}
	}

	return nil
}

// Setup the memory resolver, make sure the JWTs are properly formed but do not
// enforce expiration etc.
func (s *Server) configureResolver() error {
	opts := s.opts
	s.accResolver = opts.AccountResolver
	if opts.AccountResolver != nil && len(opts.resolverPreloads) > 0 {
		if _, ok := s.accResolver.(*MemAccResolver); !ok {
			return fmt.Errorf("resolver preloads only available for resolver type MEM")
		}
		for k, v := range opts.resolverPreloads {
			_, err := jwt.DecodeAccountClaims(v)
			if err != nil {
				return fmt.Errorf("preload account error for %q: %v", k, err)
			}
			s.accResolver.Store(k, v)
		}
	}
	return nil
}

// This will check preloads for validation issues.
func (s *Server) checkResolvePreloads() {
	opts := s.getOpts()
	// We can just check the read-only opts versions here, that way we do not need
	// to grab server lock or access s.accResolver.
	for k, v := range opts.resolverPreloads {
		claims, err := jwt.DecodeAccountClaims(v)
		if err != nil {
			s.Errorf("Preloaded account [%s] not valid", k)
		}
		// Check if it is expired.
		vr := jwt.CreateValidationResults()
		claims.Validate(vr)
		if vr.IsBlocking(true) {
			s.Warnf("Account [%s] has validation issues:", k)
			for _, v := range vr.Issues {
				s.Warnf("  - %s", v.Description)
			}
		}
	}
}

func (s *Server) generateRouteInfoJSON() {
	// New proto wants a nonce.
	var raw [nonceLen]byte
	nonce := raw[:]
	s.generateNonce(nonce)
	s.routeInfo.Nonce = string(nonce)
	b, _ := json.Marshal(s.routeInfo)
	pcs := [][]byte{[]byte("INFO"), b, []byte(CR_LF)}
	s.routeInfoJSON = bytes.Join(pcs, []byte(" "))
}

// isTrustedIssuer will check that the issuer is a trusted public key.
// This is used to make sure an account was signed by a trusted operator.
func (s *Server) isTrustedIssuer(issuer string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	// If we are not running in trusted mode and there is no issuer, that is ok.
	if len(s.trustedKeys) == 0 && issuer == "" {
		return true
	}
	for _, tk := range s.trustedKeys {
		if tk == issuer {
			return true
		}
	}
	return false
}

// processTrustedKeys will process binary stamped and
// options-based trusted nkeys. Returns success.
func (s *Server) processTrustedKeys() bool {
	if trustedKeys != "" && !s.initStampedTrustedKeys() {
		return false
	} else if s.opts.TrustedKeys != nil {
		for _, key := range s.opts.TrustedKeys {
			if !nkeys.IsValidPublicOperatorKey(key) {
				return false
			}
		}
		s.trustedKeys = s.opts.TrustedKeys
	}
	return true
}

// checkTrustedKeyString will check that the string is a valid array
// of public operator nkeys.
func checkTrustedKeyString(keys string) []string {
	tks := strings.Fields(keys)
	if len(tks) == 0 {
		return nil
	}
	// Walk all the keys and make sure they are valid.
	for _, key := range tks {
		if !nkeys.IsValidPublicOperatorKey(key) {
			return nil
		}
	}
	return tks
}

// initStampedTrustedKeys will check the stamped trusted keys
// and will set the server field 'trustedKeys'. Returns whether
// it succeeded or not.
func (s *Server) initStampedTrustedKeys() bool {
	// Check to see if we have an override in options, which will cause us to fail.
	if len(s.opts.TrustedKeys) > 0 {
		return false
	}
	tks := checkTrustedKeyString(trustedKeys)
	if len(tks) == 0 {
		return false
	}
	s.trustedKeys = tks
	return true
}

// PrintAndDie is exported for access in other packages.
func PrintAndDie(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

// PrintServerAndExit will print our version and exit.
func PrintServerAndExit() {
	fmt.Printf("nats-server: v%s\n", VERSION)
	os.Exit(0)
}

// ProcessCommandLineArgs takes the command line arguments
// validating and setting flags for handling in case any
// sub command was present.
func ProcessCommandLineArgs(cmd *flag.FlagSet) (showVersion bool, showHelp bool, err error) {
	if len(cmd.Args()) > 0 {
		arg := cmd.Args()[0]
		switch strings.ToLower(arg) {
		case "version":
			return true, false, nil
		case "help":
			return false, true, nil
		default:
			return false, false, fmt.Errorf("unrecognized command: %q", arg)
		}
	}

	return false, false, nil
}

// Protected check on running state
func (s *Server) isRunning() bool {
	s.mu.Lock()
	running := s.running
	s.mu.Unlock()
	return running
}

func (s *Server) logPid() error {
	pidStr := strconv.Itoa(os.Getpid())
	return ioutil.WriteFile(s.getOpts().PidFile, []byte(pidStr), 0660)
}

// NewAccountsAllowed returns whether or not new accounts can be created on the fly.
func (s *Server) NewAccountsAllowed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.opts.AllowNewAccounts
}

// numReservedAccounts will return the number of reserved accounts configured in the server.
// Currently this is 1 for the global default service.
func (s *Server) numReservedAccounts() int {
	return 1
}

// NumActiveAccounts reports number of active accounts on this server.
func (s *Server) NumActiveAccounts() int32 {
	return atomic.LoadInt32(&s.activeAccounts)
}

// incActiveAccounts() just adds one under lock.
func (s *Server) incActiveAccounts() {
	atomic.AddInt32(&s.activeAccounts, 1)
}

// decActiveAccounts() just subtracts one under lock.
func (s *Server) decActiveAccounts() {
	atomic.AddInt32(&s.activeAccounts, -1)
}

// This should be used for testing only. Will be slow since we have to
// range over all accounts in the sync.Map to count.
func (s *Server) numAccounts() int {
	count := 0
	s.mu.Lock()
	s.accounts.Range(func(k, v interface{}) bool {
		count++
		return true
	})
	s.mu.Unlock()
	return count
}

// NumLoadedAccounts returns the number of loaded accounts.
func (s *Server) NumLoadedAccounts() int {
	return s.numAccounts()
}

// LookupOrRegisterAccount will return the given account if known or create a new entry.
func (s *Server) LookupOrRegisterAccount(name string) (account *Account, isNew bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.accounts.Load(name); ok {
		return v.(*Account), false
	}
	acc := NewAccount(name)
	s.registerAccountNoLock(acc)
	return acc, true
}

// RegisterAccount will register an account. The account must be new
// or this call will fail.
func (s *Server) RegisterAccount(name string) (*Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.accounts.Load(name); ok {
		return nil, ErrAccountExists
	}
	acc := NewAccount(name)
	s.registerAccountNoLock(acc)
	return acc, nil
}

// SetSystemAccount will set the internal system account.
// If root operators are present it will also check validity.
func (s *Server) SetSystemAccount(accName string) error {
	// Lookup from sync.Map first.
	if v, ok := s.accounts.Load(accName); ok {
		return s.setSystemAccount(v.(*Account))
	}

	// If we are here we do not have local knowledge of this account.
	// Do this one by hand to return more useful error.
	ac, jwt, err := s.fetchAccountClaims(accName)
	if err != nil {
		return err
	}
	acc := s.buildInternalAccount(ac)
	acc.claimJWT = jwt
	// Due to race, we need to make sure that we are not
	// registering twice.
	if racc := s.registerAccount(acc); racc != nil {
		return nil
	}
	return s.setSystemAccount(acc)
}

// SystemAccount returns the system account if set.
func (s *Server) SystemAccount() *Account {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sys != nil {
		return s.sys.account
	}
	return nil
}

// For internal sends.
const internalSendQLen = 4096

// Assign a system account. Should only be called once.
// This sets up a server to send and receive messages from
// inside the server itself.
func (s *Server) setSystemAccount(acc *Account) error {
	if acc == nil {
		return ErrMissingAccount
	}
	// Don't try to fix this here.
	if acc.IsExpired() {
		return ErrAccountExpired
	}
	// If we are running with trusted keys for an operator
	// make sure we check the account is legit.
	if !s.isTrustedIssuer(acc.Issuer) {
		return ErrAccountValidation
	}

	s.mu.Lock()

	if s.sys != nil {
		s.mu.Unlock()
		return ErrAccountExists
	}

	// This is here in an attempt to quiet the race detector and not have to place
	// locks on fast path for inbound messages and checking service imports.
	acc.mu.Lock()
	if acc.imports.services == nil {
		acc.imports.services = make(map[string]*serviceImport)
	}
	acc.mu.Unlock()

	now := time.Now()
	s.sys = &internal{
		account: acc,
		client:  &client{srv: s, kind: SYSTEM, opts: internalOpts, msubs: -1, mpay: -1, start: now, last: now},
		seq:     1,
		sid:     1,
		servers: make(map[string]*serverUpdate),
		subs:    make(map[string]msgHandler),
		replies: make(map[string]msgHandler),
		sendq:   make(chan *pubMsg, internalSendQLen),
		statsz:  eventsHBInterval,
		orphMax: 5 * eventsHBInterval,
		chkOrph: 3 * eventsHBInterval,
	}
	s.sys.client.initClient()
	s.sys.client.echo = false
	s.sys.wg.Add(1)
	s.mu.Unlock()

	// Register with the account.
	s.sys.client.registerWithAccount(acc)

	// Start our internal loop to serialize outbound messages.
	// We do our own wg here since we will stop first during shutdown.
	go s.internalSendLoop(&s.sys.wg)

	// Start up our general subscriptions
	s.initEventTracking()

	// Track for dead remote servers.
	s.wrapChk(s.startRemoteServerSweepTimer)()

	// Send out statsz updates periodically.
	s.wrapChk(s.startStatszTimer)()

	return nil
}

func (s *Server) systemAccount() *Account {
	var sacc *Account
	s.mu.Lock()
	if s.sys != nil {
		sacc = s.sys.account
	}
	s.mu.Unlock()
	return sacc
}

// Determine if accounts should track subscriptions for
// efficient propagation.
// Lock should be held on entry.
func (s *Server) shouldTrackSubscriptions() bool {
	return (s.opts.Cluster.Port != 0 || s.opts.Gateway.Port != 0)
}

// Invokes registerAccountNoLock under the protection of the server lock.
// That is, server lock is acquired/released in this function.
// See registerAccountNoLock for comment on returned value.
func (s *Server) registerAccount(acc *Account) *Account {
	s.mu.Lock()
	racc := s.registerAccountNoLock(acc)
	s.mu.Unlock()
	return racc
}

// Helper to set the sublist based on preferences.
func (s *Server) setAccountSublist(acc *Account) {
	if acc != nil && acc.sl == nil {
		opts := s.getOpts()
		if opts != nil && opts.NoSublistCache {
			acc.sl = NewSublistNoCache()
		} else {
			acc.sl = NewSublistWithCache()
		}
	}
}

// Registers an account in the server.
// Due to some locking considerations, we may end-up trying
// to register the same account twice. This function will
// then return the already registered account.
// Lock should be held on entry.
func (s *Server) registerAccountNoLock(acc *Account) *Account {
	// We are under the server lock. Lookup from map, if present
	// return existing account.
	if a, _ := s.accounts.Load(acc.Name); a != nil {
		s.tmpAccounts.Delete(acc.Name)
		return a.(*Account)
	}
	// Finish account setup and store.
	s.setAccountSublist(acc)
	if acc.maxnae == 0 {
		acc.maxnae = DEFAULT_MAX_ACCOUNT_AE_RESPONSE_MAPS
	}
	if acc.maxaettl == 0 {
		acc.maxaettl = DEFAULT_TTL_AE_RESPONSE_MAP
	}
	if acc.maxnrm == 0 {
		acc.maxnrm = DEFAULT_MAX_ACCOUNT_INTERNAL_RESPONSE_MAPS
	}
	if acc.clients == nil {
		acc.clients = make(map[*client]*client)
	}
	// If we are capable of routing we will track subscription
	// information for efficient interest propagation.
	// During config reload, it is possible that account was
	// already created (global account), so use locking and
	// make sure we create only if needed.
	acc.mu.Lock()
	// TODO(dlc)- Double check that we need this for GWs.
	if acc.rm == nil && s.opts != nil && s.shouldTrackSubscriptions() {
		acc.rm = make(map[string]int32)
		acc.lqws = make(map[string]int32)
	}
	acc.srv = s
	acc.mu.Unlock()
	s.accounts.Store(acc.Name, acc)
	s.tmpAccounts.Delete(acc.Name)
	s.enableAccountTracking(acc)
	return nil
}

// lookupAccount is a function to return the account structure
// associated with an account name.
// Lock MUST NOT be held upon entry.
func (s *Server) lookupAccount(name string) (*Account, error) {
	var acc *Account
	if v, ok := s.accounts.Load(name); ok {
		acc = v.(*Account)
	} else if v, ok := s.tmpAccounts.Load(name); ok {
		acc = v.(*Account)
	}
	if acc != nil {
		// If we are expired and we have a resolver, then
		// return the latest information from the resolver.
		if acc.IsExpired() {
			s.Debugf("Requested account [%s] has expired", name)
			if s.AccountResolver() != nil {
				if err := s.updateAccount(acc); err != nil {
					// This error could mask expired, so just return expired here.
					return nil, ErrAccountExpired
				}
			} else {
				return nil, ErrAccountExpired
			}
		}
		return acc, nil
	}
	// If we have a resolver see if it can fetch the account.
	if s.AccountResolver() == nil {
		return nil, ErrMissingAccount
	}
	return s.fetchAccount(name)
}

// LookupAccount is a public function to return the account structure
// associated with name.
func (s *Server) LookupAccount(name string) (*Account, error) {
	return s.lookupAccount(name)
}

// This will fetch new claims and if found update the account with new claims.
// Lock MUST NOT be held upon entry.
func (s *Server) updateAccount(acc *Account) error {
	// TODO(dlc) - Make configurable
	if time.Since(acc.updated) < time.Second {
		s.Debugf("Requested account update for [%s] ignored, too soon", acc.Name)
		return ErrAccountResolverUpdateTooSoon
	}
	claimJWT, err := s.fetchRawAccountClaims(acc.Name)
	if err != nil {
		return err
	}
	return s.updateAccountWithClaimJWT(acc, claimJWT)
}

// updateAccountWithClaimJWT will check and apply the claim update.
// Lock MUST NOT be held upon entry.
func (s *Server) updateAccountWithClaimJWT(acc *Account, claimJWT string) error {
	if acc == nil {
		return ErrMissingAccount
	}
	acc.updated = time.Now()
	if acc.claimJWT != "" && acc.claimJWT == claimJWT {
		s.Debugf("Requested account update for [%s], same claims detected", acc.Name)
		return ErrAccountResolverSameClaims
	}
	accClaims, _, err := s.verifyAccountClaims(claimJWT)
	if err == nil && accClaims != nil {
		acc.claimJWT = claimJWT
		s.updateAccountClaims(acc, accClaims)
		return nil
	}
	return err
}

// fetchRawAccountClaims will grab raw account claims iff we have a resolver.
// Lock is NOT held upon entry.
func (s *Server) fetchRawAccountClaims(name string) (string, error) {
	accResolver := s.AccountResolver()
	if accResolver == nil {
		return "", ErrNoAccountResolver
	}
	// Need to do actual Fetch
	start := time.Now()
	claimJWT, err := accResolver.Fetch(name)
	fetchTime := time.Since(start)
	if fetchTime > time.Second {
		s.Warnf("Account [%s] fetch took %v", name, fetchTime)
	} else {
		s.Debugf("Account [%s] fetch took %v", name, fetchTime)
	}
	if err != nil {
		s.Warnf("Account fetch failed: %v", err)
		return "", err
	}
	return claimJWT, nil
}

// fetchAccountClaims will attempt to fetch new claims if a resolver is present.
// Lock is NOT held upon entry.
func (s *Server) fetchAccountClaims(name string) (*jwt.AccountClaims, string, error) {
	claimJWT, err := s.fetchRawAccountClaims(name)
	if err != nil {
		return nil, _EMPTY_, err
	}
	return s.verifyAccountClaims(claimJWT)
}

// verifyAccountClaims will decode and validate any account claims.
func (s *Server) verifyAccountClaims(claimJWT string) (*jwt.AccountClaims, string, error) {
	accClaims, err := jwt.DecodeAccountClaims(claimJWT)
	if err != nil {
		return nil, _EMPTY_, err
	}
	vr := jwt.CreateValidationResults()
	accClaims.Validate(vr)
	if vr.IsBlocking(true) {
		return nil, _EMPTY_, ErrAccountValidation
	}
	return accClaims, claimJWT, nil
}

// This will fetch an account from a resolver if defined.
// Lock is NOT held upon entry.
func (s *Server) fetchAccount(name string) (*Account, error) {
	accClaims, claimJWT, err := s.fetchAccountClaims(name)
	if accClaims != nil {
		acc := s.buildInternalAccount(accClaims)
		acc.claimJWT = claimJWT
		// Due to possible race, if registerAccount() returns a non
		// nil account, it means the same account was already
		// registered and we should use this one.
		if racc := s.registerAccount(acc); racc != nil {
			// Update with the new claims in case they are new.
			// Following call will return ErrAccountResolverSameClaims
			// if claims are the same.
			err = s.updateAccountWithClaimJWT(racc, claimJWT)
			if err != nil && err != ErrAccountResolverSameClaims {
				return nil, err
			}
			return racc, nil
		}
		return acc, nil
	}
	return nil, err
}

// Start up the server, this will block.
// Start via a Go routine if needed.
func (s *Server) Start() {
	s.Noticef("Starting nats-server version %s", VERSION)
	s.Debugf("Go build version %s", s.info.GoVersion)
	gc := gitCommit
	if gc == "" {
		gc = "not set"
	}
	s.Noticef("Git commit [%s]", gc)

	// Check for insecure configurations.op
	s.checkAuthforWarnings()

	// Avoid RACE between Start() and Shutdown()
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	s.grMu.Lock()
	s.grRunning = true
	s.grMu.Unlock()

	// Snapshot server options.
	opts := s.getOpts()

	hasOperators := len(opts.TrustedOperators) > 0
	if hasOperators {
		s.Noticef("Trusted Operators")
	}
	for _, opc := range opts.TrustedOperators {
		s.Noticef("  System  : %q", opc.Audience)
		s.Noticef("  Operator: %q", opc.Name)
		s.Noticef("  Issued  : %v", time.Unix(opc.IssuedAt, 0))
		s.Noticef("  Expires : %v", time.Unix(opc.Expires, 0))
	}
	if hasOperators && opts.SystemAccount == _EMPTY_ {
		s.Warnf("Trusted Operators should utilize a System Account")
	}

	// If we have a memory resolver, check the accounts here for validation exceptions.
	// This allows them to be logged right away vs when they are accessed via a client.
	if hasOperators && len(opts.resolverPreloads) > 0 {
		s.checkResolvePreloads()
	}

	// Log the pid to a file
	if opts.PidFile != _EMPTY_ {
		if err := s.logPid(); err != nil {
			PrintAndDie(fmt.Sprintf("Could not write pidfile: %v\n", err))
		}
	}

	// Start monitoring if needed
	if err := s.StartMonitoring(); err != nil {
		s.Fatalf("Can't start monitoring: %v", err)
		return
	}

	// Setup system account which will start eventing stack.
	if sa := opts.SystemAccount; sa != _EMPTY_ {
		if err := s.SetSystemAccount(sa); err != nil {
			s.Fatalf("Can't set system account: %v", err)
			return
		}
	}

	// Start expiration of mapped GW replies, regardless if
	// this server is configured with gateway or not.
	s.startGWReplyMapExpiration()

	// Start up gateway if needed. Do this before starting the routes, because
	// we want to resolve the gateway host:port so that this information can
	// be sent to other routes.
	if opts.Gateway.Port != 0 {
		s.startGateways()
	}

	// Start up listen if we want to accept leaf node connections.
	if opts.LeafNode.Port != 0 {
		// Spin up the accept loop if needed
		ch := make(chan struct{})
		go s.leafNodeAcceptLoop(ch)
		// This ensure that we have resolved or assigned the advertise
		// address for the leafnode listener. We need that in StartRouting().
		<-ch
	}

	// Solicit remote servers for leaf node connections.
	if len(opts.LeafNode.Remotes) > 0 {
		s.solicitLeafNodeRemotes(opts.LeafNode.Remotes)
	}

	// The Routing routine needs to wait for the client listen
	// port to be opened and potential ephemeral port selected.
	clientListenReady := make(chan struct{})

	// Start up routing as well if needed.
	if opts.Cluster.Port != 0 {
		s.startGoRoutine(func() {
			s.StartRouting(clientListenReady)
		})
	}

	// Pprof http endpoint for the profiler.
	if opts.ProfPort != 0 {
		s.StartProfiler()
	}

	if opts.PortsFileDir != _EMPTY_ {
		s.logPorts()
	}

	// Wait for clients.
	s.AcceptLoop(clientListenReady)
}

// Shutdown will shutdown the server instance by kicking out the AcceptLoop
// and closing all associated clients.
func (s *Server) Shutdown() {
	// Shutdown the eventing system as needed.
	// This is done first to send out any messages for
	// account status. We will also clean up any
	// eventing items associated with accounts.
	s.shutdownEventing()

	s.mu.Lock()
	// Prevent issues with multiple calls.
	if s.shutdown {
		s.mu.Unlock()
		return
	}
	s.Noticef("Initiating Shutdown...")

	opts := s.getOpts()

	s.shutdown = true
	s.running = false
	s.grMu.Lock()
	s.grRunning = false
	s.grMu.Unlock()

	conns := make(map[uint64]*client)

	// Copy off the clients
	for i, c := range s.clients {
		conns[i] = c
	}
	// Copy off the connections that are not yet registered
	// in s.routes, but for which the readLoop has started
	s.grMu.Lock()
	for i, c := range s.grTmpClients {
		conns[i] = c
	}
	s.grMu.Unlock()
	// Copy off the routes
	for i, r := range s.routes {
		conns[i] = r
	}
	// Copy off the gateways
	s.getAllGatewayConnections(conns)

	// Copy off the leaf nodes
	for i, c := range s.leafs {
		conns[i] = c
	}

	// Number of done channel responses we expect.
	doneExpected := 0

	// Kick client AcceptLoop()
	if s.listener != nil {
		doneExpected++
		s.listener.Close()
		s.listener = nil
	}

	// Kick leafnodes AcceptLoop()
	if s.leafNodeListener != nil {
		doneExpected++
		s.leafNodeListener.Close()
		s.leafNodeListener = nil
	}

	// Kick route AcceptLoop()
	if s.routeListener != nil {
		doneExpected++
		s.routeListener.Close()
		s.routeListener = nil
	}

	// Kick Gateway AcceptLoop()
	if s.gatewayListener != nil {
		doneExpected++
		s.gatewayListener.Close()
		s.gatewayListener = nil
	}

	// Kick HTTP monitoring if its running
	if s.http != nil {
		doneExpected++
		s.http.Close()
		s.http = nil
	}

	// Kick Profiling if its running
	if s.profiler != nil {
		doneExpected++
		s.profiler.Close()
	}

	s.mu.Unlock()

	// Release go routines that wait on that channel
	close(s.quitCh)

	// Close client and route connections
	for _, c := range conns {
		c.setNoReconnect()
		c.closeConnection(ServerShutdown)
	}

	// Block until the accept loops exit
	for doneExpected > 0 {
		<-s.done
		doneExpected--
	}

	// Wait for go routines to be done.
	s.grWG.Wait()

	if opts.PortsFileDir != _EMPTY_ {
		s.deletePortsFile(opts.PortsFileDir)
	}

	s.Noticef("Server Exiting..")
	// Close logger if applicable. It allows tests on Windows
	// to be able to do proper cleanup (delete log file).
	s.logging.RLock()
	log := s.logging.logger
	s.logging.RUnlock()
	if log != nil {
		if l, ok := log.(*logger.Logger); ok {
			l.Close()
		}
	}
}

// AcceptLoop is exported for easier testing.
func (s *Server) AcceptLoop(clr chan struct{}) {
	// If we were to exit before the listener is setup properly,
	// make sure we close the channel.
	defer func() {
		if clr != nil {
			close(clr)
		}
	}()

	// Snapshot server options.
	opts := s.getOpts()

	hp := net.JoinHostPort(opts.Host, strconv.Itoa(opts.Port))
	l, e := net.Listen("tcp", hp)
	if e != nil {
		s.Fatalf("Error listening on port: %s, %q", hp, e)
		return
	}
	s.Noticef("Listening for client connections on %s",
		net.JoinHostPort(opts.Host, strconv.Itoa(l.Addr().(*net.TCPAddr).Port)))

	// Alert of TLS enabled.
	if opts.TLSConfig != nil {
		s.Noticef("TLS required for client connections")
	}

	s.Noticef("Server id is %s", s.info.ID)
	s.Noticef("Server is ready")

	// Setup state that can enable shutdown
	s.mu.Lock()
	s.listener = l

	// If server was started with RANDOM_PORT (-1), opts.Port would be equal
	// to 0 at the beginning this function. So we need to get the actual port
	if opts.Port == 0 {
		// Write resolved port back to options.
		opts.Port = l.Addr().(*net.TCPAddr).Port
	}

	// Now that port has been set (if it was set to RANDOM), set the
	// server's info Host/Port with either values from Options or
	// ClientAdvertise. Also generate the JSON byte array.
	if err := s.setInfoHostPortAndGenerateJSON(); err != nil {
		s.Fatalf("Error setting server INFO with ClientAdvertise value of %s, err=%v", s.opts.ClientAdvertise, err)
		s.mu.Unlock()
		return
	}
	// Keep track of client connect URLs. We may need them later.
	s.clientConnectURLs = s.getClientConnectURLs()
	s.mu.Unlock()

	// Let the caller know that we are ready
	close(clr)
	clr = nil

	tmpDelay := ACCEPT_MIN_SLEEP

	for s.isRunning() {
		conn, err := l.Accept()
		if err != nil {
			if s.isLameDuckMode() {
				// Signal that we are not accepting new clients
				s.ldmCh <- true
				// Now wait for the Shutdown...
				<-s.quitCh
				return
			}
			tmpDelay = s.acceptError("Client", err, tmpDelay)
			continue
		}
		tmpDelay = ACCEPT_MIN_SLEEP
		s.startGoRoutine(func() {
			s.createClient(conn)
			s.grWG.Done()
		})
	}
	s.done <- true
}

// This function sets the server's info Host/Port based on server Options.
// Note that this function may be called during config reload, this is why
// Host/Port may be reset to original Options if the ClientAdvertise option
// is not set (since it may have previously been).
// The function then generates the server infoJSON.
func (s *Server) setInfoHostPortAndGenerateJSON() error {
	// When this function is called, opts.Port is set to the actual listen
	// port (if option was originally set to RANDOM), even during a config
	// reload. So use of s.opts.Port is safe.
	if s.opts.ClientAdvertise != "" {
		h, p, err := parseHostPort(s.opts.ClientAdvertise, s.opts.Port)
		if err != nil {
			return err
		}
		s.info.Host = h
		s.info.Port = p
	} else {
		s.info.Host = s.opts.Host
		s.info.Port = s.opts.Port
	}
	return nil
}

// StartProfiler is called to enable dynamic profiling.
func (s *Server) StartProfiler() {
	// Snapshot server options.
	opts := s.getOpts()

	port := opts.ProfPort

	// Check for Random Port
	if port == -1 {
		port = 0
	}

	hp := net.JoinHostPort(opts.Host, strconv.Itoa(port))

	l, err := net.Listen("tcp", hp)
	s.Noticef("profiling port: %d", l.Addr().(*net.TCPAddr).Port)

	if err != nil {
		s.Fatalf("error starting profiler: %s", err)
	}

	srv := &http.Server{
		Addr:           hp,
		Handler:        http.DefaultServeMux,
		MaxHeaderBytes: 1 << 20,
	}

	s.mu.Lock()
	s.profiler = l
	s.profilingServer = srv
	s.mu.Unlock()

	// Enable blocking profile
	runtime.SetBlockProfileRate(1)

	go func() {
		// if this errors out, it's probably because the server is being shutdown
		err := srv.Serve(l)
		if err != nil {
			s.mu.Lock()
			shutdown := s.shutdown
			s.mu.Unlock()
			if !shutdown {
				s.Fatalf("error starting profiler: %s", err)
			}
		}
		srv.Close()
		s.done <- true
	}()
}

// StartHTTPMonitoring will enable the HTTP monitoring port.
// DEPRECATED: Should use StartMonitoring.
func (s *Server) StartHTTPMonitoring() {
	s.startMonitoring(false)
}

// StartHTTPSMonitoring will enable the HTTPS monitoring port.
// DEPRECATED: Should use StartMonitoring.
func (s *Server) StartHTTPSMonitoring() {
	s.startMonitoring(true)
}

// StartMonitoring starts the HTTP or HTTPs server if needed.
func (s *Server) StartMonitoring() error {
	// Snapshot server options.
	opts := s.getOpts()

	// Specifying both HTTP and HTTPS ports is a misconfiguration
	if opts.HTTPPort != 0 && opts.HTTPSPort != 0 {
		return fmt.Errorf("can't specify both HTTP (%v) and HTTPs (%v) ports", opts.HTTPPort, opts.HTTPSPort)
	}
	var err error
	if opts.HTTPPort != 0 {
		err = s.startMonitoring(false)
	} else if opts.HTTPSPort != 0 {
		if opts.TLSConfig == nil {
			return fmt.Errorf("TLS cert and key required for HTTPS")
		}
		err = s.startMonitoring(true)
	}
	return err
}

// HTTP endpoints
const (
	RootPath     = "/"
	VarzPath     = "/varz"
	ConnzPath    = "/connz"
	RoutezPath   = "/routez"
	GatewayzPath = "/gatewayz"
	LeafzPath    = "/leafz"
	SubszPath    = "/subsz"
	StackszPath  = "/stacksz"
)

// Start the monitoring server
func (s *Server) startMonitoring(secure bool) error {
	// Snapshot server options.
	opts := s.getOpts()

	// Used to track HTTP requests
	s.httpReqStats = map[string]uint64{
		RootPath:     0,
		VarzPath:     0,
		ConnzPath:    0,
		RoutezPath:   0,
		GatewayzPath: 0,
		SubszPath:    0,
	}

	var (
		hp           string
		err          error
		httpListener net.Listener
		port         int
	)

	monitorProtocol := "http"

	if secure {
		monitorProtocol += "s"
		port = opts.HTTPSPort
		if port == -1 {
			port = 0
		}
		hp = net.JoinHostPort(opts.HTTPHost, strconv.Itoa(port))
		config := opts.TLSConfig.Clone()
		config.ClientAuth = tls.NoClientCert
		httpListener, err = tls.Listen("tcp", hp, config)

	} else {
		port = opts.HTTPPort
		if port == -1 {
			port = 0
		}
		hp = net.JoinHostPort(opts.HTTPHost, strconv.Itoa(port))
		httpListener, err = net.Listen("tcp", hp)
	}

	if err != nil {
		return fmt.Errorf("can't listen to the monitor port: %v", err)
	}

	s.Noticef("Starting %s monitor on %s", monitorProtocol,
		net.JoinHostPort(opts.HTTPHost, strconv.Itoa(httpListener.Addr().(*net.TCPAddr).Port)))

	mux := http.NewServeMux()

	// Root
	mux.HandleFunc(RootPath, s.HandleRoot)
	// Varz
	mux.HandleFunc(VarzPath, s.HandleVarz)
	// Connz
	mux.HandleFunc(ConnzPath, s.HandleConnz)
	// Routez
	mux.HandleFunc(RoutezPath, s.HandleRoutez)
	// Gatewayz
	mux.HandleFunc(GatewayzPath, s.HandleGatewayz)
	// Leafz
	mux.HandleFunc(LeafzPath, s.HandleLeafz)
	// Subz
	mux.HandleFunc(SubszPath, s.HandleSubsz)
	// Subz alias for backwards compatibility
	mux.HandleFunc("/subscriptionsz", s.HandleSubsz)
	// Stacksz
	mux.HandleFunc(StackszPath, s.HandleStacksz)

	// Do not set a WriteTimeout because it could cause cURL/browser
	// to return empty response or unable to display page if the
	// server needs more time to build the response.
	srv := &http.Server{
		Addr:           hp,
		Handler:        mux,
		MaxHeaderBytes: 1 << 20,
	}
	s.mu.Lock()
	s.http = httpListener
	s.httpHandler = mux
	s.monitoringServer = srv
	s.mu.Unlock()

	go func() {
		if err := srv.Serve(httpListener); err != nil {
			s.mu.Lock()
			shutdown := s.shutdown
			s.mu.Unlock()
			if !shutdown {
				s.Fatalf("Error starting monitor on %q: %v", hp, err)
			}
		}
		srv.Close()
		srv.Handler = nil
		s.mu.Lock()
		s.httpHandler = nil
		s.mu.Unlock()
		s.done <- true
	}()

	return nil
}

// HTTPHandler returns the http.Handler object used to handle monitoring
// endpoints. It will return nil if the server is not configured for
// monitoring, or if the server has not been started yet (Server.Start()).
func (s *Server) HTTPHandler() http.Handler {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.httpHandler
}

// Perform a conditional deep copy due to reference nature of ClientConnectURLs.
// If updates are made to Info, this function should be consulted and updated.
// Assume lock is held.
func (s *Server) copyInfo() Info {
	info := s.info
	if info.ClientConnectURLs != nil {
		info.ClientConnectURLs = make([]string, len(s.info.ClientConnectURLs))
		copy(info.ClientConnectURLs, s.info.ClientConnectURLs)
	}
	if s.nonceRequired() {
		// Nonce handling
		var raw [nonceLen]byte
		nonce := raw[:]
		s.generateNonce(nonce)
		info.Nonce = string(nonce)
	}
	return info
}

func (s *Server) createClient(conn net.Conn) *client {
	// Snapshot server options.
	opts := s.getOpts()

	maxPay := int32(opts.MaxPayload)
	maxSubs := int32(opts.MaxSubs)
	// For system, maxSubs of 0 means unlimited, so re-adjust here.
	if maxSubs == 0 {
		maxSubs = -1
	}
	now := time.Now()

	c := &client{srv: s, nc: conn, opts: defaultOpts, mpay: maxPay, msubs: maxSubs, start: now, last: now}

	c.registerWithAccount(s.globalAccount())

	// Grab JSON info string
	s.mu.Lock()
	info := s.copyInfo()
	c.nonce = []byte(info.Nonce)
	s.totalClients++
	s.mu.Unlock()

	// Grab lock
	c.mu.Lock()
	if info.AuthRequired {
		c.flags.set(expectConnect)
	}

	// Initialize
	c.initClient()

	c.Debugf("Client connection created")

	// Send our information.
	// Need to be sent in place since writeLoop cannot be started until
	// TLS handshake is done (if applicable).
	c.sendProtoNow(c.generateClientInfoJSON(info))

	// Unlock to register
	c.mu.Unlock()

	// Register with the server.
	s.mu.Lock()
	// If server is not running, Shutdown() may have already gathered the
	// list of connections to close. It won't contain this one, so we need
	// to bail out now otherwise the readLoop started down there would not
	// be interrupted. Skip also if in lame duck mode.
	if !s.running || s.ldm {
		s.mu.Unlock()
		return c
	}

	// If there is a max connections specified, check that adding
	// this new client would not push us over the max
	if opts.MaxConn > 0 && len(s.clients) >= opts.MaxConn {
		s.mu.Unlock()
		c.maxConnExceeded()
		return nil
	}
	s.clients[c.cid] = c
	s.mu.Unlock()

	// Re-Grab lock
	c.mu.Lock()

	// Check for TLS
	if info.TLSRequired {
		c.Debugf("Starting TLS client connection handshake")
		c.nc = tls.Server(c.nc, opts.TLSConfig)
		conn := c.nc.(*tls.Conn)

		// Setup the timeout
		ttl := secondsToDuration(opts.TLSTimeout)
		time.AfterFunc(ttl, func() { tlsTimeout(c, conn) })
		conn.SetReadDeadline(time.Now().Add(ttl))

		// Force handshake
		c.mu.Unlock()
		if err := conn.Handshake(); err != nil {
			c.Errorf("TLS handshake error: %v", err)
			c.closeConnection(TLSHandshakeError)
			return nil
		}
		// Reset the read deadline
		conn.SetReadDeadline(time.Time{})

		// Re-Grab lock
		c.mu.Lock()

		// Indicate that handshake is complete (used in monitoring)
		c.flags.set(handshakeComplete)
	}

	// The connection may have been closed
	if c.isClosed() {
		c.mu.Unlock()
		return c
	}

	// Check for Auth. We schedule this timer after the TLS handshake to avoid
	// the race where the timer fires during the handshake and causes the
	// server to write bad data to the socket. See issue #432.
	if info.AuthRequired {
		c.setAuthTimer(secondsToDuration(opts.AuthTimeout))
	}

	// Do final client initialization

	// Set the First Ping timer.
	s.setFirstPingTimer(c)

	// Spin up the read loop.
	s.startGoRoutine(func() { c.readLoop() })

	// Spin up the write loop.
	s.startGoRoutine(func() { c.writeLoop() })

	if info.TLSRequired {
		c.Debugf("TLS handshake complete")
		cs := c.nc.(*tls.Conn).ConnectionState()
		c.Debugf("TLS version %s, cipher suite %s", tlsVersion(cs.Version), tlsCipher(cs.CipherSuite))
	}

	c.mu.Unlock()

	return c
}

// This will save off a closed client in a ring buffer such that
// /connz can inspect. Useful for debugging, etc.
func (s *Server) saveClosedClient(c *client, nc net.Conn, reason ClosedState) {
	now := time.Now()

	s.accountDisconnectEvent(c, now, reason.String())

	c.mu.Lock()

	cc := &closedClient{}
	cc.fill(c, nc, now)
	cc.Stop = &now
	cc.Reason = reason.String()

	// Do subs, do not place by default in main ConnInfo
	if len(c.subs) > 0 {
		cc.subs = make([]string, 0, len(c.subs))
		for _, sub := range c.subs {
			cc.subs = append(cc.subs, string(sub.subject))
		}
	}
	// Hold user as well.
	cc.user = c.opts.Username
	// Hold account name if not the global account.
	if c.acc != nil && c.acc.Name != globalAccountName {
		cc.acc = c.acc.Name
	}
	c.mu.Unlock()

	// Place in the ring buffer
	s.mu.Lock()
	if s.closed != nil {
		s.closed.append(cc)
	}
	s.mu.Unlock()
}

// Adds the given array of urls to the server's INFO.ClientConnectURLs
// array. The server INFO JSON is regenerated.
// Note that a check is made to ensure that given URLs are not
// already present. So the INFO JSON is regenerated only if new ULRs
// were added.
// If there was a change, an INFO protocol is sent to registered clients
// that support async INFO protocols.
func (s *Server) addClientConnectURLsAndSendINFOToClients(urls []string) {
	s.updateServerINFOAndSendINFOToClients(urls, true)
}

// Removes the given array of urls from the server's INFO.ClientConnectURLs
// array. The server INFO JSON is regenerated if needed.
// If there was a change, an INFO protocol is sent to registered clients
// that support async INFO protocols.
func (s *Server) removeClientConnectURLsAndSendINFOToClients(urls []string) {
	s.updateServerINFOAndSendINFOToClients(urls, false)
}

// Updates the server's Info object with the given array of URLs and re-generate
// the infoJSON byte array, then send an (async) INFO protocol to clients that
// support it.
func (s *Server) updateServerINFOAndSendINFOToClients(urls []string, add bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Will be set to true if we alter the server's Info object.
	wasUpdated := false
	remove := !add
	for _, url := range urls {
		_, present := s.clientConnectURLsMap[url]
		if add && !present {
			s.clientConnectURLsMap[url] = struct{}{}
			wasUpdated = true
		} else if remove && present {
			delete(s.clientConnectURLsMap, url)
			wasUpdated = true
		}
	}
	if wasUpdated {
		// Recreate the info.ClientConnectURL array from the map
		s.info.ClientConnectURLs = s.info.ClientConnectURLs[:0]
		// Add this server client connect ULRs first...
		s.info.ClientConnectURLs = append(s.info.ClientConnectURLs, s.clientConnectURLs...)
		for url := range s.clientConnectURLsMap {
			s.info.ClientConnectURLs = append(s.info.ClientConnectURLs, url)
		}
		// Update the time of this update
		s.lastCURLsUpdate = time.Now().UnixNano()
		// Send to all registered clients that support async INFO protocols.
		s.sendAsyncInfoToClients()
	}
}

// Handle closing down a connection when the handshake has timedout.
func tlsTimeout(c *client, conn *tls.Conn) {
	c.mu.Lock()
	closed := c.isClosed()
	c.mu.Unlock()
	// Check if already closed
	if closed {
		return
	}
	cs := conn.ConnectionState()
	if !cs.HandshakeComplete {
		c.Errorf("TLS handshake timeout")
		c.sendErr("Secure Connection - TLS Required")
		c.closeConnection(TLSHandshakeError)
	}
}

// Seems silly we have to write these
func tlsVersion(ver uint16) string {
	switch ver {
	case tls.VersionTLS10:
		return "1.0"
	case tls.VersionTLS11:
		return "1.1"
	case tls.VersionTLS12:
		return "1.2"
	}
	return fmt.Sprintf("Unknown [%x]", ver)
}

// We use hex here so we don't need multiple versions
func tlsCipher(cs uint16) string {
	name, present := cipherMapByID[cs]
	if present {
		return name
	}
	return fmt.Sprintf("Unknown [%x]", cs)
}

// Remove a client or route from our internal accounting.
func (s *Server) removeClient(c *client) {
	// kind is immutable, so can check without lock
	switch c.kind {
	case CLIENT:
		c.mu.Lock()
		cid := c.cid
		updateProtoInfoCount := false
		if c.kind == CLIENT && c.opts.Protocol >= ClientProtoInfo {
			updateProtoInfoCount = true
		}
		c.mu.Unlock()

		s.mu.Lock()
		delete(s.clients, cid)
		if updateProtoInfoCount {
			s.cproto--
		}
		s.mu.Unlock()
	case ROUTER:
		s.removeRoute(c)
	case GATEWAY:
		s.removeRemoteGatewayConnection(c)
	case LEAF:
		s.removeLeafNodeConnection(c)
	}
}

func (s *Server) removeFromTempClients(cid uint64) {
	s.grMu.Lock()
	delete(s.grTmpClients, cid)
	s.grMu.Unlock()
}

func (s *Server) addToTempClients(cid uint64, c *client) bool {
	added := false
	s.grMu.Lock()
	if s.grRunning {
		s.grTmpClients[cid] = c
		added = true
	}
	s.grMu.Unlock()
	return added
}

/////////////////////////////////////////////////////////////////
// These are some helpers for accounting in functional tests.
/////////////////////////////////////////////////////////////////

// NumRoutes will report the number of registered routes.
func (s *Server) NumRoutes() int {
	s.mu.Lock()
	nr := len(s.routes)
	s.mu.Unlock()
	return nr
}

// NumRemotes will report number of registered remotes.
func (s *Server) NumRemotes() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.remotes)
}

// NumLeafNodes will report number of leaf node connections.
func (s *Server) NumLeafNodes() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.leafs)
}

// NumClients will report the number of registered clients.
func (s *Server) NumClients() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.clients)
}

// GetClient will return the client associated with cid.
func (s *Server) GetClient(cid uint64) *client {
	return s.getClient(cid)
}

// getClient will return the client associated with cid.
func (s *Server) getClient(cid uint64) *client {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.clients[cid]
}

// GetLeafNode returns the leafnode associated with the cid.
func (s *Server) GetLeafNode(cid uint64) *client {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.leafs[cid]
}

// NumSubscriptions will report how many subscriptions are active.
func (s *Server) NumSubscriptions() uint32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.numSubscriptions()
}

// numSubscriptions will report how many subscriptions are active.
// Lock should be held.
func (s *Server) numSubscriptions() uint32 {
	var subs int
	s.accounts.Range(func(k, v interface{}) bool {
		acc := v.(*Account)
		if acc.sl != nil {
			subs += acc.TotalSubs()
		}
		return true
	})
	return uint32(subs)
}

// NumSlowConsumers will report the number of slow consumers.
func (s *Server) NumSlowConsumers() int64 {
	return atomic.LoadInt64(&s.slowConsumers)
}

// ConfigTime will report the last time the server configuration was loaded.
func (s *Server) ConfigTime() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.configTime
}

// Addr will return the net.Addr object for the current listener.
func (s *Server) Addr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

// MonitorAddr will return the net.Addr object for the monitoring listener.
func (s *Server) MonitorAddr() *net.TCPAddr {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.http == nil {
		return nil
	}
	return s.http.Addr().(*net.TCPAddr)
}

// ClusterAddr returns the net.Addr object for the route listener.
func (s *Server) ClusterAddr() *net.TCPAddr {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.routeListener == nil {
		return nil
	}
	return s.routeListener.Addr().(*net.TCPAddr)
}

// ProfilerAddr returns the net.Addr object for the profiler listener.
func (s *Server) ProfilerAddr() *net.TCPAddr {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.profiler == nil {
		return nil
	}
	return s.profiler.Addr().(*net.TCPAddr)
}

// ReadyForConnections returns `true` if the server is ready to accept clients
// and, if routing is enabled, route connections. If after the duration
// `dur` the server is still not ready, returns `false`.
func (s *Server) ReadyForConnections(dur time.Duration) bool {
	// Snapshot server options.
	opts := s.getOpts()

	end := time.Now().Add(dur)
	for time.Now().Before(end) {
		s.mu.Lock()
		ok := s.listener != nil && (opts.Cluster.Port == 0 || s.routeListener != nil) && (opts.Gateway.Name == "" || s.gatewayListener != nil)
		s.mu.Unlock()
		if ok {
			return true
		}
		time.Sleep(25 * time.Millisecond)
	}
	return false
}

// ID returns the server's ID
func (s *Server) ID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.info.ID
}

func (s *Server) startGoRoutine(f func()) {
	s.grMu.Lock()
	if s.grRunning {
		s.grWG.Add(1)
		go f()
	}
	s.grMu.Unlock()
}

func (s *Server) numClosedConns() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed.len()
}

func (s *Server) totalClosedConns() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed.totalConns()
}

func (s *Server) closedClients() []*closedClient {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed.closedClients()
}

// getClientConnectURLs returns suitable URLs for clients to connect to the listen
// port based on the server options' Host and Port. If the Host corresponds to
// "any" interfaces, this call returns the list of resolved IP addresses.
// If ClientAdvertise is set, returns the client advertise host and port.
// The server lock is assumed held on entry.
func (s *Server) getClientConnectURLs() []string {
	// Snapshot server options.
	opts := s.getOpts()

	urls := make([]string, 0, 1)

	// short circuit if client advertise is set
	if opts.ClientAdvertise != "" {
		// just use the info host/port. This is updated in s.New()
		urls = append(urls, net.JoinHostPort(s.info.Host, strconv.Itoa(s.info.Port)))
	} else {
		sPort := strconv.Itoa(opts.Port)
		_, ips, err := s.getNonLocalIPsIfHostIsIPAny(opts.Host, true)
		for _, ip := range ips {
			urls = append(urls, net.JoinHostPort(ip, sPort))
		}
		if err != nil || len(urls) == 0 {
			// We are here if s.opts.Host is not "0.0.0.0" nor "::", or if for some
			// reason we could not add any URL in the loop above.
			// We had a case where a Windows VM was hosed and would have err == nil
			// and not add any address in the array in the loop above, and we
			// ended-up returning 0.0.0.0, which is problematic for Windows clients.
			// Check for 0.0.0.0 or :: specifically, and ignore if that's the case.
			if opts.Host == "0.0.0.0" || opts.Host == "::" {
				s.Errorf("Address %q can not be resolved properly", opts.Host)
			} else {
				urls = append(urls, net.JoinHostPort(opts.Host, sPort))
			}
		}
	}
	return urls
}

// Returns an array of non local IPs if the provided host is
// 0.0.0.0 or ::. It returns the first resolved if `all` is
// false.
// The boolean indicate if the provided host was 0.0.0.0 (or ::)
// so that if the returned array is empty caller can decide
// what to do next.
func (s *Server) getNonLocalIPsIfHostIsIPAny(host string, all bool) (bool, []string, error) {
	ip := net.ParseIP(host)
	// If this is not an IP, we are done
	if ip == nil {
		return false, nil, nil
	}
	// If this is not 0.0.0.0 or :: we have nothing to do.
	if !ip.IsUnspecified() {
		return false, nil, nil
	}
	s.Debugf("Get non local IPs for %q", host)
	var ips []string
	ifaces, _ := net.Interfaces()
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			ipStr := ip.String()
			// Skip non global unicast addresses
			if !ip.IsGlobalUnicast() || ip.IsUnspecified() {
				ip = nil
				continue
			}
			s.Debugf(" ip=%s", ipStr)
			ips = append(ips, ipStr)
			if !all {
				break
			}
		}
	}
	return true, ips, nil
}

// if the ip is not specified, attempt to resolve it
func resolveHostPorts(addr net.Listener) []string {
	hostPorts := make([]string, 0)
	hp := addr.Addr().(*net.TCPAddr)
	port := strconv.Itoa(hp.Port)
	if hp.IP.IsUnspecified() {
		var ip net.IP
		ifaces, _ := net.Interfaces()
		for _, i := range ifaces {
			addrs, _ := i.Addrs()
			for _, addr := range addrs {
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
					hostPorts = append(hostPorts, net.JoinHostPort(ip.String(), port))
				case *net.IPAddr:
					ip = v.IP
					hostPorts = append(hostPorts, net.JoinHostPort(ip.String(), port))
				default:
					continue
				}
			}
		}
	} else {
		hostPorts = append(hostPorts, net.JoinHostPort(hp.IP.String(), port))
	}
	return hostPorts
}

// format the address of a net.Listener with a protocol
func formatURL(protocol string, addr net.Listener) []string {
	hostports := resolveHostPorts(addr)
	for i, hp := range hostports {
		hostports[i] = fmt.Sprintf("%s://%s", protocol, hp)
	}
	return hostports
}

// Ports describes URLs that the server can be contacted in
type Ports struct {
	Nats       []string `json:"nats,omitempty"`
	Monitoring []string `json:"monitoring,omitempty"`
	Cluster    []string `json:"cluster,omitempty"`
	Profile    []string `json:"profile,omitempty"`
}

// PortsInfo attempts to resolve all the ports. If after maxWait the ports are not
// resolved, it returns nil. Otherwise it returns a Ports struct
// describing ports where the server can be contacted
func (s *Server) PortsInfo(maxWait time.Duration) *Ports {
	if s.readyForListeners(maxWait) {
		opts := s.getOpts()

		s.mu.Lock()
		info := s.copyInfo()
		listener := s.listener
		httpListener := s.http
		clusterListener := s.routeListener
		profileListener := s.profiler
		s.mu.Unlock()

		ports := Ports{}

		if listener != nil {
			natsProto := "nats"
			if info.TLSRequired {
				natsProto = "tls"
			}
			ports.Nats = formatURL(natsProto, listener)
		}

		if httpListener != nil {
			monProto := "http"
			if opts.HTTPSPort != 0 {
				monProto = "https"
			}
			ports.Monitoring = formatURL(monProto, httpListener)
		}

		if clusterListener != nil {
			clusterProto := "nats"
			if opts.Cluster.TLSConfig != nil {
				clusterProto = "tls"
			}
			ports.Cluster = formatURL(clusterProto, clusterListener)
		}

		if profileListener != nil {
			ports.Profile = formatURL("http", profileListener)
		}

		return &ports
	}

	return nil
}

// Returns the portsFile. If a non-empty dirHint is provided, the dirHint
// path is used instead of the server option value
func (s *Server) portFile(dirHint string) string {
	dirname := s.getOpts().PortsFileDir
	if dirHint != "" {
		dirname = dirHint
	}
	if dirname == _EMPTY_ {
		return _EMPTY_
	}
	return filepath.Join(dirname, fmt.Sprintf("%s_%d.ports", filepath.Base(os.Args[0]), os.Getpid()))
}

// Delete the ports file. If a non-empty dirHint is provided, the dirHint
// path is used instead of the server option value
func (s *Server) deletePortsFile(hintDir string) {
	portsFile := s.portFile(hintDir)
	if portsFile != "" {
		if err := os.Remove(portsFile); err != nil {
			s.Errorf("Error cleaning up ports file %s: %v", portsFile, err)
		}
	}
}

// Writes a file with a serialized Ports to the specified ports_file_dir.
// The name of the file is `exename_pid.ports`, typically nats-server_pid.ports.
// if ports file is not set, this function has no effect
func (s *Server) logPorts() {
	opts := s.getOpts()
	portsFile := s.portFile(opts.PortsFileDir)
	if portsFile != _EMPTY_ {
		go func() {
			info := s.PortsInfo(5 * time.Second)
			if info == nil {
				s.Errorf("Unable to resolve the ports in the specified time")
				return
			}
			data, err := json.Marshal(info)
			if err != nil {
				s.Errorf("Error marshaling ports file: %v", err)
				return
			}
			if err := ioutil.WriteFile(portsFile, data, 0666); err != nil {
				s.Errorf("Error writing ports file (%s): %v", portsFile, err)
				return
			}

		}()
	}
}

// waits until a calculated list of listeners is resolved or a timeout
func (s *Server) readyForListeners(dur time.Duration) bool {
	end := time.Now().Add(dur)
	for time.Now().Before(end) {
		s.mu.Lock()
		listeners := s.serviceListeners()
		s.mu.Unlock()
		if len(listeners) == 0 {
			return false
		}

		ok := true
		for _, l := range listeners {
			if l == nil {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
		select {
		case <-s.quitCh:
			return false
		case <-time.After(25 * time.Millisecond):
			// continue - unable to select from quit - we are still running
		}
	}
	return false
}

// returns a list of listeners that are intended for the process
// if the entry is nil, the interface is yet to be resolved
func (s *Server) serviceListeners() []net.Listener {
	listeners := make([]net.Listener, 0)
	opts := s.getOpts()
	listeners = append(listeners, s.listener)
	if opts.Cluster.Port != 0 {
		listeners = append(listeners, s.routeListener)
	}
	if opts.HTTPPort != 0 || opts.HTTPSPort != 0 {
		listeners = append(listeners, s.http)
	}
	if opts.ProfPort != 0 {
		listeners = append(listeners, s.profiler)
	}
	return listeners
}

// Returns true if in lame duck mode.
func (s *Server) isLameDuckMode() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ldm
}

// This function will close the client listener then close the clients
// at some interval to avoid a reconnecting storm.
func (s *Server) lameDuckMode() {
	s.mu.Lock()
	// Check if there is actually anything to do
	if s.shutdown || s.ldm || s.listener == nil {
		s.mu.Unlock()
		return
	}
	s.Noticef("Entering lame duck mode, stop accepting new clients")
	s.ldm = true
	s.ldmCh = make(chan bool, 1)
	s.listener.Close()
	s.listener = nil
	s.mu.Unlock()

	// Wait for accept loop to be done to make sure that no new
	// client can connect
	<-s.ldmCh

	s.mu.Lock()
	// Need to recheck few things
	if s.shutdown || len(s.clients) == 0 {
		s.mu.Unlock()
		// If there is no client, we need to call Shutdown() to complete
		// the LDMode. If server has been shutdown while lock was released,
		// calling Shutdown() should be no-op.
		s.Shutdown()
		return
	}
	dur := int64(s.getOpts().LameDuckDuration)
	dur -= atomic.LoadInt64(&lameDuckModeInitialDelay)
	if dur <= 0 {
		dur = int64(time.Second)
	}
	numClients := int64(len(s.clients))
	batch := 1
	// Sleep interval between each client connection close.
	si := dur / numClients
	if si < 1 {
		// Should not happen (except in test with very small LD duration), but
		// if there are too many clients, batch the number of close and
		// use a tiny sleep interval that will result in yield likely.
		si = 1
		batch = int(numClients / dur)
	} else if si > int64(time.Second) {
		// Conversely, there is no need to sleep too long between clients
		// and spread say 10 clients for the 2min duration. Sleeping no
		// more than 1sec.
		si = int64(time.Second)
	}

	// Now capture all clients
	clients := make([]*client, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}
	s.mu.Unlock()

	t := time.NewTimer(time.Duration(atomic.LoadInt64(&lameDuckModeInitialDelay)))
	// Delay start of closing of client connections in case
	// we have several servers that we want to signal to enter LD mode
	// and not have their client reconnect to each other.
	select {
	case <-t.C:
		s.Noticef("Closing existing clients")
	case <-s.quitCh:
		return
	}
	for i, client := range clients {
		client.closeConnection(ServerShutdown)
		if i == len(clients)-1 {
			break
		}
		if batch == 1 || i%batch == 0 {
			// We pick a random interval which will be at least si/2
			v := rand.Int63n(si)
			if v < si/2 {
				v = si / 2
			}
			t.Reset(time.Duration(v))
			// Sleep for given interval or bail out if kicked by Shutdown().
			select {
			case <-t.C:
			case <-s.quitCh:
				t.Stop()
				return
			}
		}
	}
	s.Shutdown()
}

// If given error is a net.Error and is temporary, sleeps for the given
// delay and double it, but cap it to ACCEPT_MAX_SLEEP. The sleep is
// interrupted if the server is shutdown.
// An error message is displayed depending on the type of error.
// Returns the new (or unchanged) delay.
func (s *Server) acceptError(acceptName string, err error, tmpDelay time.Duration) time.Duration {
	if ne, ok := err.(net.Error); ok && ne.Temporary() {
		s.Errorf("Temporary %s Accept Error(%v), sleeping %dms", acceptName, ne, tmpDelay/time.Millisecond)
		select {
		case <-time.After(tmpDelay):
		case <-s.quitCh:
			return tmpDelay
		}
		tmpDelay *= 2
		if tmpDelay > ACCEPT_MAX_SLEEP {
			tmpDelay = ACCEPT_MAX_SLEEP
		}
	} else if s.isRunning() {
		s.Errorf("%s Accept error: %v", acceptName, err)
	}
	return tmpDelay
}

func (s *Server) getRandomIP(resolver netResolver, url string) (string, error) {
	host, port, err := net.SplitHostPort(url)
	if err != nil {
		return "", err
	}
	// If already an IP, skip.
	if net.ParseIP(host) != nil {
		return url, nil
	}
	ips, err := resolver.LookupHost(context.Background(), host)
	if err != nil {
		return "", fmt.Errorf("lookup for host %q: %v", host, err)
	}
	var address string
	if len(ips) == 0 {
		s.Warnf("Unable to get IP for %s, will try with %s: %v", host, url, err)
		address = url
	} else {
		var ip string
		if len(ips) == 1 {
			ip = ips[0]
		} else {
			ip = ips[rand.Int31n(int32(len(ips)))]
		}
		// add the port
		address = net.JoinHostPort(ip, port)
	}
	return address, nil
}

// Returns true for the first attempt and depending on the nature
// of the attempt (first connect or a reconnect), when the number
// of attempts is equal to the configured report attempts.
func (s *Server) shouldReportConnectErr(firstConnect bool, attempts int) bool {
	opts := s.getOpts()
	if firstConnect {
		if attempts == 1 || attempts%opts.ConnectErrorReports == 0 {
			return true
		}
		return false
	}
	if attempts == 1 || attempts%opts.ReconnectErrorReports == 0 {
		return true
	}
	return false
}

// Invoked for route, leaf and gateway connections. Set the very first
// PING to a lower interval to capture the initial RTT.
// After that the PING interval will be set to the user defined value.
// Client lock should be held.
func (s *Server) setFirstPingTimer(c *client) {
	opts := s.getOpts()
	d := opts.PingInterval

	if !opts.DisableShortFirstPing {
		if c.kind != CLIENT {
			if d > firstPingInterval {
				d = firstPingInterval
			}
		} else if d > firstClientPingInterval {
			d = firstClientPingInterval
		}
	}
	// We randomize the first one by an offset up to 20%, e.g. 2m ~= max 24s.
	addDelay := rand.Int63n(int64(d / 5))
	d += time.Duration(addDelay)
	c.ping.tmr = time.AfterFunc(d, c.processPingTimer)
}
