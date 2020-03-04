package explorer

import (
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/config"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/peers"
	"github.com/olympus-protocol/ogen/primitives"
	"html/template"
	"net/http"
)

type MainInfo struct {
	Version       string
	UserAgent     string
	Protocol      uint32
	Mode          string
	Network       string
	Sync          bool
	LastBlock     int32
	LastBlockHash string
	LastBlockTime string
	Size          string
	MempoolTxs    int
	Masternodes   int
	Peers         int32
	Proposals     int
	Votes         int
	LastBlocks    map[string]*primitives.Block
}

type BlocksInfo struct {
	Blocks []chain.BlockInfo
}

func LoadApi(conf *config.Config, chainInstance *chain.Blockchain, peersMan *peers.PeerMan) error {
	templates, err := loadTemplates()
	if err != nil {
		return err
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("explorer/static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mainInfo := loadChainStats(conf, chainInstance, peersMan)
		err := templates.ExecuteTemplate(w, "dbindex.html", mainInfo)
		if err != nil {
			// TODO handle error
		}
	})
	http.HandleFunc("/blocks", func(w http.ResponseWriter, r *http.Request) {
		// TODO refactor
		blocksInfo := BlocksInfo{}
		err := templates.ExecuteTemplate(w, "blocks.html", blocksInfo)
		if err != nil {
			// TODO handle error
		}
	})
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		return err
	}
	return nil
}

func loadTemplates() (*template.Template, error) {
	tplFuncMap := make(template.FuncMap)
	tpl, err := template.New("").Funcs(tplFuncMap).ParseFiles("./explorer/views/dbindex.html", "./explorer/views/blocks.html")
	if err != nil {
		return nil, err
	}
	return tpl, nil
}

func loadChainStats(conf *config.Config, chain *chain.Blockchain, peerMan *peers.PeerMan) MainInfo {
	// TODO ref.
	return MainInfo{
		Version:     config.OgenVersion(),
		UserAgent:   p2p.DefaultUserAgent,
		Protocol:    p2p.ProtocolVersion,
		Mode:        conf.Mode,
		Sync:        false,
		MempoolTxs:  0,
		Masternodes: 0,
		Peers:       peerMan.GetPeersCount(),
		Proposals:   0,
		Votes:       0,
	}
}
