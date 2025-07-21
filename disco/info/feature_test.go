// Copyright 2021 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package info_test

import (
	"encoding/xml"
	"testing"

	"mellium.im/xmlstream"
	"github.com/kamrankamilli/xmpp/disco"
	"github.com/kamrankamilli/xmpp/disco/info"
	"github.com/kamrankamilli/xmpp/internal/xmpptest"
)

var (
	_ xml.Marshaler       = info.Feature{}
	_ xmlstream.Marshaler = info.Feature{}
	_ xmlstream.WriterTo  = info.Feature{}
)

func TestEncodeFeature(t *testing.T) {
	xmpptest.RunEncodingTests(t, []xmpptest.EncodingTestCase{
		0: {
			Value:       &info.Feature{},
			XML:         `<feature xmlns="http://jabber.org/protocol/disco#info" var=""></feature>`,
			NoUnmarshal: true,
		},
		1: {
			Value: &info.Feature{
				XMLName: xml.Name{Space: disco.NSInfo, Local: "feature"},
				Var:     "urn:example",
			},
			XML: `<feature xmlns="http://jabber.org/protocol/disco#info" var="urn:example"></feature>`,
		},
	})
}
