// Copyright 2017 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package stanza_test

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"testing"

	"mellium.im/xmlstream"
	"github.com/kamrankamilli/xmpp/jid"
	"github.com/kamrankamilli/xmpp/stanza"
)

type testReader []xml.Token

func (r *testReader) Token() (t xml.Token, err error) {
	tr := *r
	if len(tr) < 1 {
		return nil, io.EOF
	}
	t, *r = tr[0], tr[1:]
	return t, nil
}

var start = xml.StartElement{
	Name: xml.Name{Local: "ping"},
}

type messageTest struct {
	to      string
	typ     stanza.MessageType
	payload xml.TokenReader
	out     string
	err     error
}

var messageTests = [...]messageTest{
	0: {
		to:      "new@example.net",
		payload: &testReader{},
	},
	1: {
		to:      "new@example.org",
		payload: &testReader{start, start.End()},
		out:     `<ping></ping>`,
		typ:     stanza.NormalMessage,
	},
}

func TestMessage(t *testing.T) {
	for i, tc := range messageTests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			b := new(bytes.Buffer)
			e := xml.NewEncoder(b)
			message := stanza.Message{To: jid.MustParse(tc.to), Type: tc.typ}.Wrap(tc.payload)
			if _, err := xmlstream.Copy(e, message); err != tc.err {
				t.Errorf("Unexpected error: want=`%v', got=`%v'", tc.err, err)
			}
			if err := e.Flush(); err != nil {
				t.Fatalf("Error flushing: %q", err)
			}

			o := b.String()
			jidattr := fmt.Sprintf(`to="%s"`, tc.to)
			if !strings.Contains(o, jidattr) {
				t.Errorf("Expected output to have attr `%s',\ngot=`%s'", jidattr, o)
			}
			typeattr := fmt.Sprintf(`type="%s"`, string(tc.typ))
			if !strings.Contains(o, typeattr) {
				t.Errorf("Expected output to have attr `%s',\ngot=`%s'", typeattr, o)
			}
			if !strings.Contains(o, tc.out) {
				t.Errorf("Expected output to contain payload `%s',\ngot=`%s'", tc.out, o)
			}
		})
	}
}
