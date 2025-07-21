// Copyright 2017 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

//go:generate go run -tags=tools golang.org/x/tools/cmd/stringer -type=CVD

// Package color implements XEP-0392: Consistent Color Generation v0.4.
package color // import "kamrankamilli/xmpp/color"

import (
	/* #nosec */
	"crypto/sha1"
	"encoding/binary"
	"hash"
	"image/color"
	"math"
)

// Size is the length of the hash output.
const Size = 2

// CVD represents a color vision deficiency to correct for.
type CVD uint8

// A list of color vision deficiencies.
const (
	None CVD = iota
	RedGreen
	Blue
)

// Hash returns a new hash.Hash computing the Y'CbCr color.
// For more information see Sum.
func Hash(cvd CVD) hash.Hash {
	return digest{
		/* #nosec */
		Hash: sha1.New(),
		cvd:  cvd,
	}
}

type digest struct {
	hash.Hash
	cvd CVD
}

func (d digest) Size() int { return Size }
func (d digest) Sum(b []byte) []byte {
	b = d.Hash.Sum(b)
	i := binary.LittleEndian.Uint16(b[:2])
	switch d.cvd {
	case None:
	case RedGreen:
		i &= 0x7fff
	case Blue:
		i = (i & 0x7fff) | (((i & 0x4000) << 1) ^ 0x8000)
	default:
		panic("color: invalid color vision deficiency")
	}
	angle := float64(i) / 65536 * 2 * math.Pi
	cr, cb := math.Sincos(angle)
	factor := 0.5 / math.Max(math.Abs(cr), math.Abs(cb))
	cb, cr = cb*factor, cr*factor

	b[0] = uint8(math.Min(math.Max(cb+0.5, 0)*255, 255))
	b[1] = uint8(math.Min(math.Max(cr+0.5, 0)*255, 255))
	return b[:Size]
}

// Sum returns a color in the Y'CbCr colorspace in the form [Cb, Cr] that is
// consistent for the same inputs.
//
// If a color vision deficiency constant is provided (other than None), the
// algorithm attempts to avoid confusable colors.
func Sum(data []byte, cvd CVD) [Size]byte {
	b := make([]byte, 0, Size)
	h := Hash(cvd)
	/* #nosec */
	h.Write(data)
	b = h.Sum(b)
	return [Size]byte{b[0], b[1]}
}

// Bytes converts a byte slice to a color.YCbCr.
//
// For more information see Sum.
func Bytes(b []byte, luma uint8, cvd CVD) color.YCbCr {
	ba := Sum(b, cvd)
	return color.YCbCr{
		Y:  luma,
		Cb: ba[0],
		Cr: ba[1],
	}
}

// String converts a string to a color.YCbCr.
//
// For more information see Sum.
func String(s string, luma uint8, cvd CVD) color.YCbCr {
	return Bytes([]byte(s), luma, cvd)
}
