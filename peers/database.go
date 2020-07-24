package peers

import (
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"go.etcd.io/bbolt"
	"strconv"
	"time"
)

// contains several functions that interact with netDB database

var configBucketKey = []byte("config")
var privKeyDbKey = []byte("privkey")

var peersDbKey = []byte("peers")
var bansDbKey = []byte("bans")
var ipDbKey = []byte("ips")
var scoresDbKey = []byte("scores")

const BanLimit = 100

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
		// peersBucket holds a peerId as key and it's multiaddr as value
		peersBucket := tx.Bucket(peersDbKey)
		if peersBucket == nil {
			peersBucket, err = tx.CreateBucketIfNotExists(peersDbKey)
			if err != nil {
				return err
			}
		}
		// bansBucket holds an IP address as key and a timestamp as value
		bansBucket := tx.Bucket(bansDbKey)
		if bansBucket == nil {
			bansBucket, err = tx.CreateBucketIfNotExists(bansDbKey)
			if err != nil {
				return err
			}
		}
		// ipBucket holds a peerId as key and an IP address as value
		ipBucket := tx.Bucket(ipDbKey)
		if ipBucket == nil {
			ipBucket, err = tx.CreateBucketIfNotExists(ipDbKey)
			if err != nil {
				return err
			}
		}
		// scoresBucket holds a multiaddress  as key and the banscore as value
		scoreBucket := tx.Bucket(scoresDbKey)
		if scoreBucket == nil {
			scoreBucket, err = tx.CreateBucketIfNotExists(scoresDbKey)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return
}

func SavePeer(netDB *bbolt.DB, pma multiaddr.Multiaddr) error {
	// get peerId from multiaddr
	peerId, err := peer.AddrInfoFromP2pAddr(pma)
	if err != nil {
		return err
	}
	// extract ip from multiaddr
	ip, err := extractIp(pma)
	if err != nil {
		return err
	}
	// check if ip is banned
	isBanned, shoulDelete, err := IsIpBanned(netDB, ip)
	if err != nil {
		return err
	}
	if !isBanned {
		err = netDB.Update(func(tx *bbolt.Tx) error {
			var err error
			peersBucket := tx.Bucket(peersDbKey)
			ipBucket := tx.Bucket(ipDbKey)
			scoreBucket := tx.Bucket(scoresDbKey)
			bansDb := tx.Bucket(bansDbKey)
			if shoulDelete {
				err = bansDb.Delete([]byte(ip))
			}
			if err != nil {
				return err
			}
			byteId, err := peerId.ID.MarshalBinary()
			if err != nil {
				return err
			}
			// save ip from peer
			err = ipBucket.Put(byteId, []byte(ip))
			// create banscore for peer if it does not have one
			if scoreBucket.Get(pma.Bytes()) == nil {
				err = scoreBucket.Put(pma.Bytes(), []byte(strconv.Itoa(0)))
			}
			// save multiaddr of peerId
			err = peersBucket.Put(byteId, pma.Bytes())
			return err
		})
	} else {
		err = fmt.Errorf("peer %s is banned", peerId.ID.String())
	}
	return err

}

