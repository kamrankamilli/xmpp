// Copyright 2021 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package delay_test

import (
	"encoding/xml"
	"strconv"
	"strings"
	"testing"
	"time"

	"mellium.im/xmlstream"
	"github.com/kamrankamilli/xmpp/delay"
	"github.com/kamrankamilli/xmpp/jid"
	"github.com/kamrankamilli/xmpp/stanza"
)

var (
	_ xml.Marshaler       = delay.Delay{}
	_ xmlstream.Marshaler = delay.Delay{}
	_ xmlstream.WriterTo  = delay.Delay{}
	_ xml.Unmarshaler     = (*delay.Delay)(nil)
)

var insertTestCases = [...]struct {
	in     string
	out    string
	stanza bool
	ns     string
}{
	0: {},
	1: {
		stanza: true,
		in:     `<message xmlns="jabber:client"/>`,
		out:    `<message xmlns="jabber:client"><delay xmlns="urn:xmpp:delay" stamp="0001-01-01T00:00:00Z" from="me@example.net">foo</delay></message>`,
		ns:     stanza.NSClient,
	},
	2: {
		stanza: true,
		in:     `<message xmlns="jabber:server"/><message xmlns="jabber:server"><body>test</body></message><message xmlns="jabber:client"/>`,
		out:    `<message xmlns="jabber:server"><delay xmlns="urn:xmpp:delay" stamp="0001-01-01T00:00:00Z" from="me@example.net">foo</delay></message><message xmlns="jabber:server"><delay xmlns="urn:xmpp:delay" stamp="0001-01-01T00:00:00Z" from="me@example.net">foo</delay><body xmlns="jabber:server">test</body></message><message xmlns="jabber:client"></message>`,
		ns:     stanza.NSServer,
	},
	3: {
		stanza: true,
		in:     `<message xmlns="jabber:badns"/>`,
		out:    `<message xmlns="jabber:badns"></message>`,
		ns:     stanza.NSClient,
	},
	4: {
		in:  `<message xmlns="jabber:client"/>`,
		out: `<message xmlns="jabber:client"><delay xmlns="urn:xmpp:delay" stamp="0001-01-01T00:00:00Z" from="me@example.net">foo</delay></message>`,
		ns:  stanza.NSClient,
	},
	5: {
		in:  `<message xmlns="jabber:server"/><message xmlns="jabber:client"><body>test</body></message>`,
		out: `<message xmlns="jabber:server"><delay xmlns="urn:xmpp:delay" stamp="0001-01-01T00:00:00Z" from="me@example.net">foo</delay></message><message xmlns="jabber:client"><delay xmlns="urn:xmpp:delay" stamp="0001-01-01T00:00:00Z" from="me@example.net">foo</delay><body xmlns="jabber:client">test</body></message>`,
		ns:  stanza.NSServer,
	},
	6: {
		in:  `<message xmlns="jabber:badns"/>`,
		out: `<message xmlns="jabber:badns"><delay xmlns="urn:xmpp:delay" stamp="0001-01-01T00:00:00Z" from="me@example.net">foo</delay></message>`,
		ns:  stanza.NSClient,
	},
	7: {
		in:  `<message xmlns="jabber:anyns"/>`,
		out: `<message xmlns="jabber:anyns"><delay xmlns="urn:xmpp:delay" stamp="0001-01-01T00:00:00Z" from="me@example.net">foo</delay></message>`,
	},
}

func TestInsert(t *testing.T) {
	for i, tc := range insertTestCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			d := delay.Delay{From: jid.MustParse("me@example.net"), Time: time.Time{}, Reason: "foo"}
			var r xml.TokenReader
			if tc.stanza {
				stanzaDelayer := delay.Stanza(d, tc.ns)
				r = stanzaDelayer(xml.NewDecoder(strings.NewReader(tc.in)))
			} else {
				r = delay.Insert(d)(xml.NewDecoder(strings.NewReader(tc.in)))
			}
			// Prevent duplicate xmlns attributes. See https://mellium.im/issue/75
			r = xmlstream.RemoveAttr(func(start xml.StartElement, attr xml.Attr) bool {
				return (start.Name.Local == "message" || start.Name.Local == "iq") && attr.Name.Local == "xmlns"
			})(r)
			var buf strings.Builder
			e := xml.NewEncoder(&buf)
			_, err := xmlstream.Copy(e, r)
			if err != nil {
				t.Fatalf("error encoding: %v", err)
			}
			if err = e.Flush(); err != nil {
				t.Fatalf("error flushing: %v", err)
			}

			if out := buf.String(); tc.out != out {
				t.Errorf("wrong output:\nwant=%s,\n got=%s", tc.out, out)
			}
		})
	}
}

var marshalTests = [...]struct {
	unmarshal bool // true if we should only unmarshal for this test.
	in        delay.Delay
	out       string
}{
	0: {
		out: `<delay xmlns="urn:xmpp:delay" stamp="0001-01-01T00:00:00Z"></delay>`,
	},
	1: {
		in:  delay.Delay{From: jid.MustParse("me@example.net")},
		out: `<delay xmlns="urn:xmpp:delay" stamp="0001-01-01T00:00:00Z" from="me@example.net"></delay>`,
	},
	2: {
		in:  delay.Delay{Time: time.Time{}.Add(24 * time.Hour)},
		out: `<delay xmlns="urn:xmpp:delay" stamp="0001-01-02T00:00:00Z"></delay>`,
	},
	3: {
		in:  delay.Delay{Reason: "foo"},
		out: `<delay xmlns="urn:xmpp:delay" stamp="0001-01-01T00:00:00Z">foo</delay>`,
	},
	4: {
		in:  delay.Delay{From: jid.MustParse("me@example.net"), Time: time.Time{}.Add(24 * time.Hour), Reason: "foo"},
		out: `<delay xmlns="urn:xmpp:delay" stamp="0001-01-02T00:00:00Z" from="me@example.net">foo</delay>`,
	},
	5: {
		unmarshal: true,
		in:        delay.Delay{Time: time.Time{}.Add(24 * time.Hour), Reason: "foo"},
		out:       `<delay xmlns="urn:xmpp:delay" stamp="0001-01-02T00:00:00Z" foo:from="me@example.net" xmlns:foo="test">foo</delay>`,
	},
}

func TestMarshal(t *testing.T) {
	for i, tc := range marshalTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !tc.unmarshal {
				b, err := xml.Marshal(tc.in)
				if err != nil {
					t.Fatalf("unexpected error marshaling: %v", err)
				}
				if out := string(b); out != tc.out {
					t.Fatalf("wrong value:\nwant=%v,\n got=%v", tc.out, out)
				}
			}

			d := delay.Delay{}
			err := xml.Unmarshal([]byte(tc.out), &d)
			if err != nil {
				t.Fatalf("error unmarshaling: %v", err)
			}
			if !d.From.Equal(tc.in.From) {
				t.Errorf("wrong from JID: want=%v, got=%v", tc.in.From, d.From)
			}
			if !d.Time.Equal(tc.in.Time) {
				t.Errorf("wrong timestamp: want=%v, got=%v", tc.in.Time, d.Time)
			}
			if d.Reason != tc.in.Reason {
				t.Errorf("wrong chardata: want=%q, got=%q", tc.in.Reason, d.Reason)
			}
		})
	}
}
