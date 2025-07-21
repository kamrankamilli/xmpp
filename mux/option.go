// Copyright 2020 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package mux

import (
	"encoding/xml"

	"github.com/kamrankamilli/xmpp"
	"github.com/kamrankamilli/xmpp/disco/info"
	"github.com/kamrankamilli/xmpp/stanza"
)

// Option configures a ServeMux.
type Option func(m *ServeMux)

// IQ returns an option that matches IQ stanzas based on their type and the name
// of the payload.
func IQ(typ stanza.IQType, payload xml.Name, h IQHandler) Option {
	return func(m *ServeMux) {
		if h == nil {
			panic("mux: nil IQ handler")
		}
		pat := pattern{Stanza: iqStanza, Payload: payload, Type: string(typ)}
		if _, ok := m.iqPatterns[pat]; ok {
			panic("mux: multiple registrations for " + pat.String())
		}
		if m.iqPatterns == nil {
			m.iqPatterns = make(map[pattern]IQHandler)
		}
		m.iqPatterns[pat] = h
	}
}

// IQFunc returns an option that matches IQ stanzas.
// For more information see IQ.
func IQFunc(typ stanza.IQType, payload xml.Name, h IQHandlerFunc) Option {
	return IQ(typ, payload, h)
}

// Message returns an option that matches message stanzas by type.
func Message(typ stanza.MessageType, payload xml.Name, h MessageHandler) Option {
	return func(m *ServeMux) {
		if h == nil {
			panic("mux: nil message handler")
		}
		pat := pattern{Stanza: msgStanza, Payload: payload, Type: string(typ)}
		if _, ok := m.msgPatterns[pat]; ok {
			panic("mux: multiple registrations for " + pat.String())
		}
		if m.msgPatterns == nil {
			m.msgPatterns = make(map[pattern]MessageHandler)
		}
		m.msgPatterns[pat] = h
	}
}

// MessageFunc returns an option that matches message stanzas.
// For more information see Message.
func MessageFunc(typ stanza.MessageType, payload xml.Name, h MessageHandlerFunc) Option {
	return Message(typ, payload, h)
}

// Presence returns an option that matches presence stanzas by type.
func Presence(typ stanza.PresenceType, payload xml.Name, h PresenceHandler) Option {
	return func(m *ServeMux) {
		if h == nil {
			panic("mux: nil presence handler")
		}
		pat := pattern{Stanza: presStanza, Payload: payload, Type: string(typ)}
		if _, ok := m.presencePatterns[pat]; ok {
			panic("mux: multiple registrations for " + pat.String())
		}
		if m.presencePatterns == nil {
			m.presencePatterns = make(map[pattern]PresenceHandler)
		}
		m.presencePatterns[pat] = h
	}
}

// PresenceFunc returns an option that matches on presence stanzas.
// For more information see Presence.
func PresenceFunc(typ stanza.PresenceType, payload xml.Name, h PresenceHandlerFunc) Option {
	return Presence(typ, payload, h)
}

// Feature registers the provided features for service discovery.
//
// Most features will be implemented by Handlers and do not need to be
// registered again, Feature is just for features that should be advertised but
// do not have any corresponding handler.
func Feature(iter info.FeatureIter) Option {
	if iter == nil {
		panic("mux: nil info.FeatureIter")
	}
	return func(m *ServeMux) {
		m.features = append(m.features, iter)
	}
}

// Ident registers the provided identities for service discovery.
//
// Most identities will be implemented by Handlers and do not need to be
// registered again, Ident is just for features that should be advertised but do
// not have any corresponding handler.
func Ident(iter info.IdentityIter) Option {
	if iter == nil {
		panic("mux: nil info.IdentityIter")
	}
	return func(m *ServeMux) {
		m.idents = append(m.idents, iter)
	}
}

// Handle returns an option that matches on the provided XML name.
// If a handler already exists for n when the option is applied, the option
// panics.
func Handle(n xml.Name, h xmpp.Handler) Option {
	return func(m *ServeMux) {
		if h == nil {
			panic("mux: nil handler")
		}
		if stanza.Is(n, "") {
			panic("mux: tried to register stanza handler with Handle, use HandleIQ, HandleMessage, or HandlePresence instead")
		}
		if _, ok := m.patterns[n]; ok {
			panic("mux: multiple registrations for {" + n.Space + "}" + n.Local)
		}
		if m.patterns == nil {
			m.patterns = make(map[xml.Name]xmpp.Handler)
		}
		m.patterns[n] = h
	}
}

// HandleFunc returns an option that matches on the provided XML name.
func HandleFunc(n xml.Name, h xmpp.HandlerFunc) Option {
	return Handle(n, h)
}
