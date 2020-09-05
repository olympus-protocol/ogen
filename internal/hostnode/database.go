package hostnode

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"go.etcd.io/bbolt"
	"path"
	"time"
)

var (
	// ErrorNotInitialize is returned when the db is not properly initialized.
	ErrorNotInitialized = errors.New("db is not initialized")

	// ErrorPeerBanned is returned when the peer trying to be stored is already banned.
	ErrorPeerBanned = errors.New("peer is already known and banned")
)

var (
	// configDBBkt is the db bucket that contains common config information
	configDBBkt = []byte("config")

	// peersDBBkt is the bucket that contains usable and known peers information
	peersDBBkt = []byte("peers")

	// privKeyDBKey is the key containing the binary serialized private key.
	privKeyDBKey = []byte("privkey")

	// bansDBBkt is the bucket that contains banned peers.
	bansDBBkt = []byte("bans")

	// ipDBBkt is the bucket that contains the serialized ip for a peer ID.
	ipDBBkt = []byte("ips")
)

type Database interface {
	SavePeer(pinfo *peer.AddrInfo) error
	BanscorePeer(pinfo *peer.AddrInfo, weight uint16) error
	GetSavedPeers() ([]*peer.AddrInfo, error)
	GetPrivKey() (priv crypto.PrivKey, err error)
}

type database struct {
	host     HostNode
	db       *bbolt.DB
	banLimit int
}

var _ Database = &database{}

