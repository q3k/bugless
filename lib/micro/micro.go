package micro

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	log "github.com/inconshreveable/log15"
	pki "github.com/q3k/hspki"
	statusz "github.com/q3k/statusz"
	"golang.org/x/net/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	flagListenAddress string
	flagDebugAddress  string
	flagDebugAllowAll bool
)

func init() {
	flag.StringVar(&flagListenAddress, "listen_address", "127.0.0.1:4200", "gRPC listen address")
	flag.StringVar(&flagDebugAddress, "debug_address", "127.0.0.1:4201", "HTTP debug/status listen address")
	flag.BoolVar(&flagDebugAllowAll, "debug_allow_all", false, "HTTP debug/status available to everyone")
	flag.Set("logtostderr", "true")
}

type Mirko struct {
	grpcListen net.Listener
	grpcServer *grpc.Server
	httpListen net.Listener
	httpServer *http.Server
	httpMux    *http.ServeMux

	ctx    context.Context
	cancel context.CancelFunc
}

func New() *Mirko {
	ctx, cancel := context.WithCancel(context.Background())
	return &Mirko{
		ctx:    ctx,
		cancel: cancel,
	}
}

func authRequest(req *http.Request) (any, sensitive bool) {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		host = req.RemoteAddr
	}

	if flagDebugAllowAll {
		return true, true
	}

	switch host {
	case "localhost", "127.0.0.1", "::1":
		return true, true
	default:
		return false, false
	}
}

func (m *Mirko) Listen() error {
	grpc.EnableTracing = true
	trace.AuthRequest = authRequest

	grpcLis, err := net.Listen("tcp", flagListenAddress)
	if err != nil {
		return fmt.Errorf("net.Listen: %v", err)
	}
	m.grpcListen = grpcLis
	m.grpcServer = grpc.NewServer(pki.WithServerHSPKI()...)
	reflection.Register(m.grpcServer)

	httpLis, err := net.Listen("tcp", flagDebugAddress)
	if err != nil {
		return fmt.Errorf("net.Listen: %v", err)
	}

	m.httpMux = http.NewServeMux()
	// Canonical URLs
	m.httpMux.HandleFunc("/debug/status", func(w http.ResponseWriter, r *http.Request) {
		any, _ := authRequest(r)
		if !any {
			http.Error(w, "not allowed", http.StatusUnauthorized)
			return
		}
		statusz.StatusHandler(w, r)
	})
	m.httpMux.HandleFunc("/debug/requests", trace.Traces)

	// -z legacy URLs
	m.httpMux.HandleFunc("/statusz", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/debug/status", http.StatusSeeOther)
	})
	m.httpMux.HandleFunc("/rpcz", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/debug/requests", http.StatusSeeOther)
	})
	m.httpMux.HandleFunc("/requestz", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/debug/requests", http.StatusSeeOther)
	})

	// root redirect
	m.httpMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/debug/status", http.StatusSeeOther)
	})

	m.httpListen = httpLis
	m.httpServer = &http.Server{
		Addr:    flagDebugAddress,
		Handler: m.httpMux,
	}

	return nil
}

// Trace logs debug information to either a context trace (if present)
// or stderr (if not)
func Trace(ctx context.Context, f string, args ...interface{}) {
	tr, ok := trace.FromContext(ctx)
	if !ok {
		fmtd := fmt.Sprintf(f, args...)
		log.Warning("no trace", "ctx", ctx, "msg", fmtd)
		return
	}
	tr.LazyPrintf(f, args...)
}

// GRPC returns the microservice's grpc.Server object
func (m *Mirko) GRPC() *grpc.Server {
	if m.grpcServer == nil {
		panic("GRPC() called before Listen()")
	}
	return m.grpcServer
}

// HTTPMux returns the microservice's debug HTTP mux
func (m *Mirko) HTTPMux() *http.ServeMux {
	if m.httpMux == nil {
		panic("HTTPMux() called before Listen()")
	}
	return m.httpMux
}

// Context returns a background microservice context that will be canceled
// when the service is shut down
func (m *Mirko) Context() context.Context {
	return m.ctx
}

// Done() returns a channel that will emit a value when the service is
// shut down. This should be used in the main() function instead of a select{}
// call, to allow the background context to be canceled fully.
func (m *Mirko) Done() <-chan struct{} {
	return m.Context().Done()
}

// Serve starts serving HTTP and gRPC requests
func (m *Mirko) Serve() error {
	errs := make(chan error, 1)
	go func() {
		if err := m.grpcServer.Serve(m.grpcListen); err != nil {
			errs <- err
		}
	}()
	go func() {
		if err := m.httpServer.Serve(m.httpListen); err != nil {
			errs <- err
		}
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		select {
		case <-signalCh:
			m.cancel()
		}
	}()

	ticker := time.NewTicker(1 * time.Second)
	select {
	case <-ticker.C:
		log.Info("listening", "grpc_addr", flagListenAddress, "http_addr", flagDebugAddress)
		return nil
	case err := <-errs:
		return err
	}
}
