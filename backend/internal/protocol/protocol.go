// Package protocol implements the narrow, deterministic Promise Messaging v0
// POC format. It is intentionally separate from HTTP and persistence.
package protocol

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"
	"sort"
	"unicode/utf8"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

const (
	GridTag          = 1735551332
	MaxEnvelopeBytes = 128 * 1024
	MaxBodyBytes     = 4096
	// FrozenProtocolCIDText identifies the exact descriptor retained in
	// backend/testdata/promise-messaging-v0-descriptor.json. It is the only
	// protocol selector accepted by the POC relay.
	FrozenProtocolCIDText        = "bafkreicutumpjz3skwoehjn3gptve5m3atgypmfbsdhbtapg2rcwt2c2hi"
	maxInvocations               = 4
	identityCodec         uint64 = 0x300001
)

// FrozenProtocolCID returns the version-zero protocol selector. The constant
// is parsed at the boundary rather than exposing a mutable global CID value.
func FrozenProtocolCID() cid.Cid {
	value, err := cid.Parse(FrozenProtocolCIDText)
	if err != nil {
		panic("invalid frozen Promise Messaging v0 protocol CID: " + err.Error())
	}
	return value
}

// IsFrozenProtocol reports whether a candidate selects this POC's one frozen
// protocol. Parsing and signature validation are intentionally separate from
// local protocol support, so callers can retain unknown envelopes as opaque
// bytes if their policy permits it.
func IsFrozenProtocol(value cid.Cid) bool { return value.Equals(FrozenProtocolCID()) }

var (
	canonical cbor.EncMode
	strict    cbor.DecMode
)

func init() {
	var err error
	canonical, err = cbor.CanonicalEncOptions().EncMode()
	if err != nil {
		panic(err)
	}
	strict, err = (cbor.DecOptions{
		DupMapKey: cbor.DupMapKeyEnforcedAPF, IndefLength: cbor.IndefLengthForbidden,
		ExtraReturnErrors: cbor.ExtraDecErrorUnknownField, MaxArrayElements: 16,
		MaxMapPairs: 16, MaxNestedLevels: 16,
	}).DecMode()
	if err != nil {
		panic(err)
	}
}

type Kind string

const (
	Offer       Kind = "promise_offer"
	Accept      Kind = "promise_accept"
	Decline     Kind = "promise_decline"
	Observation Kind = "observation"
)

type Payload struct {
	Kind      Kind       `cbor:"kind"`
	From      cbor.Tag   `cbor:"from"`
	To        cbor.Tag   `cbor:"to"`
	Body      string     `cbor:"body"`
	Invokes   []cbor.Tag `cbor:"invokes"`
	CreatedAt uint64     `cbor:"created_at"`
}

type proofHeader struct {
	Algorithm string `cbor:"algorithm"`
	PublicKey []byte `cbor:"public_key"`
}
type proof struct {
	Algorithm string `cbor:"algorithm"`
	PublicKey []byte `cbor:"public_key"`
	Signature []byte `cbor:"signature"`
}
type Candidate struct {
	CID      cid.Cid
	Protocol cid.Cid
	Payload  Payload
	Bytes    []byte
}

func cidTag(value cid.Cid) cbor.Tag {
	b := append([]byte{0}, value.Bytes()...)
	return cbor.Tag{Number: 42, Content: b}
}

func parseCIDTag(value cbor.Tag) (cid.Cid, error) {
	b, ok := value.Content.([]byte)
	if !ok || value.Number != 42 || len(b) < 2 || b[0] != 0 {
		return cid.Undef, fmt.Errorf("invalid CID tag")
	}
	parsed, err := cid.Cast(b[1:])
	if err != nil || parsed.Version() != 1 {
		return cid.Undef, fmt.Errorf("invalid CIDv1")
	}
	return parsed, nil
}

func cidFor(data []byte, codec uint64) (cid.Cid, error) {
	return cid.Prefix{Version: 1, Codec: codec, MhType: multihash.SHA2_256, MhLength: sha256.Size}.Sum(data)
}

