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
	"testing"
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

func TestPut(t *testing.T) {
	buf := make([]byte, 10)
	for _, test := range tests {
		size := Put(buf, test.num)
		if size != len(test.data) {
			t.Errorf("expected %d: %d\n", len(test.data), size)
		}
		if !bytes.Equal(test.data, buf[:size]) {
			t.Errorf("expected %b: %b\n", test.data, buf[:size])
		}
	}
}

func TestGet(t *testing.T) {
	buf := make([]byte, 10)
	for _, test := range tests {
		size := Put(buf, test.num)
		r, n := Get(buf[:size])

		if size != n {
			t.Errorf("expected %d: %d\n", size, n)
		}
		if test.num != r {
			t.Errorf("expected %d: %d\n", test.num, r)
		}
	}
}
