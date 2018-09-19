package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/gin-gonic/gin"

	"./gopushserver"
	"github.com/appleboy/gorush/gorush"
	"github.com/facebookgo/grace/gracehttp"
	"golang.org/x/sync/errgroup"
)

func main() {
	var (
		ping        bool
		releaseMode bool
	)

	flag.BoolVar(&ping, "ping", false, "ping server")
	flag.BoolVar(&releaseMode, "prod", false, "run in development mode")
	flag.Usage = usage
	flag.Parse()

	if releaseMode {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	gopushserver.InitLog()

	if ping {
		if err := pinger(); err != nil {
			gopushserver.LogError.Warnf("ping server error: %v", err)
		}
		return
	}

	gopushserver.InitWorkers(int64(runtime.NumCPU()), int64(8192))

	var g errgroup.Group
	g.Go(RunHTTPServer)

	var err error
	if err = g.Wait(); err != nil {
		gorush.LogError.Fatal(err)
	}
}

func RunHTTPServer() (err error) {
	err = gracehttp.Serve(&http.Server{
		Addr:    ":3000",
		Handler: gopushserver.GetRouterEngine(),
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
    -p, --port <port>                Use port for clients (default: 8088)
    --ping                           healthy check command for container
`

// usage will print out the flag options for the server.
func usage() {
	fmt.Printf("%s\n", usageStr)
	os.Exit(0)
}
