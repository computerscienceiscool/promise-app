package store

import (
	"bytes"
	"testing"
	"time"
)

func TestStoreKeepsExactBytesAndIdempotentMailbox(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if created, err := s.Put("btest", []byte{1, 2, 3}); err != nil || !created {
		t.Fatalf("Put: %v %v", created, err)
	}
	if created, err := s.Put("btest", []byte{1, 2, 3}); err != nil || created {
		t.Fatalf("second Put: %v %v", created, err)
	}
	got, err := s.Get("btest")
	if err != nil || !bytes.Equal(got, []byte{1, 2, 3}) {
		t.Fatalf("Get: %x %v", got, err)
	}
	if _, err := s.Put("btest", []byte{9}); err == nil {
		t.Fatal("collision accepted")
	}
	if added, err := s.AddMailbox("bdevice", "btest"); err != nil || !added {
		t.Fatalf("AddMailbox: %v %v", added, err)
	}
	if added, err := s.AddMailbox("bdevice", "btest"); err != nil || added {
		t.Fatalf("duplicate mailbox: %v %v", added, err)
	}
	if err := s.Audit(AuditEvent{Time: time.Now(), Type: "upload", Actor: "bdevice", MessageCID: "btest"}); err != nil {
		t.Fatal(err)
	}
	if events, err := s.AuditLog(); err != nil || len(events) != 1 {
		t.Fatalf("AuditLog: %#v %v", events, err)
	}
}
