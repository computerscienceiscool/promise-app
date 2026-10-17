package api

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/promisegrid/promise-app/backend/internal/protocol"
	"github.com/promisegrid/promise-app/backend/internal/store"
)

func TestUploadMailboxAndExactRetrieval(t *testing.T) {
	data, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	handler := New(data, "admin", slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))).Handler()
	fromPublic, fromPrivate, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	from, err := protocol.IdentityCID(fromPublic)
	if err != nil {
		t.Fatal(err)
	}
	toPublic, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	to, err := protocol.IdentityCID(toPublic)
	if err != nil {
		t.Fatal(err)
	}
	encoded, messageID, err := protocol.Build(protocol.FrozenProtocolCID(), fromPrivate, protocol.Payload{Kind: protocol.Offer, From: protocol.CIDTag(from), To: protocol.CIDTag(to), Body: "demo", CreatedAt: 1})
	if err != nil {
		t.Fatal(err)
	}

	upload := httptest.NewRequest(http.MethodPost, "/v1/envelopes", bytes.NewReader(encoded))
	upload.Header.Set("X-Device-ID", from.String())
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, upload)
	if response.Code != http.StatusCreated {
		t.Fatalf("upload: %d %s", response.Code, response.Body.String())
	}

	mailbox := httptest.NewRequest(http.MethodGet, "/v1/mailbox", nil)
	mailbox.Header.Set("X-Device-ID", to.String())
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, mailbox)
	if response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte(messageID.String())) {
		t.Fatalf("mailbox: %d %s", response.Code, response.Body.String())
	}

	fetch := httptest.NewRequest(http.MethodGet, "/v1/envelopes/"+messageID.String(), nil)
	fetch.Header.Set("X-Device-ID", to.String())
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, fetch)
	if response.Code != http.StatusOK || !bytes.Equal(response.Body.Bytes(), encoded) {
		t.Fatalf("fetch did not return exact bytes")
	}

	audit := httptest.NewRequest(http.MethodGet, "/v1/admin/audit", nil)
	audit.Header.Set("Authorization", "Bearer admin")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, audit)
	var decoded map[string]any
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &decoded) != nil {
		t.Fatalf("audit: %d %s", response.Code, response.Body.String())
	}
}

func TestUploadRejectsUnknownProtocolSelector(t *testing.T) {
	data, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	handler := New(data, "admin", slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))).Handler()
	fromPublic, fromPrivate, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	from, err := protocol.IdentityCID(fromPublic)
	if err != nil {
		t.Fatal(err)
	}
	toPublic, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	to, err := protocol.IdentityCID(toPublic)
	if err != nil {
		t.Fatal(err)
	}
	// This valid, signed envelope differs only in selector. It remains
	// verifiable evidence but this relay must not interpret or route it.
	other, err := cid.Parse("bafkreigh2akiscaildc5hiywzbllsgrcfqew2gozbkkfu4a2yv56zzrxmu")
	if err != nil {
		t.Fatal(err)
	}
	encoded, _, err := protocol.Build(other, fromPrivate, protocol.Payload{Kind: protocol.Offer, From: protocol.CIDTag(from), To: protocol.CIDTag(to), Body: "other", CreatedAt: 1})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/envelopes", bytes.NewReader(encoded))
	request.Header.Set("X-Device-ID", from.String())
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || !bytes.Contains(response.Body.Bytes(), []byte("unsupported protocol selector")) {
		t.Fatalf("upload: %d %s", response.Code, response.Body.String())
	}
}
