package debug

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strconv"

	"github.com/fukata/golang-stats-api-handler"
)

type ConfigServer struct {
	Port string `default:"8080" desc:"port of http server"` // CFMR_SERVER_PORT
}

type statsHandler struct {
	stats *Stats
}

// Server is used for various debugging.
// It opens runtime stats, pprof and appliclation stats.
type Server struct {
	Port   string
	Logger *log.Logger
	Stats  *Stats
}

// Start starts listening.
func NewServer(cs *ConfigServer, stats *Stats, logger *log.Logger) (*Server, error) {
	return &Server{Port: cs.Port, Stats: stats, Logger: logger}, nil
}

func (s *Server) Start() *http.Server {
	srv := &http.Server{Addr: ":" + s.Port}

	http.HandleFunc("/", index)
	http.Handle("/stats/app", &statsHandler{
		stats: s.Stats,
	})
	http.HandleFunc("/stats/runtime", stats_api.Handler)

	go func() {
		s.Logger.Printf("[INFO] Start server listening on :%s", s.Port)
		err := srv.ListenAndServe()
		if err != nil {
			s.Logger.Println("[ERROR] Failed to start Http Server", err)
		}
	}()

	return srv
}

func index(w http.ResponseWriter, _ *http.Request) {
	body := `
		<a href="https://github.com/rakutentech/cf-metrics-refinery">cf-metrics-refinery</a> 
		<ul>
		  <li><a href="/stats/runtime">stats/runtime</a></li>
		  <li><a href="/debug/pprof/">pprof</a></li>
		  <li><a href="/stats/app">stats/app</a></li>
		</ul>
		      `
	w.Header().Set("Content-type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(body))
}

func (h *statsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := h.stats.Json()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error: %s\n", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Contentt-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
