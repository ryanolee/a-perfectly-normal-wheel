// Runnable demo of the SOPS Go SDK: encrypt ("hide") and decrypt ("unhide")
// secrets in-memory, using an age key generated on the fly so there is no
// GPG/KMS/keyfile setup to do.
//
//	go get github.com/getsops/sops/v3@latest filippo.io/age@latest
//	go run ./example
//
// Notes on the SDK:
//   - Decryption has a clean high-level entrypoint: decrypt.Data / decrypt.File.
//   - Encryption has no one-call helper, so you build a sops.Tree, generate a
//     data key, encrypt the tree, then emit it. That is what encrypt() shows.
//   - The YAML store preserves comments and key order. SOPS only encrypts the
//     VALUES, so keys and comments stay readable in the encrypted output.
package main

import (
	"fmt"
	"log"
	"os"

	"filippo.io/age"
	"github.com/getsops/sops/v3"
	"github.com/getsops/sops/v3/aes"
	sopsage "github.com/getsops/sops/v3/age"
	"github.com/getsops/sops/v3/cmd/sops/common"
	"github.com/getsops/sops/v3/decrypt"
	"github.com/getsops/sops/v3/keys"
	"github.com/getsops/sops/v3/stores/yaml"
	"github.com/getsops/sops/v3/version"
)

// Plain YAML with comments + a mix of secret and non-secret keys.
const plain = `# app config
db:
  host: db.internal      # not secret, stays visible
  password: hunter2      # secret value, will be encrypted
api_key: sk-live-abc123  # secret value, will be encrypted
`

func main() {
	// 1. Generate an age keypair. Recipient = public ("who can read"),
	//    identity = private key used to decrypt.
	id, err := age.GenerateX25519Identity()
	if err != nil {
		log.Fatal(err)
	}
	recipient := id.Recipient().String() // age1...
	identity := id.String()              // AGE-SECRET-KEY-...

	// 2. Encrypt ("hide").
	encrypted, err := encrypt([]byte(plain), recipient)
	if err != nil {
		log.Fatalf("encrypt: %v", err)
	}
	fmt.Println("=== ENCRYPTED (comments + keys preserved, values hidden) ===")
	fmt.Println(string(encrypted))

	// 3. Decrypt ("unhide"). decrypt.Data reads the age identity from the
	//    SOPS_AGE_KEY env var (or SOPS_AGE_KEY_FILE).
	os.Setenv("SOPS_AGE_KEY", identity)
	decrypted, err := decrypt.Data(encrypted, "yaml")
	if err != nil {
		log.Fatalf("decrypt: %v", err)
	}
	fmt.Println("=== DECRYPTED ===")
	fmt.Println(string(decrypted))
}

// encrypt mirrors what `sops --encrypt` does, but in-process.
func encrypt(plaintext []byte, ageRecipient string) ([]byte, error) {
	store := &yaml.Store{}

	branches, err := store.LoadPlainFile(plaintext)
	if err != nil {
		return nil, err
	}

	masterKey, err := sopsage.MasterKeyFromRecipient(ageRecipient)
	if err != nil {
		return nil, err
	}

	tree := sops.Tree{
		Branches: branches,
		Metadata: sops.Metadata{
			Version:   version.Version,
			KeyGroups: []sops.KeyGroup{{masterKey}},
			// EncryptedRegex/UnencryptedRegex would scope what gets hidden.
			// Omitted here, so every value is encrypted.
		},
	}

	// Data key is the symmetric key for values; it is itself encrypted to each
	// master key (here, the age recipient) and stored in the file metadata.
	dataKey, errs := tree.GenerateDataKey()
	if len(errs) > 0 {
		return nil, fmt.Errorf("generate data key: %v", errs)
	}

	if err := common.EncryptTree(common.EncryptTreeOpts{
		DataKey: dataKey,
		Tree:    &tree,
		Cipher:  aes.NewCipher(),
	}); err != nil {
		return nil, err
	}

	return store.EmitEncryptedFile(tree)
}

// keep the keys import referenced for readers grepping the type used above.
var _ keys.MasterKey
