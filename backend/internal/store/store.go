// Package store provides an intentionally inspectable filesystem-backed POC
// store. Object bytes are immutable; mailboxes and audit events are metadata.
package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type AuditEvent struct {
	Time       time.Time `json:"time"`
	Type       string    `json:"type"`
	Actor      string    `json:"actor"`
	MessageCID string    `json:"message_cid,omitempty"`
	Detail     string    `json:"detail,omitempty"`
}

type MailItem struct {
	CID         string    `json:"cid"`
	AvailableAt time.Time `json:"available_at"`
}
type Store struct {
	root string
	mu   sync.Mutex
}

func Open(root string) (*Store, error) {
	for _, path := range []string{"objects", "mailboxes", "audit"} {
		if err := os.MkdirAll(filepath.Join(root, path), 0700); err != nil {
			return nil, err
		}
	}
	return &Store{root: root}, nil
}

func (s *Store) objectPath(cid string) string { return filepath.Join(s.root, "objects", cid+".cbor") }
func (s *Store) mailboxPath(identity string) string {
	return filepath.Join(s.root, "mailboxes", identity+".json")
}

func (s *Store) Put(cid string, data []byte) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	path := s.objectPath(cid)
	if old, err := os.ReadFile(path); err == nil {
		if !bytes.Equal(old, data) {
			return false, fmt.Errorf("CID collision with different bytes")
		}
		return false, nil
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".object-*")
	if err != nil {
		return false, err
	}
	name := temporary.Name()
	defer os.Remove(name)
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return false, err
	}
	if err := temporary.Chmod(0600); err != nil {
		temporary.Close()
		return false, err
	}
	if err := temporary.Close(); err != nil {
		return false, err
	}
	if err := os.Rename(name, path); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) Get(cid string) ([]byte, error) { return os.ReadFile(s.objectPath(cid)) }

func (s *Store) AddMailbox(identity, cid string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	path := s.mailboxPath(identity)
	items := []MailItem{}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &items); err != nil {
			return false, fmt.Errorf("read mailbox: %w", err)
		}
	}
	for _, item := range items {
		if item.CID == cid {
			return false, nil
		}
	}
	items = append(items, MailItem{CID: cid, AvailableAt: time.Now().UTC()})
	sort.Slice(items, func(i, j int) bool { return items[i].AvailableAt.Before(items[j].AvailableAt) })
	return true, writeJSONAtomic(path, items)
}

func (s *Store) Mailbox(identity string) ([]MailItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.mailboxPath(identity))
	if os.IsNotExist(err) {
		return []MailItem{}, nil
	}
	if err != nil {
		return nil, err
	}
	var items []MailItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Store) HasMailboxItem(identity, cid string) (bool, error) {
	items, err := s.Mailbox(identity)
	if err != nil {
		return false, err
	}
	for _, item := range items {
		if item.CID == cid {
			return true, nil
		}
	}
	return false, nil
}

func (s *Store) Audit(event AuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	event.Time = event.Time.UTC()
	encoded, err := json.Marshal(event)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filepath.Join(s.root, "audit", "events.ndjson"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(append(encoded, '\n')); err != nil {
		return err
	}
	return file.Sync()
}

func (s *Store) AuditLog() ([]AuditEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(filepath.Join(s.root, "audit", "events.ndjson"))
	if os.IsNotExist(err) {
		return []AuditEvent{}, nil
	}
	if err != nil {
		return nil, err
	}
	lines := bytes.Split(bytes.TrimSpace(data), []byte{'\n'})
	events := make([]AuditEvent, 0, len(lines))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var event AuditEvent
		if err := json.Unmarshal(line, &event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

func writeJSONAtomic(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".metadata-*")
	if err != nil {
		return err
	}
	name := temporary.Name()
	defer os.Remove(name)
	if _, err := temporary.Write(append(data, '\n')); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Chmod(0600); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}