// NewDatabase returns a new Database interface
func NewDatabase(dbpath string, hostnode HostNode) (Database, error) {
	netDB, err := bbolt.Open(path.Join(dbpath, "net.db"), 0600, nil)
	if err != nil {
		return nil, err
	}
	db := &database{
		host:     hostnode,
		db:       netDB,
		banLimit: 100,
	}
	err = db.load()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (d *database) load() error {
	err := d.db.Update(func(tx *bbolt.Tx) error {
		var err error
		configBucket := tx.Bucket(configDBBkt)
		if configBucket == nil {
			_, err = tx.CreateBucketIfNotExists(configDBBkt)
			if err != nil {
				return err
			}
		}
		// peersBucket holds a peerId as key and it's multiaddr as value
		peersBucket := tx.Bucket(peersDBBkt)
		if peersBucket == nil {
			_, err = tx.CreateBucketIfNotExists(peersDBBkt)
			if err != nil {
				return err
			}
		}
		// bansBucket holds an IP address as key and a timestamp as value
		bansBucket := tx.Bucket(bansDBBkt)
		if bansBucket == nil {
			_, err = tx.CreateBucketIfNotExists(bansDBBkt)
			if err != nil {
				return err
			}
		}
		// ipBucket holds a peerId as key and an IP address as value
		ipBucket := tx.Bucket(ipDBBkt)
		if ipBucket == nil {
			_, err = tx.CreateBucketIfNotExists(ipDBBkt)
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

// SavePeer stores a peer to the node hostnode database.
func (d *database) SavePeer(pinfo *peer.AddrInfo) error {

	// Get the multi-addresses of the peer as bytes
	var maBytes [][]byte

	for _, ma := range pinfo.Addrs {
		maBytes = append(maBytes, ma.Bytes())
	}

	// Check if any of this is already banned
	banned := false
	for _, ipBytes := range maBytes {
		if d.isIPBanned(ipBytes) {
			banned = true
		}
	}

	if banned {
		return ErrorPeerBanned
	}

	err := d.db.Update(func(tx *bbolt.Tx) error {
		peersBkt := tx.Bucket(peersDBBkt)
		if peersBkt == nil {
			return ErrorNotInitialized
		}

		idBytes, err := pinfo.ID.Marshal()
		if err != nil {
			return err
		}

		pinfoMarshal, err := pinfo.MarshalJSON()
		if err != nil {
			return err
		}

		err = peersBkt.Put(idBytes, pinfoMarshal)
		if err != nil {
			return err
		}

		ipsBkt := tx.Bucket(ipDBBkt)
		if ipsBkt == nil {
			return ErrorNotInitialized
		}

		for _, ipBytes := range maBytes {
			score := make([]byte, 2)
			binary.LittleEndian.PutUint16(score, 100)
			err = ipsBkt.Put(ipBytes, score)
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

// BanscorePeer reduces the banscore of a peer. If it reaches limit, it will be banned
func (d *database) BanscorePeer(pinfo *peer.AddrInfo, weight uint16) error {
	err := d.db.Update(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket(bansDBBkt)
		if bkt == nil {
			return ErrorNotInitialized
		}

		ipBkt := tx.Bucket(ipDBBkt)
		banBkt := tx.Bucket(bansDBBkt)
		for _, ma := range pinfo.Addrs {
			maBytes := ma.Bytes()
			scoreBytes := ipBkt.Get(maBytes)
			var score uint16
			if scoreBytes == nil {
				score = 100
			} else {
				score = binary.LittleEndian.Uint16(scoreBytes)
			}
			score -= weight
			if score <= 0 {

				err := d.host.DisconnectPeer(pinfo.ID)
				if err != nil {
					return err
				}

				timeToUnban := time.Now().Unix() + 86400

				timeBytes := make([]byte, 8)
				binary.LittleEndian.PutUint64(timeBytes, uint64(timeToUnban))

				err = banBkt.Put(maBytes, timeBytes)
				if err != nil {
					return err
				}

			} else {

				scoreBytes := make([]byte, 2)
				binary.LittleEndian.PutUint16(scoreBytes, score)

				err := ipBkt.Put(maBytes, scoreBytes)
				if err != nil {
					return err
				}

			}
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// isIPBanned returns true if the ip is already banned.
// It checks internally if a peer must be unbanned.
func (d *database) isIPBanned(ip []byte) bool {
	banned := false
	err := d.db.Update(func(tx *bbolt.Tx) error {

		bkt := tx.Bucket(bansDBBkt)

		// Try to fetch the ip, if it is not found, is not banned.
		banTimeBytes := bkt.Get(ip)
		if banTimeBytes == nil {
			return nil
		}

		// Convert the time to a unix timestamp
		banTime := binary.LittleEndian.Uint64(banTimeBytes)

		// Check if the peer ban already expired
		if time.Now().Unix() >= int64(banTime) {
			_ = bkt.Delete(ip)
			return nil
		} else {
			banned = true
			return nil
		}

	})

	if err != nil {
		return true
	}

	return banned
}

// GetSavedPeers returns a list of already known peers.
func (d *database) GetSavedPeers() ([]*peer.AddrInfo, error) {
	peersMap := make(map[*peer.ID]*peer.AddrInfo)

	// Fetch peers
	err := d.db.View(func(tx *bbolt.Tx) error {

		peersBkt := tx.Bucket(peersDBBkt)

		err := peersBkt.ForEach(func(k, v []byte) error {

			id := new(peer.ID)
			err := id.Unmarshal(k)
			if err != nil {
				_ = peersBkt.Delete(k)
			}

			pinfo := new(peer.AddrInfo)
			err = pinfo.UnmarshalJSON(v)
			if err != nil {
				_ = peersBkt.Delete(k)
			}

			peersMap[id] = pinfo

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})

	var peers []*peer.AddrInfo
	// Check if any of these peers are banned
	for _, pinfo := range peersMap {

		// Check if any of the multi-addresses is banned
		var maBytes [][]byte

		for _, ma := range pinfo.Addrs {
			maBytes = append(maBytes, ma.Bytes())
		}

		// Check if any of this is already banned
		banned := false
		for _, ipBytes := range maBytes {
			if d.isIPBanned(ipBytes) {
				banned = true
			}
		}

		if !banned {
			peers = append(peers, pinfo)
		}
	}

	if err != nil {
		return nil, err
	}

	return peers, nil
}

// GetPrivKey returns the private key for the HostNode
// Creates a new one if is not found, requires the database to be initialized.
func (d *database) GetPrivKey() (crypto.PrivKey, error) {

	var priv crypto.PrivKey

	err := d.db.Update(func(tx *bbolt.Tx) error {

		bkt := tx.Bucket(configDBBkt)
		if bkt == nil {
			return ErrorNotInitialized
		}

		keyBytes := bkt.Get(privKeyDBKey)

		if keyBytes == nil {

			priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
			if err != nil {
				return err
			}

			privBytes, err := crypto.MarshalPrivateKey(priv)
			if err != nil {
				return err
			}

			err = bkt.Put(privKeyDBKey, privBytes)
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
