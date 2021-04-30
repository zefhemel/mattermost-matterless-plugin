package main

import (
	"fmt"
	"github.com/mattermost/mattermost-server/v5/shared/mlog"
	"github.com/zefhemel/matterless/pkg/application"
	"github.com/zefhemel/matterless/pkg/config"
	"github.com/zefhemel/matterless/pkg/util"
	"io"
	"net/http"
	"sync"

	"github.com/mattermost/mattermost-server/v5/plugin"

	log "github.com/sirupsen/logrus"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	// Matterless stuff
	config    *config.Config
	container *application.Container
}

func (p *Plugin) OnActivate() error {
	log.SetLevel(log.DebugLevel)
	cfg := config.NewConfig()
	pluginConfig := p.getConfiguration()
	cfg.DataDir = pluginConfig.DataDir
	cfg.APIBindPort = util.FindFreePort(8000)
	cfg.AdminToken = util.TokenGenerator()
	cfg.PersistApps = true
	mlog.Info(fmt.Sprintf("All config: %+v", cfg))

	var err error
	p.container, err = application.NewContainer(cfg)
	p.config = cfg
	if err != nil {
		return err
	}

	// Subscribe to all logs and write to stdout
	p.container.EventBus().Subscribe("logs:*", func(eventName string, eventData interface{}) {
		if le, ok := eventData.(application.LogEntry); ok {
			mlog.Info(fmt.Sprintf("[%s | %s] %s", le.AppName, le.LogEntry.FunctionName, le.LogEntry.Message))
		}
	})

	// Load previously deployed apps from disk
	if err := p.container.LoadAppsFromDisk(); err != nil {
		mlog.Error("Could not load apps from disk: %s", mlog.Err(err))
	}

	return nil
}

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-Id")
	if userID != "" {
		user, err := p.API.GetUser(userID)
		if err != nil {
			mlog.Error("Error in authenticated user lookup", mlog.Err(err))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if user.IsSystemAdmin() {
			r.Header.Set("Authorization", fmt.Sprintf("bearer %s", p.config.AdminToken))
		}
	}
	// mlog.Info(fmt.Sprintf("Got HTTP request: %s: %s Headers: %+v", r.Method, r.URL, r.Header))

	// Proxy request
	p.proxy(w, r)
}

func (p *Plugin) proxy(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	req, err := http.NewRequest(r.Method, fmt.Sprintf("http://localhost:%d%s", p.config.APIBindPort, r.URL), r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Proxy error: %s", err), http.StatusInternalServerError)
		return
	}
	req.Header = r.Header
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Proxy error: %s", err), http.StatusInternalServerError)
		return
	}
	for k, vs := range res.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(res.StatusCode)
	_, err = io.Copy(w, res.Body)
	if err != nil {
		mlog.Error("Error proxying", mlog.Err(err))
	}
	res.Body.Close()
}
