package app

import (
	"context"
	"net/http"
	"time"

	"mindfs/server/internal/agent"
	"mindfs/server/internal/api"
	"mindfs/server/internal/fs"
	"mindfs/server/internal/router"
	"mindfs/server/internal/session"
)

// Start boots the HTTP/WS server for a managed directory.
func Start(ctx context.Context, addr, root string) error {
	managedDir, err := fs.EnsureManagedDir(root)
	if err != nil {
		return err
	}
	registry, err := fs.NewDefaultRegistry()
	if err != nil {
		return err
	}
	_ = registry.Load()
	_, _ = registry.Upsert(root)
	viewStores := router.NewViewStoreManager()

	actionRouter := router.New()
	actionRouter.SetFallback(router.DefaultHandler(func() ([]fs.Entry, error) {
		return fs.ListEntries(root, root)
	}))
	actionService := &router.ActionService{
		Root:       root,
		ManagedDir: managedDir,
		Router:     actionRouter,
		Registry:   registry,
	}
	if err := actionService.RegisterDefaults(); err != nil {
		return err
	}

	sessionStores := session.NewStoreManager()
	sessionService := &api.SessionService{
		Stores:   sessionStores,
		Root:     root,
		Registry: registry,
	}
	agentConfig, _ := agent.LoadConfig("")
	agentPool := agent.NewPool(agentConfig)
	agentProber := agent.NewProber(&agentConfig, 5*time.Minute)
	agentProber.Start(ctx)
	idleChecker := session.NewIdleChecker(sessionStores, 1*time.Minute, 10*time.Minute, 30*time.Minute)
	idleChecker.Start(ctx)

	httpHandler := &api.HTTPHandler{Router: actionRouter, Root: root, Views: viewStores, Registry: registry, Sessions: sessionService, Prober: agentProber}
	wsHandler := &api.WSHandler{Router: actionRouter, Root: root, Registry: registry, Sessions: sessionService, Agents: agentPool}

	mux := http.NewServeMux()
	mux.Handle("/", httpHandler.Routes())
	mux.Handle("/ws", wsHandler)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		idleChecker.Stop()
		agentProber.Stop()
		agentPool.CloseAll()
		_ = server.Shutdown(context.Background())
	}()

	return server.ListenAndServe()
}
