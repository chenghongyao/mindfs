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
	root := flag.String("root", "", "root directory to manage")
	flag.Parse()

	if *root == "" {
		fmt.Fprintln(os.Stderr, "root directory required")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := app.Start(ctx, *addr, *root); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
