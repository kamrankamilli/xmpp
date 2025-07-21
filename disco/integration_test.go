// Copyright 2021 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

//go:build integration
// +build integration

package disco_test

import (
	"context"
	"crypto/tls"
	"testing"

	"mellium.im/sasl"
	"github.com/kamrankamilli/xmpp"
	"github.com/kamrankamilli/xmpp/disco"
	"github.com/kamrankamilli/xmpp/internal/integration"
	"github.com/kamrankamilli/xmpp/internal/integration/ejabberd"
	"github.com/kamrankamilli/xmpp/internal/integration/jackal"
	"github.com/kamrankamilli/xmpp/internal/integration/prosody"
)

func TestIntegrationInfo(t *testing.T) {
	prosodyRun := prosody.Test(context.TODO(), t,
		integration.Log(),
		prosody.ListenC2S(),
	)
	prosodyRun(integrationRequestInfo)

	ejabberdRun := ejabberd.Test(context.TODO(), t,
		integration.Log(),
		ejabberd.ListenC2S(),
	)
	ejabberdRun(integrationRequestInfo)

	jackalRun := jackal.Test(context.TODO(), t,
		integration.Log(),
		jackal.ListenC2S(),
	)
	jackalRun(integrationRequestInfo)
}

func integrationRequestInfo(ctx context.Context, t *testing.T, cmd *integration.Cmd) {
	j, pass := cmd.User()
	session, err := cmd.DialClient(ctx, j, t,
		xmpp.StartTLS(&tls.Config{
			InsecureSkipVerify: true,
		}),
		xmpp.SASL("", pass, sasl.Plain, sasl.ScramSha256),
		xmpp.BindResource(),
	)
	if err != nil {
		t.Fatalf("error connecting: %v", err)
	}
	go func() {
		err := session.Serve(nil)
		if err != nil {
			t.Logf("error from serve: %v", err)
		}
	}()

	info, err := disco.GetInfo(ctx, "", j.Domain(), session)
	if err != nil {
		t.Errorf("error getting info: %v", err)
	}
	if len(info.Features) == 0 {
		t.Errorf("expected to get features back")
	}
	if len(info.Identity) == 0 {
		t.Errorf("expected to get identities back")
	}
}
