package vsphere

import (
	"log"
	"net/url"
	"sync"
)

type poolMember struct {
	Client *Client
	next   *poolMember
}

// Pool is a simple free-list based pool of vSphere clients
type Pool struct {
	u    *url.URL
	v    *VSphere
	root *poolMember
	mux  sync.Mutex
}

// Take returns a client, either by picking an available one from the pool or creating a new one
func (p *Pool) Take() (*Client, error) {
	p.mux.Lock()
	defer p.mux.Unlock()
	if p.root != nil {
		r := p.root
		p.root = r.next
		log.Printf("D! //////// Getting connection from pool")
		return r.Client, nil
	}
	// Pool is empty, create a new client!
	//
	log.Printf("D! ******* Pool is empty, creating new client")
	return NewClient(p.u, p.v)
}

// Return put a client back to the free list
func (p *Pool) Return(client *Client) {
	if client == nil {
		log.Printf("E! Connection taken out of pool due to error")
		return // Useful when you want to override a deferred Return
	}
	p.mux.Lock()
	defer p.mux.Unlock()
	r := &poolMember{client, p.root}
	p.root = r
}
