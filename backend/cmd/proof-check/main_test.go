package main

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type vectors struct {
	Identities struct {
		OfferTo string `json:"offer_to"`
	} `json:"identities"`
	Messages []struct {
		CID         string `json:"cid"`
		EnvelopeHex string `json:"envelope_hex"`
	} `json:"messages"`
}

func TestVerifyChecksRetainedEvidence(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "promise-messaging-v0-vectors.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture vectors
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	envelope, err := hex.DecodeString(fixture.Messages[0].EnvelopeHex)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	for _, dir := range []string{"objects", "mailboxes", "audit"} {
		if err := os.Mkdir(filepath.Join(root, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	cid := fixture.Messages[0].CID
	if err := os.WriteFile(filepath.Join(root, "objects", cid+".cbor"), envelope, 0600); err != nil {
		t.Fatal(err)
	}
	mailbox, _ := json.Marshal([]mailboxItem{{CID: cid}})
	if err := os.WriteFile(filepath.Join(root, "mailboxes", fixture.Identities.OfferTo+".json"), mailbox, 0600); err != nil {
		t.Fatal(err)
	}
	audit := `{"type":"envelope.upload","message_cid":"` + cid + `"}` + "\n" +
		`{"type":"envelope.fetch","message_cid":"` + cid + `"}` + "\n"
	if err := os.WriteFile(filepath.Join(root, "audit", "events.ndjson"), []byte(audit), 0600); err != nil {
		t.Fatal(err)
	}

	result, err := verify(root, cid, fixture.Identities.OfferTo, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if result == "" || result[:4] != "PASS" {
		t.Fatalf("unexpected result %q", result)
	}
}

func TestVerifyRejectsIncompleteEvidence(t *testing.T) {
	_, err := verify(t.TempDir(), "missing", "recipient", 1, 1)
	if err == nil {
		t.Fatal("verify accepted incomplete evidence")
	}
}
