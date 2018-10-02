package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/gin-gonic/gin"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/matadesigns/gopushserver/server"
	"golang.org/x/sync/errgroup"
)

func main() {
	var (
		ping        bool
		releaseMode bool
	)

	flag.StringVar(&server.Address, "A", "", "address to bind")
	flag.StringVar(&server.Address, "address", "", "address to bind")
	flag.StringVar(&server.Port, "p", "", "port number for gorush")
	flag.StringVar(&server.Port, "port", "", "port number for gorush")
	flag.BoolVar(&ping, "ping", false, "ping server")
	flag.BoolVar(&releaseMode, "prod", false, "run in development mode")
	flag.Usage = usage
	flag.Parse()

	if server.Port == "" {
		server.Port = "3000"
	}

	if releaseMode {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	server.InitLog()

	if ping {
		if err := pinger(); err != nil {
			server.LogError.Warnf("ping server error: %v", err)
		}
		return
	}

	server.InitWorkers(int64(runtime.NumCPU()), int64(8192))

	var g errgroup.Group
	g.Go(RunHTTPServer)

	var err error
	if err = g.Wait(); err != nil {
		server.LogError.Fatal(err)
	}
}

func RunHTTPServer() (err error) {
	err = gracehttp.Serve(&http.Server{
		Addr:    server.Address + ":" + server.Port,
		Handler: server.GetRouterEngine(),
	})
	return
}

// handles pinging the endpoint and returns an error if the
// agent is in an unhealthy state.
func pinger() error {
	resp, err := http.Get("http://localhost:3000/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("server returned non-200 status code")
	}
	return nil
}

var usageStr = `
Usage: gopushserver [options]

Server Options:
    -A, --address <address>          Address to bind (default: any)
    -p, --port <port>                Use port for clients (default: 3000)
    --ping                           healthy check command for container
`

// usage will print out the flag options for the server.
func usage() {
	fmt.Printf("%s\n", usageStr)
	os.Exit(0)
}
