package dashboard

import (
	"context"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/packr"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/proposer"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"html/template"
	"net/http"
)

type Dashboard struct {
	ctx      context.Context
	r        *gin.Engine
	host     hostnode.HostNode
	chain    chain.Blockchain
	proposer proposer.Proposer
}

func (d *Dashboard) Start() error {
	port := config.GlobalFlags.DashboardPort
	return d.r.Run(":" + port)
}

func (d *Dashboard) fetchData(c *gin.Context) {
	tip := d.chain.State().Tip()
	justified, _ := d.chain.State().GetJustifiedHead()
	finalized, _ := d.chain.State().GetFinalizedHead()
	peers := d.host.GetPeersStats()
	peersAhead := 0
	peersBehind := 0
	peersEqual := 0

	var peersData []PeerData
	for _, p := range peers {
		if p.FinalizedHeight > finalized.Height {
			peersAhead += 1
		}
		if p.FinalizedHeight == finalized.Height {
			peersEqual += 1
		}
		if p.FinalizedHeight < finalized.Height {
			peersBehind += 1
		}
		pData := PeerData{
			ID:        p.ID.String(),
			Finalized: p.FinalizedHeight,
			Justified: p.JustifiedHeight,
			Tip:       p.TipHeight,
		}
		peersData = append(peersData, pData)
	}

	validators := d.chain.State().TipState().GetValidatorRegistry()
	keys, err := d.proposer.Keystore().GetValidatorKeys()
	if err != nil {
		c.HTML(500, "", nil)
		return
	}

	activeValidators := 0
	keysActive := 0
	for _, v := range validators {
		if v.Status == primitives.StatusActive {
			activeValidators += 1
		}
		if _, ok := d.proposer.Keystore().GetValidatorKey(v.PubKey); ok {
			keysActive += 1
		}
	}

	slotEpoch := d.proposer.GetCurrentSlot() % config.GlobalParams.NetParams.EpochLength
	slotEpoch += 1
	data := Data{
		NodeData: NodeData{
			TipHeight:       tip.Height,
			TipSlot:         tip.Slot,
			TipHash:         hex.EncodeToString(tip.Hash[:]),
			JustifiedHeight: justified.Height,
			JustifiedSlot:   justified.Slot,
			JustifiedHash:   hex.EncodeToString(justified.Hash[:]),
			FinalizedHeight: finalized.Height,
			FinalizedSlot:   finalized.Slot,
			FinalizedHash:   hex.EncodeToString(finalized.Hash[:]),
		},
		NetworkData: NetworkData{
			ID:             d.host.GetHost().ID().String(),
			PeersConnected: len(d.host.GetHost().Network().Peers()),
			PeersAhead:     peersAhead,
			PeersBehind:    peersBehind,
			PeersEqual:     peersEqual,
		},
		KeystoreData: KeystoreData{
			Keys:              len(keys),
			Validators:        activeValidators,
			KeysParticipating: keysActive,
		},
		ProposerData: ProposerData{
			Slot:      d.proposer.GetCurrentSlot(),
			Epoch:     d.proposer.GetCurrentSlot() / config.GlobalParams.NetParams.EpochLength,
			Voting:    d.proposer.Voting(),
			Proposing: d.proposer.Proposing(),
		},
		PeerData: peersData,
		ParticipationInfo: ParticipationInfo{
			EpochSlot:               slotEpoch,
			Epoch:                   d.proposer.GetCurrentSlot() / config.GlobalParams.NetParams.EpochLength,
			ParticipationPercentage: "",
		},
	}
	c.HTML(http.StatusOK, "index.html", data)
	return
}

func (d *Dashboard) loadStatic() {
	box := packr.NewBox("./static/")

	d.r.StaticFS("/static/", box)
}

func (d *Dashboard) loadTemplate() error {

	box := packr.NewBox("./views/")

	t := template.New("")

	tmpl := t.New("index.html")

	data, err := box.FindString("index.html")
	if err != nil {
		return err
	}

	pTmpl, err := tmpl.Parse(data)
	if err != nil {
		return err
	}

	d.r.SetHTMLTemplate(pTmpl)

	return nil
}

func NewDashboard(h hostnode.HostNode, ch chain.Blockchain, prop proposer.Proposer) (*Dashboard, error) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	d := &Dashboard{
		ctx:      config.GlobalParams.Context,
		r:        r,
		host:     h,
		chain:    ch,
		proposer: prop,
	}
	err := d.loadTemplate()
	if err != nil {
		return nil, err
	}

	d.loadStatic()

	r.GET("/", d.fetchData)

	return d, nil
}
