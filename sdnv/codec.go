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

// Package sdnv packages implements the Self-Delimiting Numeric Values,
// as per https://tools.ietf.org/html/rfc5050#section-4.1
package sdnv

import (
	"math/big"
	"math/bits"
)

// Encode puts the given uint64 into the buffer, and return the number of
// bytes used in the buffer.
// Put panics if there is not enough space in the buffer.
// Design can be found at: https://tools.ietf.org/html/rfc5050#section-4.1
func Encode(buf []byte, x uint64) (n int) {
	if x == 0 {
		buf[n] = 0x00
		return n + 1
	}

	n = (bits.Len64(x) - 1) / 7
	for i := n; i >= 0; i-- {
		buf[i] = byte(x) & 0x7f
		if i != n {
			buf[i] |= 0x80
		}
		x >>= 7
	}
	return n + 1
}

func encodeBig(buf []byte, in *big.Int) (n int) {
	bLen := in.BitLen()
	if bLen == 0 {
		buf[n] = 0x00
		return n + 1
	}

	x := big.NewInt(0).SetBytes(in.Bytes())
	n = (bLen - 1) / 7
	raw := x.Bytes()
	for i := n; i >= 0; i-- {
		buf[i] = raw[len(raw)-1] & 0x7f
		if i != n {
			buf[i] |= 0x80
		}
		raw = x.Rsh(x, 7).Bytes()
	}
	return n + 1
}

// Decode retrieves a uint64 value from the buffer, returning the uint64 and
// the number of bytes consumed from the buffer.
// Get panics if it runs out of bytes in the buffer before encountering
// the delimiter byte.
// Design can be found at: https://tools.ietf.org/html/rfc5050#section-4.1
func Decode(buf []byte) (x uint64, n int) {
	for {
		x |= uint64(buf[n] & 0x7f)
		if buf[n] < 0x80 {
			return x, n + 1
		}
		x <<= 7
		n++
	}
}

func decodeBig(buf []byte) (x *big.Int, n int) {
	x = big.NewInt(0)
	for {
		bVal := int64(buf[n] & 0x7f)
		x.Or(x, big.NewInt(bVal))
		if buf[n] < 0x80 {
			return x, n + 1
		}
		x.Lsh(x, 7)
		n++
	}
}
