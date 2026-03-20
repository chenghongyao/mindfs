package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"mindfs/server/app"
)

func main() {
	addr := flag.String("addr", ":7331", "listen address")
	staticDir := flag.String("static-dir", "web/dist", "directory for serving built web assets")
	bindCode := flag.String("bind-code", "", "relay bind code for activation/binding")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := app.Start(ctx, *addr, app.StartOptions{
		StaticDir: *staticDir,
		BindCode:  *bindCode,
	}); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
