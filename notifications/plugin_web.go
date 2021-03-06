package notifications

import (
	"fmt"
	"github.com/jonas747/discordgo"
	"github.com/jonas747/yagpdb/common"
	"github.com/jonas747/yagpdb/common/configstore"
	"github.com/jonas747/yagpdb/web"
	"goji.io/pat"
	"html/template"
	"net/http"
)

func (p *Plugin) InitWeb() {
	tmplPath := "templates/plugins/notifications_general.html"
	if common.Testing {
		tmplPath = "../../notifications/assets/notifications_general.html"
	}

	web.Templates = template.Must(web.Templates.ParseFiles(tmplPath))

	getHandler := web.RenderHandler(HandleNotificationsGet, "cp_notifications_general")
	postHandler := web.ControllerPostHandler(HandleNotificationsPost, getHandler, Config{}, "Updated general notifications config.")

	web.CPMux.Handle(pat.Get("/notifications/general"), web.RequireGuildChannelsMiddleware(getHandler))
	web.CPMux.Handle(pat.Get("/notifications/general/"), web.RequireGuildChannelsMiddleware(getHandler))

	web.CPMux.Handle(pat.Post("/notifications/general"), web.RequireGuildChannelsMiddleware(postHandler))
	web.CPMux.Handle(pat.Post("/notifications/general/"), web.RequireGuildChannelsMiddleware(postHandler))
}

func HandleNotificationsGet(w http.ResponseWriter, r *http.Request) interface{} {
	ctx := r.Context()
	activeGuild, templateData := web.GetBaseCPContextData(ctx)

	formConfig, ok := ctx.Value(common.ContextKeyParsedForm).(*Config)
	if ok {
		templateData["NotifyConfig"] = formConfig
	} else {
		conf, err := GetConfig(activeGuild.ID)
		if err != nil {
			web.CtxLogger(r.Context()).WithError(err).Error("failed retrieving config")
		}

		templateData["NotifyConfig"] = conf
	}

	return templateData
}

func HandleNotificationsPost(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	ctx := r.Context()
	activeGuild, templateData := web.GetBaseCPContextData(ctx)
	templateData["VisibleURL"] = "/manage/" + discordgo.StrID(activeGuild.ID) + "/notifications/general/"

	newConfig := ctx.Value(common.ContextKeyParsedForm).(*Config)

	newConfig.GuildID = activeGuild.ID

	err := configstore.SQL.SetGuildConfig(ctx, newConfig)
	if err != nil {
		return templateData, nil
	}

	return templateData, nil
}

var _ web.PluginWithServerHomeWidget = (*Plugin)(nil)

func (p *Plugin) LoadServerHomeWidget(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	ag, templateData := web.GetBaseCPContextData(r.Context())

	templateData["WidgetTitle"] = "General notifications"
	templateData["SettingsPath"] = "/notifications/general"

	config, err := GetConfig(ag.ID)
	if err != nil {
		return templateData, err
	}

	format := `<ul>
	<li>Join Server message: %s</li>
	<li>Join DM message: %s</li>
	<li>Leave message: %s</li>
	<li>Topic change message: %s</li>
</ul>`

	if config.JoinServerEnabled || config.JoinDMEnabled || config.LeaveEnabled || config.TopicEnabled {
		templateData["WidgetEnabled"] = true
	} else {
		templateData["WidgetDisabled"] = true
	}

	templateData["WidgetBody"] = template.HTML(fmt.Sprintf(format,
		web.EnabledDisabledSpanStatus(config.JoinServerEnabled), web.EnabledDisabledSpanStatus(config.JoinDMEnabled),
		web.EnabledDisabledSpanStatus(config.LeaveEnabled), web.EnabledDisabledSpanStatus(config.TopicEnabled)))

	return templateData, nil
}

func enabledDisabled(b bool) string {
	if b {
		return "enabled"
	}

	return "disabled"
}
