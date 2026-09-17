package cryptox

import "testing"

const testKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestRoundTrip(t *testing.T) {
	c, err := New(testKey)
	if err != nil {
		t.Fatal(err)
	}
	secret := "sk-proj-abc123def456"

	enc, err := c.Encrypt(secret)
	if err != nil {
		t.Fatal(err)
	}
	if enc == secret {
		t.Fatal("ciphertext equals plaintext")
	}
	got, err := c.Decrypt(enc)
	if err != nil {
		t.Fatal(err)
	}
	if got != secret {
		t.Fatalf("got %q want %q", got, secret)
	}
}

// GCM uses a random nonce, so the same input must never produce the same
// ciphertext -- otherwise identical keys would be linkable in the database.
func TestEncryptIsNonDeterministic(t *testing.T) {
	c, _ := New(testKey)
	a, _ := c.Encrypt("same-input")
	b, _ := c.Encrypt("same-input")
	if a == b {
		t.Fatal("encrypting twice produced identical ciphertext")
	}
}

func TestDecryptRejectsTamperedCiphertext(t *testing.T) {
	c, _ := New(testKey)
	enc, _ := c.Encrypt("secret")
	tampered := []byte(enc)
	tampered[len(tampered)-2] ^= 0x01
	if _, err := c.Decrypt(string(tampered)); err == nil {
		t.Fatal("tampered ciphertext decrypted without error")
	}
}

func TestDecryptRejectsWrongKey(t *testing.T) {
	a, _ := New(testKey)
	b, _ := New("f" + testKey[1:])
	enc, _ := a.Encrypt("secret")
	if _, err := b.Decrypt(enc); err == nil {
		t.Fatal("decrypted with the wrong key")
	}
}

func TestNewRejectsBadKeys(t *testing.T) {
	for _, k := range []string{"", "tooshort", "zz" + testKey[2:]} {
		if _, err := New(k); err == nil {
			t.Fatalf("accepted invalid key %q", k)
		}
	}
}

func TestMaskKeepsEndsOnly(t *testing.T) {
	if got := Mask("sk-proj-abcdefghijkl"); got != "sk-p••••ijkl" {
		t.Fatalf("got %q", got)
	}
	if got := Mask("short"); got != "••••" {
		t.Fatalf("short key leaked: %q", got)
	}
}
