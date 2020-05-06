package wallet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"io"
	"os"
	"syscall"

	"github.com/dgraph-io/badger"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/ssh/terminal"
)

func AskPass() ([]byte, error) {
	var fd int
	if terminal.IsTerminal(syscall.Stdin) {
		fd = syscall.Stdin
	} else {
		tty, err := os.Open("/dev/tty")
		if err != nil {
			return nil, errors.Wrap(err, "error allocating terminal")
		}
		defer tty.Close()
		fd = int(tty.Fd())
	}

	pass, err := terminal.ReadPassword(fd)
	if err != nil {
		return nil, errors.Wrap(err, "error reading password")
	}
	fmt.Println()

	return pass, nil
}

func (w *Wallet) initializeWallet() error {

	fmt.Println("Creating new wallet...")

	var password []byte

	for {
		fmt.Printf("Enter a password: ")
		pass, err := AskPass()
		if err != nil {
			return err
		}
		fmt.Printf("Re-enter the password: ")
		passVerify, err := AskPass()
		if err != nil {
			return err
		}

		if bytes.Equal(pass, passVerify) {
			password = pass
			break
		} else {
			fmt.Println("Passwords do not match. Please try again.")
		}
	}

	// generate random salt
	var salt [8]byte
	_, err := rand.Reader.Read(salt[:])
	if err != nil {
		return errors.Wrap(err, "error reading from random")
	}
	encryptionKey := pbkdf2.Key(password, salt[:], 20000, 32, sha512.New)

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return errors.Wrap(err, "error creating cipher")
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return errors.Wrap(err, "error reading from random")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return errors.Wrap(err, "error creating GCM")
	}

	var privateKeyBytes [32]byte
	if _, err := io.ReadFull(rand.Reader, privateKeyBytes[:]); err != nil {
		return errors.Wrap(err, "error reading from random")
	}

	privateKey := bls.DeriveSecretKey(privateKeyBytes)
	address, err := privateKey.DerivePublicKey().ToBech32(w.params.AddressPrefixes)
	if err != nil {
		return errors.Wrap(err, "could not get public key from private key")
	}

	ciphertext := aesgcm.Seal(nil, nonce, privateKeyBytes[:], nil)

	err = w.db.Update(func(tx *badger.Txn) error {
		if err := tx.Set(walletDBKey, ciphertext[:]); err != nil {
			return err
		}

		if err := tx.Set(walletDBAddress, []byte(address)); err != nil {
			return err
		}

		if err := tx.Set(walletDBSalt, salt[:]); err != nil {
			return err
		}

		if err := tx.Set(walletDBLastTxNonce, []byte{0, 0, 0, 0, 0, 0, 0, 0}); err != nil {
			return err
		}

		return tx.Set(walletDBNonce, nonce[:])
	})
	if err != nil {
		return err
	}

	w.info.encryptedMaster = ciphertext
	w.info.salt = salt[:]
	w.info.nonce = nonce
	w.info.lastNonce = 0
	w.hasMaster = true

	return nil
}
