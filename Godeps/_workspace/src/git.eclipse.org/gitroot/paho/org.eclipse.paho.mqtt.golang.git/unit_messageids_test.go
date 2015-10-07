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
	"fmt"
	"testing"
	"time"
)

type DummyToken struct{}

func (d *DummyToken) Wait() bool {
	return true
}

func (d *DummyToken) WaitTimeout(t time.Duration) bool {
	return true
}

func (d *DummyToken) flowComplete() {}

func (d *DummyToken) Error() error {
	return nil
}

func Test_getID(t *testing.T) {
	mids := &messageIds{index: make(map[uint16]Token)}

	i1 := mids.getID(&DummyToken{})

	if i1 != 1 {
		t.Fatalf("i1 was wrong: %v", i1)
	}

	i2 := mids.getID(&DummyToken{})

	if i2 != 2 {
		t.Fatalf("i2 was wrong: %v", i2)
	}

	for i := uint16(3); i < 100; i++ {
		id := mids.getID(&DummyToken{})
		if id != i {
			t.Fatalf("id was wrong expected %v got %v", i, id)
		}
	}
}

func Test_freeID(t *testing.T) {
	mids := &messageIds{index: make(map[uint16]Token)}

	i1 := mids.getID(&DummyToken{})
	mids.freeID(i1)

	if i1 != 1 {
		t.Fatalf("i1 was wrong: %v", i1)
	}

	i2 := mids.getID(&DummyToken{})
	fmt.Printf("i2: %v\n", i2)
}

func Test_messageids_mix(t *testing.T) {
	mids := &messageIds{index: make(map[uint16]Token)}

	done := make(chan bool)
	a := make(chan uint16, 3)
	b := make(chan uint16, 20)
	c := make(chan uint16, 100)

	go func() {
		for i := 0; i < 10000; i++ {
			a <- mids.getID(&DummyToken{})
			mids.freeID(<-b)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10000; i++ {
			b <- mids.getID(&DummyToken{})
			mids.freeID(<-c)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10000; i++ {
			c <- mids.getID(&DummyToken{})
			mids.freeID(<-a)
		}
		done <- true
	}()

	<-done
	<-done
	<-done
}