// Reduces the banscore of a peer. If it reaches limit, it will be banned
func BanscorePeer(netDB *bbolt.DB, id peer.ID, weight int) error {
	err := netDB.Update(func(tx *bbolt.Tx) error {
		var err error
		savedDb := tx.Bucket(peersDbKey)
		bansDb := tx.Bucket(bansDbKey)
		ipb := tx.Bucket(ipDbKey)
		scoreDb := tx.Bucket(scoresDbKey)
		//get multiaddr from peerId
		byteId, err := id.MarshalBinary()
		if err != nil {
			return err
		}
		multiAddrBytes := savedDb.Get(byteId)
		if multiAddrBytes == nil {
			return errors.New("could not find peer")
		}
		//get banscore from multiaddr
		scoreBytes := scoreDb.Get(multiAddrBytes)
		if scoreBytes == nil {
			return errors.New("could not find peer score")
		}
		score, err := strconv.Atoi(string(scoreBytes))
		if err != nil {
			return err
		}
		fmt.Printf("peer %s has a banscore of: %s \n", id.String(), strconv.Itoa(score))
		score += weight
		if score >= BanLimit {
			// add to banlist
			ipBytes := ipb.Get(byteId)
			if ipBytes == nil {
				return errors.New("could not find peer ip")
			}
			timestamp := strconv.FormatInt(time.Now().Unix()+24*3600, 10)
			err = bansDb.Put(ipBytes, []byte(timestamp))
			if err != nil {
				return err
			}
			fmt.Printf("%s is banned until %s \n", string(ipBytes), timestamp)
			// remove from saved list and score list
			err = savedDb.Delete(byteId)
			err = scoreDb.Delete(multiAddrBytes)
		} else {
			//update banscore
			err = scoreDb.Put(multiAddrBytes, []byte(strconv.Itoa(score)))
		}
		return err
	})
	return err
}

func IsPeerBanned(netDB *bbolt.DB, id peer.ID) (bool, error) {
	var savedIp []byte
	err := netDB.View(func(tx *bbolt.Tx) error {
		var err error
		ipb := tx.Bucket(ipDbKey)
		byteId, err := id.MarshalBinary()
		if err != nil {
			return err
		}
		savedIp = ipb.Get(byteId)
		return nil
	})
	if savedIp == nil {
		return true, nil
	}
	isBanned, shoudDelete, err := IsIpBanned(netDB, string(savedIp))
	if shoudDelete {
		err = netDB.Update(func(tx *bbolt.Tx) error {
			var err error
			banDb := tx.Bucket(bansDbKey)
			err = banDb.Delete(savedIp)
			return err
		})
	}
	return isBanned, err
}

func IsIpBanned(netDB *bbolt.DB, ip string) (bool, bool, error) {
	isBanned := false
	shouldDelete := false
	err := netDB.View(func(tx *bbolt.Tx) error {
		bansB := tx.Bucket(bansDbKey)
		ipBytes := []byte(ip)
		bannedTime := bansB.Get(ipBytes)
		if bannedTime != nil {
			timeBan, err := strconv.ParseInt(string(bannedTime), 10, 64)
			if err != nil {
				return err
			}
			if timeBan <= time.Now().Unix() {
				// if time has passed, unban
				shouldDelete = true
			} else {
				isBanned = true
			}
			return err
		}
		return nil
	})
	return isBanned, shouldDelete, err
}

func GetSavedPeers(netDB *bbolt.DB) (savedAddresses []multiaddr.Multiaddr, err error) {
	//retrieve the saved addresses
	err = netDB.Update(func(tx *bbolt.Tx) error {
		savedBucket := tx.Bucket(peersDbKey)
		err = savedBucket.ForEach(func(k, v []byte) error {
			addr, err := multiaddr.NewMultiaddrBytes(v)
			if err == nil {
				peerId, err := peer.AddrInfoFromP2pAddr(addr)
				if err != nil {
					fmt.Println("peer error: " + err.Error() + ", removing from db")
					// if the saved peer cannot be validated, delete
					err = savedBucket.Delete(k)
				} else {
					isBanned, err := IsPeerBanned(netDB, peerId.ID)
					if !isBanned && err == nil {
						savedAddresses = append(savedAddresses, addr)
					}
					// if saved peer is banned, delete
					if isBanned && err == nil {
						err = savedBucket.Delete(k)
					}
				}
			}
			return err
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

func extractIp(pma multiaddr.Multiaddr) (ip string, err error) {
	protocols := pma.Protocols()
	for _, s := range protocols {
		if s.Name == "ip4" || s.Name == "ip6" {
			value, err := pma.ValueForProtocol(s.Code)
			if err != nil {
				return "", errors.New("could not get ip from multiaddr")
			}
			ip = s.Name + "/" + value
			break
		}
	}
	if ip == "" {
		err = errors.New("no ip found in multiaddr " + pma.String())
	}
	return
}