// IdentityCID is the stable public device identity derived from an Ed25519 key.
func IdentityCID(publicKey ed25519.PublicKey) (cid.Cid, error) {
	if len(publicKey) != ed25519.PublicKeySize {
		return cid.Undef, fmt.Errorf("invalid public key")
	}
	record, err := canonical.Marshal(struct {
		Version   uint8  `cbor:"version"`
		PublicKey []byte `cbor:"public_key"`
	}{0, publicKey})
	if err != nil {
		return cid.Undef, err
	}
	return cidFor(record, identityCodec)
}

func validatePayload(value Payload, signer cid.Cid) error {
	if value.Kind != Offer && value.Kind != Accept && value.Kind != Decline && value.Kind != Observation {
		return fmt.Errorf("unknown message kind")
	}
	if !utf8.ValidString(value.Body) || len([]byte(value.Body)) == 0 || len([]byte(value.Body)) > MaxBodyBytes {
		return fmt.Errorf("invalid body")
	}
	from, err := parseCIDTag(value.From)
	if err != nil {
		return fmt.Errorf("from: %w", err)
	}
	to, err := parseCIDTag(value.To)
	if err != nil {
		return fmt.Errorf("to: %w", err)
	}
	if from.Equals(to) || !from.Equals(signer) {
		return fmt.Errorf("sender identity mismatch")
	}
	if len(value.Invokes) > maxInvocations {
		return fmt.Errorf("too many invocations")
	}
	invokes := make([]cid.Cid, len(value.Invokes))
	for i, tag := range value.Invokes {
		invokes[i], err = parseCIDTag(tag)
		if err != nil {
			return fmt.Errorf("invocation %d: %w", i, err)
		}
	}
	sort.Slice(invokes, func(i, j int) bool { return bytes.Compare(invokes[i].Bytes(), invokes[j].Bytes()) < 0 })
	for i := range invokes {
		if i > 0 && invokes[i-1].Equals(invokes[i]) {
			return fmt.Errorf("duplicate invocation")
		}
	}
	if value.Kind == Offer && len(invokes) != 0 {
		return fmt.Errorf("offer must not invoke another message")
	}
	if (value.Kind == Accept || value.Kind == Decline) && len(invokes) != 1 {
		return fmt.Errorf("response must invoke one offer")
	}
	if value.Kind == Observation && len(invokes) == 0 {
		return fmt.Errorf("observation requires evidence")
	}
	return nil
}

func projection(selector cbor.Tag, payloadBytes []byte, header proofHeader) ([]byte, error) {
	headerBytes, err := canonical.Marshal(header)
	if err != nil {
		return nil, err
	}
	return canonical.Marshal(cbor.Tag{Number: GridTag, Content: []any{selector, payloadBytes, headerBytes}})
}

// Build signs one exact envelope. The caller must provide the frozen protocol CID.
func Build(protocolID cid.Cid, privateKey ed25519.PrivateKey, payload Payload) ([]byte, cid.Cid, error) {
	if !protocolID.Defined() || len(privateKey) != ed25519.PrivateKeySize {
		return nil, cid.Undef, fmt.Errorf("protocol ID and private key are required")
	}
	public := privateKey.Public().(ed25519.PublicKey)
	// A missing invokes field and an empty invokes array have different CBOR
	// encodings. The v0 schema requires an array, including for offers.
	payload.Invokes = append([]cbor.Tag{}, payload.Invokes...)
	signer, err := IdentityCID(public)
	if err != nil {
		return nil, cid.Undef, err
	}
	if err := validatePayload(payload, signer); err != nil {
		return nil, cid.Undef, err
	}
	payloadBytes, err := canonical.Marshal(payload)
	if err != nil {
		return nil, cid.Undef, err
	}
	selector := cidTag(protocolID)
	header := proofHeader{Algorithm: "Ed25519", PublicKey: public}
	toSign, err := projection(selector, payloadBytes, header)
	if err != nil {
		return nil, cid.Undef, err
	}
	proofBytes, err := canonical.Marshal(proof{Algorithm: header.Algorithm, PublicKey: public, Signature: ed25519.Sign(privateKey, toSign)})
	if err != nil {
		return nil, cid.Undef, err
	}
	encoded, err := canonical.Marshal(cbor.Tag{Number: GridTag, Content: []any{selector, payloadBytes, proofBytes}})
	if err != nil {
		return nil, cid.Undef, err
	}
	messageID, err := cidFor(encoded, cid.Raw)
	if err != nil {
		return nil, cid.Undef, err
	}
	return encoded, messageID, nil
}

