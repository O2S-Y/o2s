package dev

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/ui"
)

func jwtCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "jwt",
		Short: "JWT utilities (decode without verifying signatures)",
	}
	c.AddCommand(jwtDecodeCmd())
	return c
}

func jwtDecodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "decode <token>",
		Short: "Decode JWT header + payload (does NOT verify signature)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			raw := strings.TrimSpace(args[0])
			parts := strings.Split(raw, ".")
			if len(parts) < 2 {
				return fmt.Errorf("malformed jwt")
			}

			headerJSON, err := jwtB64URLDecodeJSON(parts[0])
			if err != nil {
				return fmt.Errorf("decode header: %w", err)
			}
			payloadJSON, err := jwtB64URLDecodeJSON(parts[1])
			if err != nil {
				return fmt.Errorf("decode payload: %w", err)
			}

			var hdr map[string]any
			var pay map[string]any
			_ = json.Unmarshal(headerJSON, &hdr)
			_ = json.Unmarshal(payloadJSON, &pay)

			out := map[string]any{
				"header":  hdr,
				"payload": pay,
			}

			pretty := ui.Headline("JWT") + "\n" +
				ui.KV([][2]string{
					{"alg", stringifyMap(hdr, "alg")},
					{"typ", stringifyMap(hdr, "typ")},
					{"kid", stringifyMap(hdr, "kid")},
				}) + "\n\n" +
				ui.T().Heading.Render("Header") + "\n" + string(prettyJSONBytes(headerJSON)) + "\n\n" +
				ui.T().Heading.Render("Payload") + "\n" + string(prettyJSONBytes(payloadJSON))

			return ui.Render(out, pretty)
		},
	}
}

func jwtB64URLDecodeJSON(seg string) ([]byte, error) {
	if m := len(seg) % 4; m != 0 {
		seg += strings.Repeat("=", 4-m)
	}
	seg = strings.ReplaceAll(seg, "-", "+")
	seg = strings.ReplaceAll(seg, "_", "/")
	raw, err := base64.StdEncoding.DecodeString(seg)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func stringifyMap(m map[string]any, key string) string {
	if m == nil {
		return "-"
	}
	v, ok := m[key]
	if !ok {
		return "-"
	}
	switch x := v.(type) {
	case string:
		return x
	default:
		b, _ := json.Marshal(x)
		return string(b)
	}
}

func prettyJSONBytes(b []byte) []byte {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return b
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return b
	}
	return out
}
