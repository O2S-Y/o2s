package dev

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/ui"
)

func serveCmd() *cobra.Command {
	var port int
	c := &cobra.Command{
		Use:   "serve [path]",
		Short: "Quick static HTTP server for the given directory",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}
			abs, err := filepath.Abs(root)
			if err != nil {
				return err
			}
			info, err := os.Stat(abs)
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return fmt.Errorf("%s is not a directory", abs)
			}

			addr := fmt.Sprintf(":%d", port)
			t := ui.T()
			ips := localIPs()
			ui.Println(ui.Headline("o2s dev serve"))
			ui.Println(ui.KV([][2]string{
				{"root", abs},
				{"local", "http://localhost" + addr},
				{"network", joinDash(ips, addr)},
			}))
			ui.Println(t.Muted.Render("press Ctrl+C to stop"))

			srv := &http.Server{
				Addr:              addr,
				Handler:           withLog(http.FileServer(http.Dir(abs))),
				ReadHeaderTimeout: 5 * time.Second,
			}
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
			defer cancel()
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					ui.Danger("serve error: " + err.Error())
				}
			}()
			<-ctx.Done()
			c2, c2cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer c2cancel()
			_ = srv.Shutdown(c2)
			ui.Success("stopped")
			return nil
		},
	}
	c.Flags().IntVarP(&port, "port", "p", 8080, "TCP port to listen on")
	return c
}

func withLog(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := ui.T()
		ui.Println(t.Muted.Render(time.Now().Format("15:04:05")) + " " + t.Key.Render(r.Method) + " " + r.URL.Path)
		h.ServeHTTP(w, r)
	})
}

func localIPs() []string {
	out := []string{}
	ifaces, _ := net.Interfaces()
	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagLoopback != 0 || ifc.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, _ := ifc.Addrs()
		for _, a := range addrs {
			if ipn, ok := a.(*net.IPNet); ok {
				if v4 := ipn.IP.To4(); v4 != nil {
					out = append(out, v4.String())
				}
			}
		}
	}
	return out
}

func joinDash(xs []string, suffix string) string {
	if len(xs) == 0 {
		return "-"
	}
	out := ""
	for i, x := range xs {
		if i > 0 {
			out += ", "
		}
		out += "http://" + x + suffix
	}
	return out
}
