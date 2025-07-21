// Copyright 2020 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package stream_test

import (
	"encoding/xml"
	"errors"
	"io"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"mellium.im/xmlstream"
	"github.com/kamrankamilli/xmpp/internal/stream"
	"github.com/kamrankamilli/xmpp/internal/xmpptest"
	streamerr "github.com/kamrankamilli/xmpp/stream"
)

var readerTestCases = [...]struct {
	in      string
	skip    int
	err     error
	errType error
}{
	0: {},
	1: {
		in: `<stream></stream>`,
	},
	2: {
		in: `<stream:stream
					version='1.0'
					xmlns='jabber:client'
					xmlns:stream='http://wrong.namespace.example.org/'/>`,
	},
	3: {
		in: `<other:stream
					version='1.0'
					xmlns='jabber:client'
					xmlns:other='http://etherx.jabber.org/streams'/>`,
		err: stream.ErrUnexpectedRestart,
	},
	4: {
		in: `<stream:stream
					version='1.0'
					xmlns='jabber:client'
					xmlns:stream='http://etherx.jabber.org/streams'/>`,
		err: stream.ErrUnexpectedRestart,
	},
	5: {
		in: `<stream:unknown
					version='1.0'
					xmlns='jabber:client'
					xmlns:stream='http://etherx.jabber.org/streams'/>`,
		err: stream.ErrUnknownStreamElement,
	},
	6: {
		in: `<stream:error/>`,
	},
	7: {
		in:  `<stream:error xmlns:stream='http://etherx.jabber.org/streams'/>`,
		err: streamerr.Error{},
	},
	8: {
		in:  `<error xmlns='http://etherx.jabber.org/streams'/>`,
		err: streamerr.Error{},
	},
	9: {
		in: `<stream:error xmlns:stream='http://etherx.jabber.org/streams'>
	<not-well-formed xmlns='urn:ietf:params:xml:ns:xmpp-streams'/>
</stream:error>`,
		err: streamerr.NotWellFormed,
	},
	10: {
		in: `<stream:error xmlns:stream='http://etherx.jabber.org/streams'>
	</notwellformed>
</stream:error>`,
		errType: &xml.SyntaxError{},
	},
	11: {
		in:   `<stream:stream xmlns:stream='http://etherx.jabber.org/streams'></stream:stream>`,
		skip: 1,
	},
}

func TestReader(t *testing.T) {
	for i, tc := range readerTestCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var out strings.Builder
			d := xml.NewDecoder(strings.NewReader(tc.in))
			e := xml.NewEncoder(&out)
			for tc.skip > 0 {
				// Throw away any tokens that we need to make the stream well-formed,
				// but that we don't want in the actual test output.
				tok, err := d.Token()
				if err != nil {
					t.Fatalf("error skipping tokens: %v", err)
				}
				err = e.EncodeToken(tok)
				if err != nil {
					t.Fatalf("error encoding skipped tokens: %v", err)
				}
				tc.skip--
			}
			_, err := xmlstream.Copy(e, stream.Reader(d, false))
			switch {
			case tc.errType != nil:
				if reflect.TypeOf(tc.errType) != reflect.TypeOf(err) {
					t.Errorf("unexpected error type: want=%T, got=%T", tc.err, err)
				}
			case !errors.Is(err, tc.err):
				t.Errorf("unexpected error: want=%v, got=%#v", tc.err, err)
			}
			if err = e.Flush(); err != nil {
				t.Fatalf("error flushing output to buffer: %v", err)
			}
			t.Logf("output: %q", out.String())
		})
	}
}

func TestBadFormat(t *testing.T) {
	toks := &xmpptest.Tokens{
		xml.EndElement{Name: xml.Name{Local: "error", Space: streamerr.NS}},
	}
	r := stream.Reader(toks, false)
	var out strings.Builder
	e := xml.NewEncoder(&out)
	_, err := xmlstream.Copy(e, r)
	if !errors.Is(err, streamerr.BadFormat) {
		t.Errorf("unexpected error: want=%v, got=%v", streamerr.BadFormat, err)
	}
	err = e.Flush()
	if err != nil {
		t.Errorf("error flushing: %v", err)
	}
	t.Logf("output: %q", out.String())
}

func TestDisallowedTokenType(t *testing.T) {
	toks := &xmpptest.Tokens{
		xml.Comment("foo"),
		xml.ProcInst{
			Target: "test",
			Inst:   []byte("test"),
		},
		xml.Directive("test"),
	}
	r := stream.Reader(toks, false)
	for {
		tok, err := r.Token()
		if err == io.EOF {
			if tok != nil {
				t.Fatalf("got token %T(%[1]v) and io.EOF", t)
			}
			break
		}
		if err == nil {
			t.Errorf("expected error when hitting disallowed %T token, got none", tok)
		}
		if tok != nil {
			t.Errorf("expected no token to be returned, got %T(%[1]v)", tok)
		}
	}
}
