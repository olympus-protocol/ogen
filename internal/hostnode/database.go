package hostnode

import (
	"crypto/rand"
	"errors"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"go.etcd.io/bbolt"
	"path"
	"strconv"
	"time"
)

// contains several functions that interact with netDB database

var configBucketKey = []byte("config")
var privKeyDbKey = []byte("privkey")

var peersDbKey = []byte("hostnode")
var bansDbKey = []byte("bans")
var ipDbKey = []byte("ips")
var scoresDbKey = []byte("scores")

type Database interface {
	Initialize() (err error)
	SavePeer(pma multiaddr.Multiaddr) error
	BanscorePeer(id peer.ID, weight int) (bool, error)
	IsPeerBanned(id peer.ID) (bool, error)
	IsIPBanned(ip string) (bool, bool, error)
	GetSavedPeers() (savedAddresses []multiaddr.Multiaddr, err error)
	GetPrivKey() (priv crypto.PrivKey, err error)
}

type database struct {
	db       *bbolt.DB
	BanLimit int
}

var _ Database = &database{}

// NewDatabase returns a new Database interface
func NewDatabase(dbpath string) (Database, error) {
	netDB, err := bbolt.Open(path.Join(dbpath, "net.db"), 0600, nil)
	if err != nil {
		return nil, err
	}
	return &database{
		db:       netDB,
		BanLimit: 100,
	}, nil
}

// InitBuckets initializes the peer database buckets
func (d *database) Initialize() (err error) {
	err = d.db.Update(func(tx *bbolt.Tx) error {
		var err error
		configBucket := tx.Bucket(configBucketKey)
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

// SavePeer stores a peer to the node hostnode database.
func (d *database) SavePeer(pma multiaddr.Multiaddr) error {

	// get peerID from multiaddr
	peerID, err := peer.AddrInfoFromP2pAddr(pma)
	if err != nil {
		return err
	}

	// extract ip from multiaddr
	ip, err := d.extractIP(pma)
	if err != nil {
		return err
	}

	// check if ip is banned
	ban, del, err := d.IsIPBanned(ip)
	if err != nil {
		return err
	}

	if ban {
		return nil
	}

	err = d.db.Update(func(tx *bbolt.Tx) error {

		peersBucket := tx.Bucket(peersDbKey)
		ipBucket := tx.Bucket(ipDbKey)
		scoreBucket := tx.Bucket(scoresDbKey)
		bansDb := tx.Bucket(bansDbKey)

		if del {
			err = bansDb.Delete([]byte(ip))
			if err != nil {
				return err
			}
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

	return nil

}

// BanscorePeer reduces the banscore of a peer. If it reaches limit, it will be banned
func (d *database) BanscorePeer(id peer.ID, weight int) (bool, error) {

	shouldBan := false

	err := d.db.Update(func(tx *bbolt.Tx) error {

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

		if score >= d.BanLimit {

			shouldBan = true
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
		return false, err
	}
	return shouldBan, nil
}

// IsPeerBanned returns a boolean if a peer is already known and banned.
func (d *database) IsPeerBanned(id peer.ID) (bool, error) {
	var savedIP []byte
	err := d.db.View(func(tx *bbolt.Tx) error {
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
	isBanned, shoudDelete, err := d.IsIPBanned(string(savedIP))
	if shoudDelete {
		err = d.db.Update(func(tx *bbolt.Tx) error {
			var err error
			banDb := tx.Bucket(bansDbKey)
			err = banDb.Delete(savedIP)
			return err
		})
	}
	return isBanned, err
}

// IsIPBanned returns booleans if a peer is already known and banned, this function will also return
// the second boolean if the ip needs to be unbanned.
func (d *database) IsIPBanned(ip string) (bool, bool, error) {

	banned := false
	del := false

	err := d.db.View(func(tx *bbolt.Tx) error {

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
				del = true

			} else {

				banned = true
			}

			return err
		}
		return nil
	})
	if err != nil {
		return true, true, err
	}
	return banned, del, nil
}

// GetSavedPeers returns a list of already known hostnode.
func (d *database) GetSavedPeers() ([]multiaddr.Multiaddr, error) {
	var savedPeers []multiaddr.Multiaddr
	err := d.db.Update(func(tx *bbolt.Tx) error {

		savedBucket := tx.Bucket(peersDbKey)

		err := savedBucket.ForEach(func(k, v []byte) error {

			addr, err := multiaddr.NewMultiaddrBytes(v)
			if err != nil {
				return err
			} else {

				peerID, err := peer.AddrInfoFromP2pAddr(addr)
				if err != nil {

					// if the saved peer cannot be validated, delete
					_ = savedBucket.Delete(k)

				} else {

					isBanned, err := d.IsPeerBanned(peerID.ID)

					if !isBanned && err == nil {
						savedPeers = append(savedPeers, addr)
					}

					// if saved peer is banned, delete
					if isBanned && err == nil {
						err = savedBucket.Delete(k)
						if err != nil {
							return err
						}
					}
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return savedPeers, nil
}

// GetPrivKey returns the private key for the HostNode.
func (d *database) GetPrivKey() (crypto.PrivKey, error) {
	var priv crypto.PrivKey

	err := d.db.Update(func(tx *bbolt.Tx) error {

		configBucket := tx.Bucket(configBucketKey)

		keyBytes := configBucket.Get(privKeyDbKey)

		if keyBytes == nil {

			priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
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
	if err != nil {
		return nil, err
	}
	return priv, nil
}

func (d *database) extractIP(pma multiaddr.Multiaddr) (ip string, err error) {
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
