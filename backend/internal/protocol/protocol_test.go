package protocol

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/ipfs/go-cid"
)

func testIdentity(t *testing.T) (ed25519.PrivateKey, cid.Cid) {
	t.Helper()
	pub, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	id, err := IdentityCID(pub)
	if err != nil {
		t.Fatal(err)
	}
	return private, id
}

func TestBuildAndVerifyOffer(t *testing.T) {
	private, from := testIdentity(t)
	_, to := testIdentity(t)
	protocolID, err := cid.Parse("bafkreigh2akiscaildc5hiywzbllsgrcfqew2gozbkkfu4a2yv56zzrxmu")
	if err != nil {
		t.Fatal(err)
	}
	encoded, want, err := Build(protocolID, private, Payload{Kind: Offer, From: CIDTag(from), To: CIDTag(to), Body: "I promise to demonstrate the POC.", CreatedAt: 1})
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := Verify(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !candidate.CID.Equals(want) || candidate.Payload.Body != "I promise to demonstrate the POC." {
		t.Fatalf("unexpected candidate: %#v", candidate)
	}
}

func TestVerifyRejectsChangedBytes(t *testing.T) {
	private, from := testIdentity(t)
	_, to := testIdentity(t)
	protocolID, _ := cid.Parse("bafkreigh2akiscaildc5hiywzbllsgrcfqew2gozbkkfu4a2yv56zzrxmu")
	encoded, _, err := Build(protocolID, private, Payload{Kind: Offer, From: CIDTag(from), To: CIDTag(to), Body: "original", CreatedAt: 1})
	if err != nil {
		t.Fatal(err)
	}
	encoded[len(encoded)-1] ^= 1
	if _, err := Verify(encoded); err == nil {
		t.Fatal("Verify accepted altered bytes")
	}
}
