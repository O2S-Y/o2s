package dev

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/ui"
)

func httpCmd() *cobra.Command {
	var (
		header []string
		body   string
		data   string
		tout   time.Duration
	)
	c := &cobra.Command{
		Use:   "http <method> <url>",
		Short: "Tiny curl-like HTTP client (good for quick API checks)",
		Long: `Examples:

  o2s dev http GET https://example.com
  o2s dev http POST https://httpbin.org/post --data '{"hello":"world"}' -H 'Content-Type: application/json'`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			method := strings.ToUpper(args[0])
			url := args[1]
			if !strings.HasPrefix(strings.ToLower(url), "http://") && !strings.HasPrefix(strings.ToLower(url), "https://") {
				url = "https://" + url
			}

			var rdr io.Reader
			switch {
			case body != "" && data != "":
				return fmt.Errorf("use only one of --body or --data")
			case body != "":
				b, err := os.ReadFile(body)
				if err != nil {
					return err
				}
				rdr = bytes.NewReader(b)
			case data != "":
				rdr = strings.NewReader(data)
			}

			req, err := http.NewRequest(method, url, rdr)
			if err != nil {
				return err
			}
			for _, h := range header {
				k, v, ok := strings.Cut(h, ":")
				if !ok {
					k, v, ok = strings.Cut(h, "=")
				}
				if !ok {
					return fmt.Errorf("bad header %q (expected Key: Value or Key=Value)", h)
				}
				req.Header.Set(strings.TrimSpace(k), strings.TrimSpace(v))
			}

			cli := &http.Client{Timeout: tout}
			ui.Info(method + " " + url)
			res, err := cli.Do(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			b, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			type out struct {
				Status     int               `json:"status"`
				StatusText string            `json:"status_text"`
				Headers    map[string]string `json:"headers"`
				Body       string            `json:"body"`
			}
			hmap := map[string]string{}
			for k, vals := range res.Header {
				if len(vals) > 0 {
					hmap[k] = vals[0]
				}
			}
			payload := out{
				Status:     res.StatusCode,
				StatusText: res.Status,
				Headers:    hmap,
				Body:       string(b),
			}

			t := ui.T()
			head := fmt.Sprintf("%s %d", t.Title.Render("HTTP"), res.StatusCode)
			lines := []string{head, t.Muted.Render(res.Status)}
			for k, v := range hmap {
				if strings.EqualFold(k, "Set-Cookie") {
					continue
				}
				lines = append(lines, t.Key.Render(k+":")+" "+v)
			}
			lines = append(lines, "", string(b))
			pretty := strings.Join(lines, "\n")
			return ui.Render(payload, pretty)
		},
	}
	c.Flags().StringArrayVarP(&header, "header", "H", nil, "header as 'Key: Value' or 'Key=value' (repeatable)")
	c.Flags().StringVar(&body, "body", "", "read request body from file")
	c.Flags().StringVar(&data, "data", "", "request body string")
	c.Flags().DurationVar(&tout, "timeout", 30*time.Second, "request timeout")
	return c
}
