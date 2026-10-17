package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/promisegrid/promise-app/backend/internal/protocol"
)

func TestIdentityIsPersistent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "identity")
	first, created, err := loadOrCreateIdentity(path)
	if err != nil || !created {
		t.Fatalf("first identity: created=%v err=%v", created, err)
	}
	second, created, err := loadOrCreateIdentity(path)
	if err != nil || created {
		t.Fatalf("second identity: created=%v err=%v", created, err)
	}
	a, _ := first.cid()
	b, _ := second.cid()
	if a != b {
		t.Fatalf("identity changed: %s != %s", a, b)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("private identity mode = %o", info.Mode().Perm())
	}
}

func TestFetchInboxVerifiesAndRetainsExactBytes(t *testing.T) {
	recipientKey := mustKey(t)
	recipient := identity{private: recipientKey}
	recipientCID, err := recipient.cid()
	if err != nil {
		t.Fatal(err)
	}
	sender := mustKey(t)
	senderCID, err := protocol.IdentityCID(sender.Public().(ed25519.PublicKey))
	if err != nil {
		t.Fatal(err)
	}
	to, _ := protocol.IdentityCID(recipientKey.Public().(ed25519.PublicKey))
	bytes, messageCID, err := protocol.Build(protocol.FrozenProtocolCID(), sender, protocol.Payload{Kind: protocol.Offer, From: protocol.CIDTag(senderCID), To: protocol.CIDTag(to), Body: "I promise to demo this.", CreatedAt: 1})
	if err != nil {
		t.Fatal(err)
	}
	mailboxBody, err := json.Marshal(map[string]any{"items": []map[string]string{{"cid": messageCID.String()}}})
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("X-Device-ID") != recipientCID {
			t.Errorf("wrong device header: %q", r.Header.Get("X-Device-ID"))
		}
		switch r.URL.Path {
		case "/v1/mailbox":
			return testResponse(http.StatusOK, mailboxBody), nil
		case "/v1/envelopes/" + messageCID.String():
			return testResponse(http.StatusOK, bytes), nil
		default:
			return testResponse(http.StatusNotFound, []byte("not found")), nil
		}
	})}
	dir := t.TempDir()
	got, err := fetchInbox(client, "http://relay.test", recipient, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].CID != messageCID.String() || got[0].Body != "I promise to demo this." {
		t.Fatalf("display = %#v", got)
	}
	stored, err := os.ReadFile(filepath.Join(dir, messageCID.String()+".cbor"))
	if err != nil {
		t.Fatal(err)
	}
	if string(stored) != string(bytes) {
		t.Fatal("stored bytes changed")
	}
}

func TestWriteEvidenceIsPrivateAndContainsVerifiedMessage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "evidence", "fetch.json")
	report := fetchEvidence{
		FetchedAt: "2026-10-17T00:00:00Z",
		Relay:     "http://192.168.1.10:8080",
		Recipient: "bafkrecipient",
		Messages:  []displayMessage{{CID: "bafkmessage", Stored: ".promise-peer/inbox/bafkmessage.cbor"}},
	}
	if err := writeEvidence(path, report); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got fetchEvidence
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.Recipient != report.Recipient || len(got.Messages) != 1 || got.Messages[0].CID != report.Messages[0].CID {
		t.Fatalf("evidence = %#v", got)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("evidence mode = %o", info.Mode().Perm())
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func testResponse(status int, body []byte) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}

func mustKey(t *testing.T) ed25519.PrivateKey {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return key
}
