/*
 * Copyright (c) 2013 IBM Corp.
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v1.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v10.html
 *
 * Contributors:
 *    Seth Hoenig
 *    Allan Stockdill-Mander
 *    Mike Robertson
 */

package mqtt

import (
	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git/packets"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

const (
	msgExt = ".msg"
	bkpExt = ".bkp"
)

// FileStore implements the store interface using the filesystem to provide
// true persistence, even across client failure. This is designed to use a
// single directory per running client. If you are running multiple clients
// on the same filesystem, you will need to be careful to specify unique
// store directories for each.
type FileStore struct {
	sync.RWMutex
	directory string
	opened    bool
}

// NewFileStore will create a new FileStore which stores its messages in the
// directory provided.
func NewFileStore(directory string) *FileStore {
	store := &FileStore{
		directory: directory,
		opened:    false,
	}
	return store
}

// Open will allow the FileStore to be used.
func (store *FileStore) Open() {
	store.Lock()
	defer store.Unlock()
	// if no store directory was specified in ClientOpts, by default use the
	// current working directory
	if store.directory == "" {
		store.directory, _ = os.Getwd()
	}

	// if store dir exists, great, otherwise, create it
	if !exists(store.directory) {
		perms := os.FileMode(0770)
		merr := os.MkdirAll(store.directory, perms)
		chkerr(merr)
	}
	store.opened = true
	DEBUG.Println(STR, "store is opened at", store.directory)
}

// Close will disallow the FileStore from being used.
func (store *FileStore) Close() {
	store.Lock()
	defer store.Unlock()
	store.opened = false
	WARN.Println(STR, "store is not open")
}

// Put will put a message into the store, associated with the provided
// key value.
func (store *FileStore) Put(key string, m packets.ControlPacket) {
	store.Lock()
	defer store.Unlock()
	chkcond(store.opened)
	full := fullpath(store.directory, key)
	if exists(full) {
		backup(store.directory, key) // make a copy of what already exists
		defer unbackup(store.directory, key)
	}
	write(store.directory, key, m)
	chkcond(exists(full))
}

// Get will retrieve a message from the store, the one associated with
// the provided key value.
func (store *FileStore) Get(key string) packets.ControlPacket {
	store.RLock()
	defer store.RUnlock()
	chkcond(store.opened)
	filepath := fullpath(store.directory, key)
	if !exists(filepath) {
		return nil
	}
	mfile, oerr := os.Open(filepath)
	chkerr(oerr)
	//all, rerr := ioutil.ReadAll(mfile)
	//chkerr(rerr)
	msg, rerr := packets.ReadPacket(mfile)
	chkerr(rerr)
	cerr := mfile.Close()
	chkerr(cerr)
	return msg
}

// All will provide a list of all of the keys associated with messages
// currenly residing in the FileStore.
func (store *FileStore) All() []string {
	store.RLock()
	defer store.RUnlock()
	return store.all()
}

// Del will remove the persisted message associated with the provided
// key from the FileStore.
func (store *FileStore) Del(key string) {
	store.Lock()
	defer store.Unlock()
	store.del(key)
}

// Reset will remove all persisted messages from the FileStore.
func (store *FileStore) Reset() {
	store.Lock()
	defer store.Unlock()
	WARN.Println(STR, "FileStore Reset")
	for _, key := range store.all() {
		store.del(key)
	}
}

// lockless
func (store *FileStore) all() []string {
	chkcond(store.opened)
	keys := []string{}
	files, rderr := ioutil.ReadDir(store.directory)
	chkerr(rderr)
	for _, f := range files {
		DEBUG.Println(STR, "file in All():", f.Name())
		key := f.Name()[0 : len(f.Name())-4] // remove file extension
		keys = append(keys, key)
	}
	return keys
}

// lockless
func (store *FileStore) del(key string) {
	chkcond(store.opened)
	DEBUG.Println(STR, "store del filepath:", store.directory)
	DEBUG.Println(STR, "store delete key:", key)
	filepath := fullpath(store.directory, key)
	DEBUG.Println(STR, "path of deletion:", filepath)
	if !exists(filepath) {
		WARN.Println(STR, "store could not delete key:", key)
		return
	}
	rerr := os.Remove(filepath)
	chkerr(rerr)
	DEBUG.Println(STR, "del msg:", key)
	chkcond(!exists(filepath))
}

func fullpath(store string, key string) string {
	p := path.Join(store, key+msgExt)
	return p
}

func bkppath(store string, key string) string {
	p := path.Join(store, key+bkpExt)
	return p
}

// create file called "X.[messageid].msg" located in the store
// the contents of the file is the bytes of the message
// if a message with m's message id already exists, it will
// be overwritten
// X will be 'i' for inbound messages, and O for outbound messages
func write(store, key string, m packets.ControlPacket) {
	filepath := fullpath(store, key)
	f, err := os.Create(filepath)
	chkerr(err)
	werr := m.Write(f)
	chkerr(werr)
	cerr := f.Close()
	chkerr(cerr)
}

func exists(file string) bool {
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		chkerr(err)
	}
	return true
}

func backup(store, key string) {
	bkpp := bkppath(store, key)
	fulp := fullpath(store, key)
	backup, err := os.Create(bkpp)
	chkerr(err)
	mfile, oerr := os.Open(fulp)
	chkerr(oerr)
	_, cerr := io.Copy(backup, mfile)
	chkerr(cerr)
	clberr := backup.Close()
	chkerr(clberr)
	clmerr := mfile.Close()
	chkerr(clmerr)
}

// Identify .bkp files in the store and turn them into .msg files,
// whether or not it overwrites an existing file. This is safe because
// I'm copying the Paho Java client and they say it is.
func restore(store string) {
	files, rderr := ioutil.ReadDir(store)
	chkerr(rderr)
	for _, f := range files {
		fname := f.Name()
		if len(fname) > 4 {
			if fname[len(fname)-4:] == bkpExt {
				key := fname[0 : len(fname)-4]
				fulp := fullpath(store, key)
				msg, cerr := os.Create(fulp)
				chkerr(cerr)
				bkpp := path.Join(store, fname)
				bkp, oerr := os.Open(bkpp)
				chkerr(oerr)
				n, cerr := io.Copy(msg, bkp)
				chkerr(cerr)
				chkcond(n > 0)
				clmerr := msg.Close()
				chkerr(clmerr)
				clberr := bkp.Close()
				chkerr(clberr)
				remerr := os.Remove(bkpp)
				chkerr(remerr)
			}
		}
	}
}

func unbackup(store, key string) {
	bkpp := bkppath(store, key)
	remerr := os.Remove(bkpp)
	chkerr(remerr)
}
