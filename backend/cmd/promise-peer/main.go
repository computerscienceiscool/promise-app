// Command promise-peer is a small desktop peer for the Promise App POC.
//
// It owns a separate Ed25519 device identity, reads its mailbox from a relay,
// and verifies the exact CBOR bytes returned by that relay before displaying
// them.  It intentionally does not treat retrieval as acceptance.
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/promisegrid/promise-app/backend/internal/protocol"
)

const defaultIdentityFile = ".promise-peer/identity.ed25519"

type identity struct{ private ed25519.PrivateKey }

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "identity":
		identityCommand(os.Args[2:])
	case "inbox":
		inboxCommand(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: promise-peer identity [-identity-file path] | inbox -relay URL [-identity-file path] [-inbox-dir path] [-evidence-file path] [-expect-cid message-cid]")
}

func identityCommand(args []string) {
	fs := flag.NewFlagSet("identity", flag.ExitOnError)
	path := fs.String("identity-file", defaultIdentityFile, "private identity file (created with mode 0600)")
	_ = fs.Parse(args)
	id, created, err := loadOrCreateIdentity(*path)
	if err != nil {
		fail(err)
	}
	cid, err := id.cid()
	if err != nil {
		fail(err)
	}
	if created {
		fmt.Fprintf(os.Stderr, "created private identity at %s\n", *path)
	}
	fmt.Println(cid)
}

func inboxCommand(args []string) {
	fs := flag.NewFlagSet("inbox", flag.ExitOnError)
	relay := fs.String("relay", "", "relay base URL, e.g. http://192.168.1.10:8080")
	identityPath := fs.String("identity-file", defaultIdentityFile, "private identity file")
	inboxDir := fs.String("inbox-dir", ".promise-peer/inbox", "directory to retain verified exact CBOR bytes")
	evidenceFile := fs.String("evidence-file", "", "optional JSON report of this verified fetch (mode 0600)")
	expectCID := fs.String("expect-cid", "", "require this exact verified message CID in the fetched mailbox")
	_ = fs.Parse(args)
	if *relay == "" {
		fail(errors.New("-relay is required"))
	}
	id, _, err := loadOrCreateIdentity(*identityPath)
	if err != nil {
		fail(err)
	}
	items, err := fetchInbox(http.DefaultClient, strings.TrimRight(*relay, "/"), id, *inboxDir)
	if err != nil {
		fail(err)
	}
	if *expectCID != "" && !containsCID(items, *expectCID) {
		fail(fmt.Errorf("expected message CID %q was not fetched and verified", *expectCID))
	}
	if *evidenceFile != "" {
		cid, err := id.cid()
		if err != nil {
			fail(err)
		}
		report := fetchEvidence{
			FetchedAt: time.Now().UTC().Format(time.RFC3339),
			Relay:     strings.TrimRight(*relay, "/"),
			Recipient: cid,
			Messages:  items,
		}
		if err := writeEvidence(*evidenceFile, report); err != nil {
			fail(err)
		}
	}
	for _, item := range items {
		encoded, _ := json.Marshal(item)
		fmt.Println(string(encoded))
	}
}

func containsCID(items []displayMessage, expected string) bool {
	for _, item := range items {
		if item.CID == expected {
			return true
		}
	}
	return false
}

// fetchEvidence deliberately contains no private key or relay admin token. It
// correlates a computer-side verification run with the retained exact bytes.
type fetchEvidence struct {
	FetchedAt string           `json:"fetched_at"`
	Relay     string           `json:"relay"`
	Recipient string           `json:"recipient"`
	Messages  []displayMessage `json:"messages"`
}

func writeEvidence(path string, report fetchEvidence) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create evidence directory: %w", err)
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode evidence: %w", err)
	}
	if err := os.WriteFile(path, append(encoded, '\n'), 0600); err != nil {
		return fmt.Errorf("write evidence: %w", err)
	}
	return nil
}

