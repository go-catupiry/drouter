package drouter

import (
	"github.com/go-catupiry/catu"
	"github.com/gookit/event"
	"github.com/sirupsen/logrus"
)

type DRouterPlugin struct {
	catu.Pluginer
	Name string

	UrlAliasController *UrlAliasController
}

func (r *DRouterPlugin) GetName() string {
	return r.Name
}

func (r *DRouterPlugin) Init(app catu.App) error {
	logrus.Debug(r.GetName() + " Init")

	app.GetEvents().On("bindRoutes", event.ListenerFunc(func(e event.Event) error {
		return r.BindRoutes(app)
	}), event.Normal)

	return nil
}

func (r *DRouterPlugin) BindRoutes(app catu.App) error {
	logrus.Debug(r.GetName() + " BindRoutes")

	ctl := r.UrlAliasController

	router := app.GetRouter()
	router.Pre(urlAliasMiddleware())

	routerApi := app.SetRouterGroup("url-alia-api", "/api/url-alia")
	app.SetResource("url-alia", ctl, routerApi)
	return nil
}

func (r *DRouterPlugin) SetTemplateFuncMap(app catu.App) error {
	return nil
}

type PluginCfgs struct{}

func NewPlugin(cfg *PluginCfgs) *DRouterPlugin {
	p := DRouterPlugin{Name: "droute"}
	return &p
}
