package api

import (
	"net/http"
	"sync"

	log "github.com/golang/glog"
)

var running = false
var runningLock sync.Mutex

func SetRunning(value bool) {
	runningLock.Lock()
	running = value
	runningLock.Unlock()
}

func IsRunning() bool {
	runningLock.Lock()
	defer runningLock.Unlock()
	return running
}

func RunHttpServer() *http.Server {
	srv := &http.Server{Addr: ":8080"}

	go func() {
		log.Info("Server ready")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("Http listener failed: %v", err)
		} else if err == http.ErrServerClosed {
			log.Info("Server stopped")
			SetRunning(false)
		}
	}()

	SetRunning(true)

	return srv

}
