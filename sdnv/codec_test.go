// Copyright © 2017-2019 Nelz
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
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"strings"
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
	buf := make([]byte, MaxByteSize)
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

		// Uint64 Writer version
		bb := bytes.NewBufferString("")
		wSize, err := Write(bb, test.num)
		if err != nil {
			t.Errorf("unexpected: %v\n", err)
		}
		if size != len(test.data) {
			t.Errorf("expected %d: %d\n", len(test.data), wSize)
		}
		if !bytes.Equal(test.data, bb.Bytes()) {
			t.Errorf("expected %b: %b\n", test.data, bb.Bytes())
		}
	}
}

func TestDecodes(t *testing.T) {
	buf := make([]byte, MaxByteSize)
	for _, test := range tests {
		size := Encode(buf, test.num)

		// Uint64 version
		r1, n1 := Decode(buf[:size])
		if size != n1 {
			t.Errorf("expected %d: %d\n", size, n1)
		}
		if test.num != r1 {
			t.Errorf("expected %d: %d\n", test.num, r1)
		}

		// big.Int version
		x, n2 := decodeBig(buf[:size])
		if size != n2 {
			t.Errorf("expected %d: %d\n", size, n2)
		}
		if test.num != x.Uint64() {
			t.Errorf("expected %d: %d\n", test.num, x.Uint64())
		}

		// Uint64 Reader version
		bb := bytes.NewBuffer(buf)
		r3 := uint64(0)
		n3, err := Read(bb, &r3)
		if err != nil {
			t.Errorf("unexpected: %v\n", err)
		}
		if size != n3 {
			t.Errorf("expected %d: %d\n", size, n3)
		}
		if test.num != r3 {
			t.Errorf("expected %d: %d\n", test.num, r3)
		}

		// Uint64 ByteReader version
		br := bytes.NewBuffer(buf)
		r4 := uint64(0)
		n4, err := ReadBytes(br, &r4)
		if err != nil {
			t.Errorf("unexpected: %v\n", err)
		}
		if size != n4 {
			t.Errorf("expected %d: %d\n", size, n4)
		}
		if test.num != r4 {
			t.Errorf("expected %d: %d\n", test.num, r4)
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

func TestDecodeErrors(t *testing.T) {
	var testCases = []struct {
		name string
		data []byte
		size int
		err  string
	}{
		{
			"zero length",
			[]byte{},
			0,
			io.EOF.Error(),
		},
		{
			"non-zero underflow",
			[]byte{0xff, 0xff},
			2,
			io.ErrUnexpectedEOF.Error(),
		},
		{
			"greater than 64 bit integer",
			[]byte{
				0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff,
			},
			10,
			ErrOverflow64,
		},
		{
			"one too many most significant bits",
			[]byte{
				0x83,
				0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff,
				0x7f,
			},
			10,
			ErrOverflow64,
		},
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("%s io.Reader", tc.name),
			func(t *testing.T) {
				bb := bytes.NewBuffer(tc.data)
				val := uint64(0)
				n, err := Read(bb, &val)
				if tc.size != n {
					t.Errorf("expected %d: %d\n", tc.size, n)
				}
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Errorf("expected [%s]: %v\n", tc.err, err)
				}
			},
		)

		t.Run(
			fmt.Sprintf("%s io.ByteReader", tc.name),
			func(t *testing.T) {
				bb := bytes.NewBuffer(tc.data)
				val := uint64(0)
				n, err := ReadBytes(bb, &val)
				if tc.size != n {
					t.Errorf("expected %d: %d\n", tc.size, n)
				}
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Errorf("expected [%s]: %v\n", tc.err, err)
				}
			},
		)
	}
}

// TODO: TestPanics