func loadOrCreateIdentity(path string) (identity, bool, error) {
	b, err := os.ReadFile(path)
	if err == nil {
		raw, err := base64.RawStdEncoding.DecodeString(strings.TrimSpace(string(b)))
		if err != nil || len(raw) != ed25519.PrivateKeySize {
			return identity{}, false, errors.New("identity file is not a valid Ed25519 private key")
		}
		return identity{private: ed25519.PrivateKey(raw)}, false, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return identity{}, false, fmt.Errorf("read identity: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return identity{}, false, fmt.Errorf("create identity directory: %w", err)
	}
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return identity{}, false, fmt.Errorf("generate identity: %w", err)
	}
	encoded := []byte(base64.RawStdEncoding.EncodeToString(key) + "\n")
	if err := os.WriteFile(path, encoded, 0600); err != nil {
		return identity{}, false, fmt.Errorf("write identity: %w", err)
	}
	return identity{private: key}, true, nil
}

func (id identity) cid() (string, error) {
	c, err := protocol.IdentityCID(id.private.Public().(ed25519.PublicKey))
	if err != nil {
		return "", err
	}
	return c.String(), nil
}

type mailboxResponse struct {
	Items []struct {
		CID string `json:"cid"`
	} `json:"items"`
}
type displayMessage struct {
	CID       string        `json:"cid"`
	Protocol  string        `json:"protocol"`
	From      string        `json:"from"`
	To        string        `json:"to"`
	Kind      protocol.Kind `json:"kind"`
	Body      string        `json:"body"`
	CreatedAt uint64        `json:"created_at"`
	Stored    string        `json:"stored"`
}

func fetchInbox(client *http.Client, relay string, id identity, inboxDir string) ([]displayMessage, error) {
	deviceID, err := id.cid()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, relay+"/v1/mailbox", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Device-ID", deviceID)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch mailbox: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, responseError(resp)
	}
	var mailbox mailboxResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&mailbox); err != nil {
		return nil, fmt.Errorf("decode mailbox: %w", err)
	}
	if err := os.MkdirAll(inboxDir, 0700); err != nil {
		return nil, fmt.Errorf("create inbox: %w", err)
	}
	out := make([]displayMessage, 0, len(mailbox.Items))
	for _, item := range mailbox.Items {
		if item.CID == "" || strings.Contains(item.CID, "/") {
			return nil, errors.New("relay returned invalid mailbox CID")
		}
		req, err := http.NewRequest(http.MethodGet, relay+"/v1/envelopes/"+item.CID, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("X-Device-ID", deviceID)
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetch %s: %w", item.CID, err)
		}
		data, readErr := io.ReadAll(io.LimitReader(resp.Body, protocol.MaxEnvelopeBytes+1))
		resp.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		if resp.StatusCode != http.StatusOK {
			return nil, responseBytesError(resp.StatusCode, data)
		}
		candidate, err := protocol.Verify(data)
		if err != nil {
			return nil, fmt.Errorf("verify %s: %w", item.CID, err)
		}
		if !protocol.IsFrozenProtocol(candidate.Protocol) {
			return nil, fmt.Errorf("verify %s: unsupported protocol", item.CID)
		}
		if candidate.CID.String() != item.CID {
			return nil, fmt.Errorf("verify %s: returned bytes have CID %s", item.CID, candidate.CID)
		}
		to, err := protocol.ParseCIDTag(candidate.Payload.To)
		if err != nil {
			return nil, err
		}
		if to.String() != deviceID {
			return nil, fmt.Errorf("verify %s: recipient mismatch", item.CID)
		}
		from, err := protocol.ParseCIDTag(candidate.Payload.From)
		if err != nil {
			return nil, err
		}
		path := filepath.Join(inboxDir, item.CID+".cbor")
		if err := os.WriteFile(path, data, 0600); err != nil {
			return nil, fmt.Errorf("retain %s: %w", item.CID, err)
		}
		out = append(out, displayMessage{CID: item.CID, Protocol: candidate.Protocol.String(), From: from.String(), To: to.String(), Kind: candidate.Payload.Kind, Body: candidate.Payload.Body, CreatedAt: candidate.Payload.CreatedAt, Stored: path})
	}
	return out, nil
}

func responseError(resp *http.Response) error {
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return responseBytesError(resp.StatusCode, b)
}
func responseBytesError(status int, body []byte) error {
	return fmt.Errorf("relay returned HTTP %d: %s", status, strings.TrimSpace(string(body)))
}
func fail(err error) { fmt.Fprintln(os.Stderr, "FAIL:", err); os.Exit(1) }
