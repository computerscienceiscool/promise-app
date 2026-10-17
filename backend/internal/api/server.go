// Package api exposes the deliberately small POC relay API.
package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/promisegrid/promise-app/backend/internal/protocol"
	"github.com/promisegrid/promise-app/backend/internal/store"
)

type Server struct {
	store      *store.Store
	adminToken string
	logger     *slog.Logger
}

func New(data *store.Store, adminToken string, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{store: data, adminToken: adminToken, logger: logger}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("POST /v1/envelopes", s.upload)
	mux.HandleFunc("GET /v1/envelopes/{cid}", s.get)
	mux.HandleFunc("GET /v1/mailbox", s.mailbox)
	mux.HandleFunc("GET /v1/admin/audit", s.audit)
	return requestLog(s.logger, mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) upload(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, protocol.MaxEnvelopeBytes))
	if err != nil {
		problem(w, http.StatusRequestEntityTooLarge, "body too large")
		return
	}
	candidate, err := protocol.Verify(body)
	if err != nil {
		problem(w, http.StatusBadRequest, "invalid envelope: "+err.Error())
		return
	}
	if !protocol.IsFrozenProtocol(candidate.Protocol) {
		problem(w, http.StatusBadRequest, "unsupported protocol selector")
		return
	}
	actor := r.Header.Get("X-Device-ID")
	from, err := protocol.ParseCIDTag(candidate.Payload.From)
	if err != nil || actor == "" || actor != from.String() {
		problem(w, http.StatusUnauthorized, "X-Device-ID must be the signed sender")
		return
	}
	to, _ := protocol.ParseCIDTag(candidate.Payload.To)
	created, err := s.store.Put(candidate.CID.String(), candidate.Bytes)
	if err != nil {
		problem(w, http.StatusInternalServerError, "store envelope")
		return
	}
	mailboxAdded, err := s.store.AddMailbox(to.String(), candidate.CID.String())
	if err != nil {
		problem(w, http.StatusInternalServerError, "store mailbox")
		return
	}
	_ = s.store.Audit(store.AuditEvent{Time: time.Now(), Type: "envelope.upload", Actor: actor, MessageCID: candidate.CID.String(), Detail: "created=" + boolText(created) + ", mailbox_added=" + boolText(mailboxAdded)})
	writeJSON(w, http.StatusCreated, map[string]any{"cid": candidate.CID.String(), "created": created, "mailbox_added": mailboxAdded})
}

func (s *Server) mailbox(w http.ResponseWriter, r *http.Request) {
	identity := r.Header.Get("X-Device-ID")
	if identity == "" {
		problem(w, http.StatusUnauthorized, "X-Device-ID is required")
		return
	}
	items, err := s.store.Mailbox(identity)
	if err != nil {
		problem(w, http.StatusInternalServerError, "read mailbox")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	identity, messageID := r.Header.Get("X-Device-ID"), r.PathValue("cid")
	if identity == "" {
		problem(w, http.StatusUnauthorized, "X-Device-ID is required")
		return
	}
	allowed, err := s.store.HasMailboxItem(identity, messageID)
	if err != nil {
		problem(w, http.StatusInternalServerError, "read mailbox")
		return
	}
	if !allowed {
		problem(w, http.StatusForbidden, "message is not in this mailbox")
		return
	}
	data, err := s.store.Get(messageID)
	if err != nil {
		problem(w, http.StatusNotFound, "message not found")
		return
	}
	w.Header().Set("Content-Type", "application/cbor")
	w.Header().Set("Content-Length", stringInt(len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	_ = s.store.Audit(store.AuditEvent{Time: time.Now(), Type: "envelope.fetch", Actor: identity, MessageCID: messageID})
}

func (s *Server) audit(w http.ResponseWriter, r *http.Request) {
	if s.adminToken == "" || r.Header.Get("Authorization") != "Bearer "+s.adminToken {
		problem(w, http.StatusUnauthorized, "administrator token required")
		return
	}
	events, err := s.store.AuditLog()
	if err != nil {
		problem(w, http.StatusInternalServerError, "read audit log")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": events})
}

func requestLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start).String())
	})
}
func problem(w http.ResponseWriter, status int, detail string) {
	writeJSON(w, status, map[string]string{"error": detail})
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func boolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
func stringInt(value int) string { return strconv.Itoa(value) }
