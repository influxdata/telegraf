package mbserver

import "sync"

// Derived from https://github.com/lammertb/libcrc
/*
 * Library: libcrc
 * File:    src/crc16.c
 * Author:  Lammert Bies
 *
 * This file is licensed under the MIT License as stated below
 *
 * Copyright (c) 1999-2016 Lammert Bies
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 * Description
 * -----------
 * The source file src/crc16.c contains routines which calculate the common
 * CRC16 cyclic redundancy check values for an incomming byte string.
 */

var crcTable []uint16
var mux sync.Mutex

func crcModbus(data []byte) (crc uint16) {
	if crcTable == nil {
		// Thread safe initialization.
		mux.Lock()
		if crcTable == nil {
			crcInitTable()
		}
		mux.Unlock()
	}

	crc = 0xffff
	for _, v := range data {
		crc = (crc >> 8) ^ crcTable[(crc^uint16(v))&0x00FF]
	}

	return crc
}

func crcInitTable() {
	crc16IBM := uint16(0xA001)
	crcTable = make([]uint16, 256)

	for i := uint16(0); i < 256; i++ {

		crc := uint16(0)
		c := uint16(i)

		for j := uint16(0); j < 8; j++ {
			if ((crc ^ c) & 0x0001) > 0 {
				crc = (crc >> 1) ^ crc16IBM
			} else {
				crc = crc >> 1
			}
			c = c >> 1
		}
		crcTable[i] = crc
	}
}
