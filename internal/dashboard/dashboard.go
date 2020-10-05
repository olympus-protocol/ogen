package dashboard

import (
	"github.com/gin-gonic/gin"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
)

type dashboard struct {
	server *gin.Engine
	host   hostnode.HostNode
	chain  chain.Blockchain
}

func (d *dashboard) Start() error {
	return d.server.Run(":3000")
}

func NewDashboard(h hostnode.HostNode, ch chain.Blockchain) *dashboard {
	return &dashboard{
		server: gin.Default(),
		host:   h,
		chain:  ch,
	}
}
