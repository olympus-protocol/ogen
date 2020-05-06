package explorer

import (
	"html/template"
	"net/http"

	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/config"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/peers"
)

type MainInfo struct {
	Version       string
	Protocol      uint32
	UserAgent     string
	Network       string
	Sync          bool
	LastBlock     uint64
	LastBlockHash string
	LastBlockTime string
	Peers         int32
	LastBlocks    map[string]*index.BlockRow
}

type BlockInfo struct {
	Height       uint64
	Hash         string
	Transactions int
}

type BlocksInfo struct {
	Blocks map[string]BlockInfo
}

func LoadApi(conf *config.Config, chainInstance *chain.Blockchain, peersMan *peers.PeerMan) error {
	templates, err := loadTemplates()
	if err != nil {
		return err
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("explorer/static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mainInfo := loadChainStats(conf, chainInstance, peersMan)
		err := templates.ExecuteTemplate(w, "index.html", mainInfo)
		if err != nil {
			// TODO handle error
		}
	})
	http.HandleFunc("/blocks", func(w http.ResponseWriter, r *http.Request) {
		blocks := getBlocks(chainInstance)
		// TODO refactor
		blocksInfo := BlocksInfo{Blocks: blocks}
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
	tpl, err := template.New("").Funcs(tplFuncMap).ParseFiles("./explorer/views/index.html", "./explorer/views/blocks.html")
	if err != nil {
		return nil, err
	}
	return tpl, nil
}

func loadChainStats(conf *config.Config, chain *chain.Blockchain, peerMan *peers.PeerMan) MainInfo {
	lastBlock := chain.State().Tip()
	lastBlocks := map[string]*index.BlockRow{
		lastBlock.Hash.String(): lastBlock,
	}
	if lastBlock.Height != 0 && lastBlock.Parent != nil {
		currBlock := lastBlock.Parent
		for i := 0; i < 4; i++ {
			lastBlocks[currBlock.Hash.String()] = currBlock
			currBlock = currBlock.Parent
		}
	}
	info := MainInfo{
		LastBlockHash: lastBlock.Hash.String(),
		Protocol:      p2p.ProtocolVersion,
		LastBlock:     lastBlock.Height,
		Version:       config.OgenVersion(),
		UserAgent:     p2p.DefaultUserAgent,
		Sync:          true,
		Peers:         peerMan.GetPeersCount(),
		LastBlocks:    lastBlocks,
	}
	return info
}

func getBlocks(chain *chain.Blockchain) map[string]BlockInfo {
	lastBlock := chain.State().Tip()
	lastBlockInfo := BlockInfo{
		Height:       lastBlock.Height,
		Hash:         lastBlock.Hash.String(),
		Transactions: 0,
	}
	blocks := map[string]BlockInfo{
		lastBlockInfo.Hash: lastBlockInfo,
	}
	if lastBlock.Height != 0 && lastBlock.Parent != nil {
		currBlock := lastBlock.Parent
		for currBlock.Height == 0 {
			blocks[currBlock.Hash.String()] = BlockInfo{
				Height:       currBlock.Height,
				Hash:         currBlock.Hash.String(),
				Transactions: 0,
			}
			currBlock = currBlock.Parent
		}
	}
	return blocks
}
