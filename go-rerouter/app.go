package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cilium/ebpf"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name: "start",
				Usage: "Start and attach the rerouter to the kernel, then exit. \n" +
					"If it was already started, this will simply reconfigure it.",
				Action: startCmd,
			},
			{
				Name:   "stop",
				Usage:  "Stop and detach the rerouter from the kernel, then exit",
				Action: stopCmd,
			},
			{
				Name:   "run",
				Usage:  "Like `start`, but the rerouter will exit when this process exits. (mostly meant for testing.)",
				Action: runCmd,
			},
			{
				Name:   "config",
				Usage:  "Manage various aspects of the rerouter",
				Action: configCmd,
				Flags: []cli.Flag{
					&cli.IntSliceFlag{
						Name:  "whitelist",
						Usage: "Don't reroute connections from the process listening on the given port. Can be repeated.",
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func reloadWhitelist() error {

}

func configCmd(ctx *cli.Context) error {
	removeMemlock()

	return fmt.Errorf("configCmd: not implemented :(")
}

func runCmd(ctx *cli.Context) error {
	removeMemlock()

	// Load the compiled eBPF ELF and load it into the kernel
	var objs rerouterObjects
	if err := loadRerouterObjects(&objs, nil); err != nil {
		log.Fatalf("Error loading eBPF program: %v", err)
	}
	defer objs.Close()

	links, err := attachProgs(objs)
	if err != nil {
		log.Fatalf("Couldn't attach some programs: %v", err)
	}
	defer links.Close()

	// Update the rerouter map with the proxy server configuration, because we need to know the proxy server PID in order
	// to filter out eBPF events generated by the proxy server itself so it would not proxy its own packets in a loop.
	var key uint32 = 0
	config := rerouterConfig{
		ProxyPort:     TUNNEL_PORT,
		ProxyPid:      uint64(os.Getpid()),
		RealProxyPort: REAL_PROXY_PORT,
		// RealProxyPid:  realProxyPid,
		WhitelistCount: 0,
	}
	err = objs.rerouterMaps.MapConfig.Update(&key, &config, ebpf.UpdateAny)
	if err != nil {
		log.Fatalf("Failed to update proxyMaps map: %v", err)
	}

}

func startCmd(ctx *cli.Context) error {
	removeMemlock()

	return fmt.Errorf("startCmd: not implemented :(")
}

func stopCmd(ctx *cli.Context) error {
	removeMemlock()

	return fmt.Errorf("stopCmd: not implemented :(")
}
