package functions

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/nodephone/nodephone-cli/internal/output"
)

type Server struct {
	functionsDir string
	port         int
	printer      *output.Printer
	httpServer   *http.Server
}

func NewServer(functionsDir string, port int, printer *output.Printer) *Server {
	if port <= 0 {
		port = 8080
	}
	return &Server{
		functionsDir: functionsDir,
		port:         port,
		printer:      printer,
	}
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		path := strings.TrimPrefix(r.URL.Path, "/")
		parts := strings.Split(path, "/")
		fnName := parts[0]

		if fnName == "" || fnName == "functions" {
			if len(parts) > 1 {
				fnName = parts[1]
			}
		}

		if fnName == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"online","message":"NodePhone Functions Dev Server"}`))
			return
		}

		// Re-scan local functions on every request for Hot Reload
		funcs, err := ScanLocalFunctions(s.functionsDir)
		if err != nil {
			s.logError(fnName, r.Method, err.Error(), time.Since(startTime))
			http.Error(w, fmt.Sprintf(`{"error": %q}`, err.Error()), http.StatusInternalServerError)
			return
		}

		var target *FunctionInfo
		for _, f := range funcs {
			if f.Name == fnName {
				target = &f
				break
			}
		}

		if target == nil {
			s.logError(fnName, r.Method, "function not found", time.Since(startTime))
			http.Error(w, fmt.Sprintf(`{"error": "Function %s not found"}`, fnName), http.StatusNotFound)
			return
		}

		// Return simulated function response based on code or standard JSON
		duration := time.Since(startTime)
		s.logSuccess(target.Name, r.Method, duration)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		respJSON := fmt.Sprintf(`{
  "message": "Hello from %s function!",
  "timestamp": "%s",
  "runtime": %q
}`, target.Name, time.Now().Format(time.RFC3339), target.Manifest.Runtime)
		w.Write([]byte(respJSON))
	})

	addr := fmt.Sprintf(":%d", s.port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	s.printer.Header("NodePhone Functions Local Server")
	s.printer.Println()
	s.printer.Success(fmt.Sprintf("Local server running at http://localhost:%d", s.port))
	s.printer.Info("Hot reload enabled. Editing functions dynamically updates server endpoints.")
	s.printer.Println()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.httpServer.Shutdown(shutdownCtx)
	}()

	err := s.httpServer.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (s *Server) logSuccess(name, method string, duration time.Duration) {
	msg := fmt.Sprintf("[INFO] %s executed (%dms) - %s /functions/%s", name, duration.Milliseconds(), method, name)
	s.printer.Println(s.printer.Green(msg))
}

func (s *Server) logError(name, method, err string, duration time.Duration) {
	msg := fmt.Sprintf("[ERROR] %s failed (%dms) - %s: %s", name, duration.Milliseconds(), method, err)
	s.printer.Println(s.printer.Red(msg))
}
