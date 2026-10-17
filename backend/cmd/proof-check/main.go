// Command proof-check verifies retained POC evidence for one message CID.
// It does not infer user acceptance, message ordering, or offline state.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/promisegrid/promise-app/backend/internal/protocol"
)

type auditEvent struct {
	Type       string `json:"type"`
	Actor      string `json:"actor"`
	MessageCID string `json:"message_cid"`
	Detail     string `json:"detail"`
}

type mailboxItem struct {
	CID string `json:"cid"`
}

func main() {
	dataDir := flag.String("data-dir", "", "relay PROMISE_DATA_DIR")
	cid := flag.String("cid", "", "message CID to inspect")
	recipient := flag.String("recipient", "", "recipient device identity CID")
	minUploads := flag.Int("min-uploads", 1, "minimum envelope.upload events for the message")
	minFetches := flag.Int("min-fetches", 1, "minimum envelope.fetch events for the message")
	flag.Parse()

	if *dataDir == "" || *cid == "" || *recipient == "" {
		fail(errors.New("-data-dir, -cid, and -recipient are required"))
	}
	if *minUploads < 1 || *minFetches < 0 {
		fail(errors.New("minimum event counts must be non-negative, and uploads must be at least one"))
	}

	result, err := verify(*dataDir, *cid, *recipient, *minUploads, *minFetches)
	if err != nil {
		fail(err)
	}
	fmt.Println(result)
}

func verify(dataDir, cid, recipient string, minUploads, minFetches int) (string, error) {
	objectPath := filepath.Join(dataDir, "objects", cid+".cbor")
	bytes, err := os.ReadFile(objectPath)
	if err != nil {
		return "", fmt.Errorf("read immutable object %q: %w", objectPath, err)
	}
	candidate, err := protocol.Verify(bytes)
	if err != nil {
		return "", fmt.Errorf("verify immutable object: %w", err)
	}
	if candidate.CID.String() != cid {
		return "", fmt.Errorf("object CID mismatch: requested %s, computed %s", cid, candidate.CID.String())
	}

	mailboxPath := filepath.Join(dataDir, "mailboxes", recipient+".json")
	mailboxBytes, err := os.ReadFile(mailboxPath)
	if err != nil {
		return "", fmt.Errorf("read recipient mailbox %q: %w", mailboxPath, err)
	}
	var items []mailboxItem
	if err := json.Unmarshal(mailboxBytes, &items); err != nil {
		return "", fmt.Errorf("decode recipient mailbox: %w", err)
	}
	mailboxMatches := 0
	for _, item := range items {
		if item.CID == cid {
			mailboxMatches++
		}
	}
	if mailboxMatches != 1 {
		return "", fmt.Errorf("expected exactly one mailbox entry for %s, found %d", cid, mailboxMatches)
	}

	auditPath := filepath.Join(dataDir, "audit", "events.ndjson")
	audit, err := os.Open(auditPath)
	if err != nil {
		return "", fmt.Errorf("open audit evidence %q: %w", auditPath, err)
	}
	defer audit.Close()
	uploads, fetches := 0, 0
	scanner := bufio.NewScanner(audit)
	for scanner.Scan() {
		var event auditEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return "", fmt.Errorf("decode audit event: %w", err)
		}
		if event.MessageCID != cid {
			continue
		}
		switch event.Type {
		case "envelope.upload":
			uploads++
		case "envelope.fetch":
			fetches++
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read audit evidence: %w", err)
	}
	if uploads < minUploads || fetches < minFetches {
		return "", fmt.Errorf("audit counts for %s: uploads=%d (need >=%d), fetches=%d (need >=%d)", cid, uploads, minUploads, fetches, minFetches)
	}
	return fmt.Sprintf("PASS cid=%s bytes=%d mailbox_entries=%d uploads=%d fetches=%d", cid, len(bytes), mailboxMatches, uploads, fetches), nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "FAIL:", err)
	os.Exit(1)
}
