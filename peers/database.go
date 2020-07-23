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

var configBucketKey = []byte("config")
var privKeyDbKey = []byte("privkey")
var peersDbKey = []byte("peers")
var bansDbKey = []byte("bans")

const BanLimit = 5
const BanMinScore = 1

func InitBuckets(netDB *bbolt.DB) (configBucket *bbolt.Bucket, err error) {
	err = netDB.Update(func(tx *bbolt.Tx) error {
		var err error
		configBucket = tx.Bucket(configBucketKey)
		if configBucket == nil {
			configBucket, err = tx.CreateBucketIfNotExists(configBucketKey)
			if err != nil {
				return err
			}
		}
		peersBucket := tx.Bucket(peersDbKey)
		if peersBucket == nil {
			peersBucket, err = tx.CreateBucketIfNotExists(peersDbKey)
			if err != nil {
				return err
			}
		}
		bansBucket := tx.Bucket(bansDbKey)
		if bansBucket == nil {
			bansBucket, err = tx.CreateBucketIfNotExists(bansDbKey)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return
}

func SavePeer(netDB *bbolt.DB, pma multiaddr.Multiaddr) error {
	err := netDB.Update(func(tx *bbolt.Tx) error {
		var err error
		b := tx.Bucket(peersDbKey)
		peerId, err := peer.AddrInfoFromP2pAddr(pma)
		if err != nil {
			return err
		}
		isBanned, err := IsPeerBanned(netDB, peerId.ID)
		if err != nil {
			return err
		}
		if !isBanned {
			err = b.Put(pma.Bytes(), []byte(strconv.Itoa(0)))
		} else {
			err = fmt.Errorf("peer %s is banned", peerId.ID.String())
		}
		return err
	})
	return err
}

// Reduces the banscore of a peer. If it reaches limit, it will be banned
func BanscorePeer(netDB *bbolt.DB, id peer.ID, weight int) error {
	err := netDB.Update(func(tx *bbolt.Tx) error {
		var err error
		savedDb := tx.Bucket(peersDbKey)
		bansDb := tx.Bucket(bansDbKey)

		err = savedDb.ForEach(func(k, v []byte) error {
			addr, err := multiaddr.NewMultiaddrBytes(k)
			if err != nil {
				return err
			}
			parsedPeer, err := peer.AddrInfoFromP2pAddr(addr)
			if err != nil {
				return err
			}
			if parsedPeer.ID == id {
				// reduce score. If it reaches 0, ban
				score, _ := strconv.Atoi(string(v))
				fmt.Printf("peer %s has a banscore of: %s \n", id.String(), strconv.Itoa(score))

				score += weight
				if score >= BanLimit {
					// add to banlist
					byteId, err := parsedPeer.ID.MarshalBinary()
					if err == nil {
						err = bansDb.Put(byteId, nil)
						if err != nil {
							return err
						}
						fmt.Printf("peer %s was banned \n", id.String())
					}
					// remove from saved list
					err = savedDb.Delete(k)
				} else {
					err = savedDb.Put(k, []byte(strconv.Itoa(score)))
					if err != nil {
						return err
					}
				}
			}
			return nil
		})
		return err
	})
	return err
}

func IsPeerBanned(netDB *bbolt.DB, id peer.ID) (bool, error) {
	var res bool
	err := netDB.View(func(tx *bbolt.Tx) error {
		var err error
		b := tx.Bucket(bansDbKey)
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
	return res, err
}

func GetSavedPeers(netDB *bbolt.DB) (savedAddresses []multiaddr.Multiaddr, err error) {
	//retrieve the saved addresses
	err = netDB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(peersDbKey)
		bannedP := tx.Bucket(bansDbKey)
		err = b.ForEach(func(k, v []byte) error {
			addr, err := multiaddr.NewMultiaddrBytes(k)
			if err == nil {
				peerId, err := peer.AddrInfoFromP2pAddr(addr)
				if err != nil {
					fmt.Println("peer error: " + err.Error() + ", removing from db")
					// if the saved peer cannot be validated, delete
					err = b.Delete(k)
				} else {
					byteId, err := peerId.ID.MarshalBinary()
					if err != nil {
						fmt.Println("cannot unmarshal peer: " + err.Error())
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

func GetPrivKey(netDB *bbolt.DB) (priv crypto.PrivKey, err error) {
	err = netDB.Update(func(tx *bbolt.Tx) error {
		configBucket := tx.Bucket(configBucketKey)
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
