//go:generate go run types/codegen/cleanup/main.go
//go:generate go run types/codegen/main.go

package main

import (
	"context"
	"net/http"
	"os"

	"github.com/rancher/norman"
	"github.com/rancher/norman/pkg/resolvehome"
	"github.com/rancher/norman/signal"
	"github.com/rancher/webhookinator/pkg/server"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	VERSION = "v0.0.0-dev"
)

func main() {
	app := cli.NewApp()
	app.Name = "webhookinator"
	app.Version = VERSION
	app.Usage = "webhookinator needs help!"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "kubeconfig",
			EnvVar: "KUBECONFIG",
			Value:  "${HOME}/.kube/config",
		},
		cli.StringFlag{
			Name:  "listen-address",
			Value: ":8888",
		},
	}
	app.Action = run

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	logrus.Info("Starting controller")
	ctx := signal.SigTermCancelContext(context.Background())

	kubeConfig, err := resolvehome.Resolve(c.String("kubeconfig"))
	if err != nil {
		return err
	}

	ctx, srv, err := server.Config().Build(ctx, &norman.Options{
		K8sMode:    "external",
		KubeConfig: kubeConfig,
	})
	if err != nil {
		return err
	}

	addr := c.String("listen-address")
	logrus.Infof("Listening on %s", addr)

	go func() {
		if err := http.ListenAndServe(addr, srv.APIHandler); err != nil {
			logrus.Fatalf("Failed to listen on %s: %v", addr, err)
			return
		}
	}()
	<-ctx.Done()
	return nil
}