// Verify performs format, canonical-byte, proof, and sender-identity checks.
func Verify(data []byte) (Candidate, error) {
	if len(data) == 0 || len(data) > MaxEnvelopeBytes {
		return Candidate{}, fmt.Errorf("invalid envelope size")
	}
	var outer cbor.Tag
	if err := strict.Unmarshal(data, &outer); err != nil || outer.Number != GridTag {
		return Candidate{}, fmt.Errorf("invalid grid envelope")
	}
	slots, ok := outer.Content.([]any)
	if !ok || len(slots) != 3 {
		return Candidate{}, fmt.Errorf("envelope must have three slots")
	}
	selector, ok := slots[0].(cbor.Tag)
	if !ok {
		return Candidate{}, fmt.Errorf("invalid protocol selector")
	}
	protocolID, err := parseCIDTag(selector)
	if err != nil {
		return Candidate{}, fmt.Errorf("protocol selector: %w", err)
	}
	payloadBytes, ok := slots[1].([]byte)
	if !ok {
		return Candidate{}, fmt.Errorf("invalid payload slot")
	}
	proofBytes, ok := slots[2].([]byte)
	if !ok {
		return Candidate{}, fmt.Errorf("invalid proof slot")
	}
	var payload Payload
	if err := strict.Unmarshal(payloadBytes, &payload); err != nil {
		return Candidate{}, fmt.Errorf("payload: %w", err)
	}
	canonicalPayload, _ := canonical.Marshal(payload)
	if !bytes.Equal(payloadBytes, canonicalPayload) {
		return Candidate{}, fmt.Errorf("payload is not canonical")
	}
	var signed proof
	if err := strict.Unmarshal(proofBytes, &signed); err != nil {
		return Candidate{}, fmt.Errorf("proof: %w", err)
	}
	if signed.Algorithm != "Ed25519" || len(signed.PublicKey) != ed25519.PublicKeySize || len(signed.Signature) != ed25519.SignatureSize {
		return Candidate{}, fmt.Errorf("invalid proof")
	}
	canonicalProof, _ := canonical.Marshal(signed)
	if !bytes.Equal(proofBytes, canonicalProof) {
		return Candidate{}, fmt.Errorf("proof is not canonical")
	}
	toVerify, err := projection(selector, payloadBytes, proofHeader{Algorithm: signed.Algorithm, PublicKey: signed.PublicKey})
	if err != nil {
		return Candidate{}, err
	}
	if !ed25519.Verify(ed25519.PublicKey(signed.PublicKey), toVerify, signed.Signature) {
		return Candidate{}, fmt.Errorf("invalid signature")
	}
	signer, err := IdentityCID(ed25519.PublicKey(signed.PublicKey))
	if err != nil {
		return Candidate{}, err
	}
	if err := validatePayload(payload, signer); err != nil {
		return Candidate{}, err
	}
	canonicalEnvelope, _ := canonical.Marshal(cbor.Tag{Number: GridTag, Content: []any{selector, payloadBytes, proofBytes}})
	if !bytes.Equal(data, canonicalEnvelope) {
		return Candidate{}, fmt.Errorf("envelope is not canonical")
	}
	messageID, err := cidFor(data, cid.Raw)
	if err != nil {
		return Candidate{}, err
	}
	return Candidate{CID: messageID, Protocol: protocolID, Payload: payload, Bytes: bytes.Clone(data)}, nil
}

func CIDTag(value cid.Cid) cbor.Tag               { return cidTag(value) }
func ParseCIDTag(value cbor.Tag) (cid.Cid, error) { return parseCIDTag(value) }
