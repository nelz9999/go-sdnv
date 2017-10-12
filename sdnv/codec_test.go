// Copyright © 2017 Nelz
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package sdnv

import (
	"bytes"
	"math/big"
	"math/rand"
	"testing"
	"time"
)

var tests = []struct {
	num  uint64
	data []byte
}{
	{ // via RFC 5050 4.1
		uint64(0xabc),
		[]byte{0x95, 0x3c},
	},
	{ // via RFC 5050 4.1
		uint64(0x1234),
		[]byte{0xa4, 0x34},
	},
	{ // via RFC 5050 4.1
		uint64(0x4234),
		[]byte{0x81, 0x84, 0x34},
	},
	{ // via RFC 5050 4.1
		uint64(0x7f),
		[]byte{0x7f},
	},
	{ // Lower bound
		uint64(0x00),
		[]byte{0x00},
	},
	{ // Upper bound
		uint64(0xffffffffffffffff),
		[]byte{
			0x81,
			0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff,
			0x7f,
		},
	},
}

func TestEncodes(t *testing.T) {
	buf := make([]byte, 10)
	for _, test := range tests {
		// Uint64 version
		size := Encode(buf, test.num)
		if size != len(test.data) {
			t.Errorf("expected %d: %d\n", len(test.data), size)
		}
		if !bytes.Equal(test.data, buf[:size]) {
			t.Errorf("expected %b: %b\n", test.data, buf[:size])
		}

		// big.Int version
		x := big.NewInt(0).SetUint64(test.num)
		bSize := encodeBig(buf, x)
		if bSize != len(test.data) {
			t.Errorf("expected %d: %d\n", len(test.data), bSize)
		}
		if !bytes.Equal(test.data, buf[:bSize]) {
			t.Errorf("expected %b: %b\n", test.data, buf[:bSize])
		}
		if x.Uint64() != test.num {
			t.Errorf("modified %d: %d\n", test.num, x.Uint64())
		}
	}
}

func TestDecodes(t *testing.T) {
	buf := make([]byte, 10)
	for _, test := range tests {
		size := Encode(buf, test.num)

		// Uint64 version
		r, n := Decode(buf[:size])
		if size != n {
			t.Errorf("expected %d: %d\n", size, n)
		}
		if test.num != r {
			t.Errorf("expected %d: %d\n", test.num, r)
		}

		// big.Int version
		x, l := decodeBig(buf[:size])
		if size != l {
			t.Errorf("expected %d: %d\n", size, l)
		}
		if test.num != x.Uint64() {
			t.Errorf("expected %d: %d\n", test.num, x.Uint64())
		}
	}
}

func TestBigInts(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	xLen := 16 + r.Intn(8)
	one := big.NewInt(1)

	// Build a value that has the least significant bit
	// set to 1 for each hept (grouping of 7)
	xVal := big.NewInt(0)
	xBuf := make([]byte, xLen)
	for i := 0; i < xLen; i++ {
		xVal.Lsh(xVal, 7).Add(xVal, one)
		xBuf[i] = 0x81
	}
	xBuf[len(xBuf)-1] = 0x01

	// Encode it
	buf := make([]byte, xLen)
	size := encodeBig(buf, xVal)
	if size != xLen {
		t.Errorf("expected %d: %d\n", xLen, size)
	}

	if !bytes.Equal(xBuf, buf) {
		t.Errorf("expected [% #x]: [% #x]\n", xBuf, buf)
	}

	// And then Decode it
	val, n := decodeBig(buf)
	if n != xLen {
		t.Errorf("expected %d: %d\n", xLen, n)
	}
	if xVal.Cmp(val) != 0 {
		t.Errorf("expected %s: %s\n", xVal, val)
	}
}
