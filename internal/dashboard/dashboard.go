package dashboard

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/packr"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"html/template"
	"net/http"
)

type Dashboard struct {
	ctx   context.Context
	r     *gin.Engine
	host  hostnode.HostNode
	chain chain.Blockchain
}

func (d *Dashboard) Start() error {
	port := config.GlobalFlags.DashboardPort
	return d.r.Run(":" + port)
}

func (d *Dashboard) fetchData(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
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

func NewDashboard(h hostnode.HostNode, ch chain.Blockchain) (*Dashboard, error) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	d := &Dashboard{
		ctx:   config.GlobalParams.Context,
		r:     r,
		host:  h,
		chain: ch,
	}
	err := d.loadTemplate()
	if err != nil {
		return nil, err
	}

	d.loadStatic()

	r.GET("/", d.fetchData)

	return d, nil
}
