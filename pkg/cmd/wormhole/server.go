package wormhole

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/sapcc/kubernikus/pkg/cmd"
	"github.com/sapcc/kubernikus/pkg/wormhole"
)

func NewServerCommand(logger log.Logger) *cobra.Command {
	o := NewServerOptions(log.With(logger, "wormhole", "server"))

	c := &cobra.Command{
		Use:   "server",
		Short: "Creates a Wormhole Server",
		Run: func(c *cobra.Command, args []string) {
			cmd.CheckError(o.Validate(c, args))
			cmd.CheckError(o.Complete(args))
			cmd.CheckError(o.Run(c))
		},
	}

	o.BindFlags(c.Flags())

	return c
}

type ServerOptions struct {
	wormhole.ServerOptions
}

func NewServerOptions(logger log.Logger) *ServerOptions {
	o := &ServerOptions{}
	o.Logger = logger
	return o
}

func (o *ServerOptions) BindFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.KubeConfig, "kubeconfig", o.KubeConfig, "Path to the kubeconfig file to use to talk to the Kubernetes apiserver. If unset, try the environment variable KUBECONFIG, as well as in-cluster configuration")
	flags.StringVar(&o.Context, "context", "", "Override context")
	flags.StringVar(&o.ClientCA, "ca", o.ClientCA, "CA to use for validating tunnel clients")
	flags.StringVar(&o.Certificate, "cert", o.Certificate, "Certificate for the tunnel server")
	flags.StringVar(&o.PrivateKey, "key", o.PrivateKey, "Key for the tunnel server")
	flags.StringVar(&o.ServiceCIDR, "service-cidr", "", "Cluster service IP range")
}

func (o *ServerOptions) Validate(c *cobra.Command, args []string) error {
	if o.ServiceCIDR == "" {
		return errors.New("You must specify service-cidr")
	}
	return nil
}

func (o *ServerOptions) Complete(args []string) error {
	return nil
}

func (o *ServerOptions) Run(c *cobra.Command) error {
	sigs := make(chan os.Signal, 1)
	stop := make(chan struct{})
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM) // Push signals into channel
	wg := &sync.WaitGroup{}                            // Goroutines can add themselves to this to be waited on

	server, err := wormhole.NewServer(&o.ServerOptions)
	if err != nil {
		return fmt.Errorf("Failed to initialize server: %s", err)
	}

	go server.Run(stop, wg)

	<-sigs // Wait for signals (this hangs until a signal arrives)
	o.Logger.Log("msg", "Shutting down...")
	close(stop) // Tell goroutines to stop themselves
	wg.Wait()   // Wait for all to be stopped

	return nil
}
