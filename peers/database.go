package peers

import (
	"crypto/rand"
	"errors"
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

// BanLimit is the maximum ban score for a peer to get banned.
const BanLimit = 100

// InitBuckets initializes the peer database buckets
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
			_, err = tx.CreateBucketIfNotExists(peersDbKey)
			if err != nil {
				return err
			}
		}
		// bansBucket holds an IP address as key and a timestamp as value
		bansBucket := tx.Bucket(bansDbKey)
		if bansBucket == nil {
			_, err = tx.CreateBucketIfNotExists(bansDbKey)
			if err != nil {
				return err
			}
		}
		// ipBucket holds a peerId as key and an IP address as value
		ipBucket := tx.Bucket(ipDbKey)
		if ipBucket == nil {
			_, err = tx.CreateBucketIfNotExists(ipDbKey)
			if err != nil {
				return err
			}
		}
		// scoresBucket holds a multiaddress  as key and the banscore as value
		scoreBucket := tx.Bucket(scoresDbKey)
		if scoreBucket == nil {
			_, err = tx.CreateBucketIfNotExists(scoresDbKey)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return
}

// SavePeer stores a peer to the node peers database.
func SavePeer(netDB *bbolt.DB, pma multiaddr.Multiaddr) error {

	// get peerID from multiaddr
	peerID, err := peer.AddrInfoFromP2pAddr(pma)
	if err != nil {
		return err
	}
	// extract ip from multiaddr
	ip, err := extractIP(pma)
	if err != nil {
		return err
	}
	// check if ip is banned
	isBanned, shoulDelete, err := IsIPBanned(netDB, ip)
	if err != nil {
		return err
	}
	if !isBanned {
		err = netDB.Update(func(tx *bbolt.Tx) error {
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
			byteID, err := peerID.ID.MarshalBinary()
			if err != nil {
				return err
			}
			// save ip from peer
			err = ipBucket.Put(byteID, []byte(ip))
			if err != nil {
				return err
			}
			// create banscore for peer if it does not have one
			if scoreBucket.Get(pma.Bytes()) == nil {
				err = scoreBucket.Put(pma.Bytes(), []byte(strconv.Itoa(0)))
				if err != nil {
					return err
				}
			}
			// save multiaddr of peerId
			err = peersBucket.Put(byteID, pma.Bytes())
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	} 
	return nil

}

// BanscorePeer reduces the banscore of a peer. If it reaches limit, it will be banned
func BanscorePeer(netDB *bbolt.DB, id peer.ID, weight int) error {
	err := netDB.Update(func(tx *bbolt.Tx) error {
		var err error
		savedDb := tx.Bucket(peersDbKey)
		bansDb := tx.Bucket(bansDbKey)
		ipb := tx.Bucket(ipDbKey)
		scoreDb := tx.Bucket(scoresDbKey)
		//get multiaddr from peerId
		byteID, err := id.MarshalBinary()
		if err != nil {
			return err
		}
		multiAddrBytes := savedDb.Get(byteID)
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
		score += weight
		if score >= BanLimit {
			// add to banlist
			ipBytes := ipb.Get(byteID)
			if ipBytes == nil {
				return errors.New("could not find peer ip")
			}
			timestamp := strconv.FormatInt(time.Now().Unix()+24*3600, 10)
			err = bansDb.Put(ipBytes, []byte(timestamp))
			if err != nil {
				return err
			}
			// remove from saved list and score list
			err = savedDb.Delete(byteID)
			if err != nil {
				return err
			}
			err = scoreDb.Delete(multiAddrBytes)
			if err != nil {
				return err
			}
		} else {
			//update banscore
			err = scoreDb.Put(multiAddrBytes, []byte(strconv.Itoa(score)))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// IsPeerBanned returns a boolean if a peer is already known and banned.
func IsPeerBanned(netDB *bbolt.DB, id peer.ID) (bool, error) {
	var savedIP []byte
	err := netDB.View(func(tx *bbolt.Tx) error {
		var err error
		ipb := tx.Bucket(ipDbKey)
		byteID, err := id.MarshalBinary()
		if err != nil {
			return err
		}
		savedIP = ipb.Get(byteID)
		return nil
	})
	if err != nil {
		return false, err
	}
	if savedIP == nil {
		return true, nil
	}
	isBanned, shoudDelete, err := IsIPBanned(netDB, string(savedIP))
	if shoudDelete {
		err = netDB.Update(func(tx *bbolt.Tx) error {
			var err error
			banDb := tx.Bucket(bansDbKey)
			err = banDb.Delete(savedIP)
			return err
		})
	}
	return isBanned, err
}

// IsIPBanned returns booleans if a peer is alreday known and banned.
func IsIPBanned(netDB *bbolt.DB, ip string) (bool, bool, error) {
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

// GetSavedPeers returns a list of already known peers.
func GetSavedPeers(netDB *bbolt.DB) (savedAddresses []multiaddr.Multiaddr, err error) {
	// retrieve the saved addresses
	err = netDB.Update(func(tx *bbolt.Tx) error {
		savedBucket := tx.Bucket(peersDbKey)
		err = savedBucket.ForEach(func(k, v []byte) error {
			addr, err := multiaddr.NewMultiaddrBytes(v)
			if err == nil {
				peerID, err := peer.AddrInfoFromP2pAddr(addr)
				if err != nil {
					// if the saved peer cannot be validated, delete
					savedBucket.Delete(k)
				} else {
					isBanned, err := IsPeerBanned(netDB, peerID.ID)
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

// GetPrivKey returns the private key for the hostnode.
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

func extractIP(pma multiaddr.Multiaddr) (ip string, err error) {
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
