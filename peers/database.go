package peers

import (
	"crypto/rand"
	"fmt"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"go.etcd.io/bbolt"
	"strconv"
)

// contains several functions that interact with netDB database

func SavePeer(netDB *bbolt.DB, pma multiaddr.Multiaddr) error {
	err := netDB.Update(func(tx *bbolt.Tx) error {
		var err error
		b := tx.Bucket(configBucketKey).Bucket(peersDbKey)
		if b == nil {
			b, err = tx.Bucket(configBucketKey).CreateBucketIfNotExists(peersDbKey)
			if err != nil {
				return err
			}
		}
		err = b.Put(pma.Bytes(), []byte(strconv.Itoa(initialBanScore)))
		return err
	})
	return err
}

// Reduces the banscore of a peer. If it reaches limit, it will be banned
func BanscorePeer(netDB *bbolt.DB, id peer.ID) error {
	err := netDB.Update(func(tx *bbolt.Tx) error {
		var err error
		savedDb := tx.Bucket(configBucketKey).Bucket(peersDbKey)
		if savedDb == nil {
			savedDb, err = tx.Bucket(configBucketKey).CreateBucketIfNotExists(peersDbKey)
			if err != nil {
				return err
			}
		}
		bansDb := tx.Bucket(configBucketKey).Bucket(bansDbKey)
		if bansDb == nil {
			bansDb, err = tx.Bucket(configBucketKey).CreateBucketIfNotExists(bansDbKey)
			if err != nil {
				return err
			}
		}
		err = savedDb.ForEach(func(k, v []byte) error {
			addr, _ := multiaddr.NewMultiaddrBytes(k)
			parsedPeer, err := peer.AddrInfoFromP2pAddr(addr)
			if err == nil {
				if parsedPeer.ID == id {
					// reduce score. If it reaches 0, ban
					score, _ := strconv.Atoi(string(v))
					fmt.Printf("peer %s has a banscore of: %s", id.String(), strconv.Itoa(score))
					score -= 1
					if score == 0 {

						// add to banlist
						byteId, err := parsedPeer.ID.MarshalBinary()
						if err == nil {
							_ = bansDb.Put(byteId, nil)
						}
						// remove from saved list
						err = savedDb.Delete(k)
					} else {
						_ = savedDb.Put(k, []byte(strconv.Itoa(score)))
					}
				}
			}
			return nil
		})
		return err
	})
	return err
}

func IsPeerBanned(netDB *bbolt.DB, id peer.ID) bool {
	var res bool
	_ = netDB.Update(func(tx *bbolt.Tx) error {
		var err error
		b := tx.Bucket(configBucketKey).Bucket(bansDbKey)
		if b == nil {
			b, err = tx.Bucket(configBucketKey).CreateBucketIfNotExists(bansDbKey)
			if err != nil {
				return err
			}
		}
		byteId, err := id.MarshalBinary()
		if err != nil {
			return err
		}
		bannedPeerId := b.Get(byteId)
		if bannedPeerId != nil {
			res = true
			return nil
		}
		return nil
	})
	return res
}

func GetSavedPeers(netDB *bbolt.DB) (savedAddresses []multiaddr.Multiaddr, err error) {
	//retrieve the saved addresses
	err = netDB.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(configBucketKey).Bucket(peersDbKey)
		if b == nil {
			b, err = tx.Bucket(configBucketKey).CreateBucketIfNotExists(peersDbKey)
			if err != nil {
				return err
			}
		}
		bannedP := tx.Bucket(configBucketKey).Bucket(bansDbKey)
		if bannedP == nil {
			bannedP, err = tx.Bucket(configBucketKey).CreateBucketIfNotExists(bansDbKey)
			if err != nil {
				return err
			}
		}
		_ = b.ForEach(func(k, v []byte) error {
			addr, err := multiaddr.NewMultiaddrBytes(k)
			if err == nil {
				peerId, err := peer.AddrInfoFromP2pAddr(addr)
				if err != nil {
					fmt.Println("120: " + err.Error() + ", removing from db")
					// if the saved peer cannot be validated, delete
					err = b.Delete(k)
				} else {
					byteId, err := peerId.ID.MarshalBinary()
					if err != nil {
						fmt.Println("124: " + err.Error())
					} else {
						bannedPeerId := bannedP.Get(byteId)
						if bannedPeerId == nil {
							savedAddresses = append(savedAddresses, addr)
						}
					}

				}
			}
			return nil
		})
		return err
	})
	return
}

func GetPrivKey(netDB *bbolt.DB) (priv crypto.PrivKey, configBucket *bbolt.Bucket, err error) {
	err = netDB.Update(func(tx *bbolt.Tx) error {
		configBucket = tx.Bucket(configBucketKey)
		// If the bucket doesn't exist, initialize the database
		if configBucket == nil {
			configBucket, err = tx.CreateBucketIfNotExists(configBucketKey)
			if err != nil {
				return err
			}

		}
		var keyBytes []byte
		keyBytes = configBucket.Get(privKeyDbKey)
		if keyBytes == nil {
			priv, _, err = crypto.GenerateEd25519Key(rand.Reader)
			if err != nil {
				return err
			}
			privBytes, err := crypto.MarshalPrivateKey(priv)
			if err != nil {
				return err
			}
			err = configBucket.Put(privKeyDbKey, privBytes)
			if err != nil {
				return err
			}
			keyBytes = privBytes
		}

		key, err := crypto.UnmarshalPrivateKey(keyBytes)
		if err != nil {
			return err
		}

		priv = key

		return nil
	})
	return
}
