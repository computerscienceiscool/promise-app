package protocol

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
)

type vectorFile struct {
	ProtocolCID string `json:"protocol_cid"`
	Signers     struct {
		OfferSeedHex  string `json:"offer_seed_hex"`
		AcceptSeedHex string `json:"accept_seed_hex"`
	} `json:"signers"`
	Identities struct {
		OfferFrom string `json:"offer_from"`
		OfferTo   string `json:"offer_to"`
	} `json:"identities"`
	Messages []struct {
		Name        string   `json:"name"`
		Kind        Kind     `json:"kind"`
		Body        string   `json:"body"`
		CreatedAt   uint64   `json:"created_at"`
		Invokes     []string `json:"invokes"`
		CID         string   `json:"cid"`
		EnvelopeHex string   `json:"envelope_hex"`
	} `json:"messages"`
}

func readVectors(t *testing.T) vectorFile {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "promise-messaging-v0-vectors.json"))
	if err != nil {
		t.Fatal(err)
	}
	var vectors vectorFile
	if err := json.Unmarshal(data, &vectors); err != nil {
		t.Fatal(err)
	}
	return vectors
}

func TestFrozenDescriptorAndVectors(t *testing.T) {
	vectors := readVectors(t)
	if vectors.ProtocolCID != FrozenProtocolCIDText || len(vectors.Messages) != 2 {
		t.Fatal("vectors do not identify the frozen protocol or expected messages")
	}
	descriptor, err := os.ReadFile(filepath.Join("..", "..", "testdata", "promise-messaging-v0-descriptor.json"))
	if err != nil {
		t.Fatal(err)
	}
	derived, err := cidFor(descriptor, cid.Raw)
	if err != nil {
		t.Fatal(err)
	}
	if !derived.Equals(FrozenProtocolCID()) {
		t.Fatalf("descriptor CID = %s, want %s", derived, FrozenProtocolCID())
	}

	offerSeed, _ := hex.DecodeString(vectors.Signers.OfferSeedHex)
	acceptSeed, _ := hex.DecodeString(vectors.Signers.AcceptSeedHex)
	offerPrivate := ed25519.NewKeyFromSeed(offerSeed)
	acceptPrivate := ed25519.NewKeyFromSeed(acceptSeed)
	from, _ := IdentityCID(offerPrivate.Public().(ed25519.PublicKey))
	to, _ := IdentityCID(acceptPrivate.Public().(ed25519.PublicKey))
	if from.String() != vectors.Identities.OfferFrom || to.String() != vectors.Identities.OfferTo {
		t.Fatal("fixed signing keys produced unexpected identities")
	}

	for index, vector := range vectors.Messages {
		encoded, err := hex.DecodeString(vector.EnvelopeHex)
		if err != nil {
			t.Fatalf("%s hex: %v", vector.Name, err)
		}
		candidate, err := Verify(encoded)
		if err != nil {
			t.Fatalf("%s did not verify: %v", vector.Name, err)
		}
		if candidate.CID.String() != vector.CID || !candidate.Protocol.Equals(FrozenProtocolCID()) {
			t.Fatalf("%s identifiers differ", vector.Name)
		}

		private, sender, recipient := offerPrivate, from, to
		if index == 1 {
			private, sender, recipient = acceptPrivate, to, from
		}
		invokes := make([]cbor.Tag, len(vector.Invokes))
		for i, invoked := range vector.Invokes {
			value, err := cid.Parse(invoked)
			if err != nil {
				t.Fatal(err)
			}
			invokes[i] = CIDTag(value)
		}
		built, builtCID, err := Build(FrozenProtocolCID(), private, Payload{Kind: vector.Kind, From: CIDTag(sender), To: CIDTag(recipient), Body: vector.Body, Invokes: invokes, CreatedAt: vector.CreatedAt})
		if err != nil {
			t.Fatalf("%s build: %v", vector.Name, err)
		}
		if !bytes.Equal(built, encoded) || builtCID.String() != vector.CID {
			t.Fatalf("%s is not byte-for-byte reproducible", vector.Name)
		}
	}
}

func TestVectorRejectionCases(t *testing.T) {
	vectors := readVectors(t)
	valid, _ := hex.DecodeString(vectors.Messages[0].EnvelopeHex)
	changed := append([]byte(nil), valid...)
	changed[len(changed)-1] ^= 1
	if _, err := Verify(changed); err == nil {
		t.Fatal("accepted changed vector bytes")
	}

	private := ed25519.NewKeyFromSeed(make([]byte, ed25519.SeedSize))
	from, _ := IdentityCID(private.Public().(ed25519.PublicKey))
	to, _ := cid.Parse(vectors.Identities.OfferTo)
	wrong, _ := cid.Parse("bafkreigh2akiscaildc5hiywzbllsgrcfqew2gozbkkfu4a2yv56zzrxmu")
	encoded, _, err := Build(wrong, private, Payload{Kind: Offer, From: CIDTag(from), To: CIDTag(to), Body: "wrong selector", CreatedAt: 1})
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := Verify(encoded)
	if err != nil || IsFrozenProtocol(candidate.Protocol) {
		t.Fatal("wrong selector was not distinguishable from frozen protocol")
	}
}
