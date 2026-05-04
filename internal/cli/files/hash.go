package files

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/ui"
)

func hashCmd() *cobra.Command {
	var algo string
	c := &cobra.Command{
		Use:   "hash <file> [file...]",
		Short: "Compute checksums (md5, sha1, sha256, sha512)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			factory, err := hashFactory(algo)
			if err != nil {
				return err
			}
			results := make(map[string]string, len(args))
			for _, p := range args {
				sp := ui.NewSpinner("hashing " + p)
				sp.Start()
				digest, err := hashFile(p, factory())
				if err != nil {
					sp.Failf("%s: %v", p, err)
					continue
				}
				sp.Successf("%s  %s", digest, p)
				results[p] = digest
			}
			if ui.JSON() {
				return ui.Render(map[string]any{"algo": algo, "results": results}, "")
			}
			return nil
		},
	}
	c.Flags().StringVarP(&algo, "algo", "a", "sha256", "md5 | sha1 | sha256 | sha512")
	return c
}

func hashFactory(algo string) (func() hash.Hash, error) {
	switch algo {
	case "md5":
		return func() hash.Hash { return md5.New() }, nil
	case "sha1":
		return func() hash.Hash { return sha1.New() }, nil
	case "sha256":
		return func() hash.Hash { return sha256.New() }, nil
	case "sha512":
		return func() hash.Hash { return sha512.New() }, nil
	default:
		return nil, fmt.Errorf("unsupported algo: %s", algo)
	}
}

func hashFile(path string, h hash.Hash) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
