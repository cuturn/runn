/*
Copyright © 2022 Ken'ichiro Oyama <k1lowxb@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/k1LoW/runn"
	"github.com/k1LoW/runn/capture"
	"github.com/k1LoW/runn/internal/scope"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// newCmd represents the new command.
var newCmd = &cobra.Command{
	Use:     "new",
	Short:   "create new runbook or append step to runbook",
	Long:    `create new runbook or append step to runbook.`,
	Aliases: []string{"append"},
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			o   *os.File
			err error
			al  [][]string
		)
		if len(args) == 0 {
			if isatty.IsTerminal(os.Stdin.Fd()) {
				return errors.New("interactive mode is planned, but not yet implemented")
			}
			al = argsListFromStdin(os.Stdin)
		} else {
			al = [][]string{args}
		}
		ctx := context.Background()
		rb := runn.NewRunbook(flgs.Desc)
		if flgs.Out != "" {
			p := filepath.Clean(flgs.Out)
			if _, err := os.Stat(p); err == nil {
				f, err := os.Open(p)
				if err != nil {
					return err
				}
				rb, err = runn.ParseRunbook(f)
				if err != nil {
					return err
				}
				if flgs.Desc != "" {
					rb.Desc = flgs.Desc
				}
			}
		}
		for _, args := range al {
			if err := rb.AppendStep(args...); err != nil {
				return err
			}
		}
		if flgs.Out == "" {
			o = os.Stdout
		} else {
			o, err = os.Create(filepath.Clean(flgs.Out))
			if err != nil {
				return err
			}
			defer func() {
				if err := o.Close(); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
					os.Exit(1)
				}
			}()
		}

		fn := func(o *os.File) error {
			enc := yaml.NewEncoder(o)
			if err := enc.Encode(rb); err != nil {
				return err
			}
			return nil
		}

		if flgs.AndRun {
			if err := runAndCapture(ctx, o, fn); err != nil {
				return err
			}
		} else {
			if err := fn(o); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVarP(&flgs.Desc, "desc", "", "", flgs.Usage("Desc"))
	newCmd.Flags().StringVarP(&flgs.Out, "out", "", "", flgs.Usage("Out"))
	newCmd.Flags().BoolVarP(&flgs.AndRun, "and-run", "", false, flgs.Usage("AndRun"))
	newCmd.Flags().BoolVarP(&flgs.GRPCNoTLS, "grpc-no-tls", "", false, flgs.Usage("GRPCNoTLS"))
	newCmd.Flags().StringSliceVarP(&flgs.GRPCProtos, "grpc-proto", "", []string{}, flgs.Usage("GRPCProtos"))
	newCmd.Flags().StringSliceVarP(&flgs.GRPCImportPaths, "grpc-import-path", "", []string{}, flgs.Usage("GRPCImportPaths"))
}

func runAndCapture(ctx context.Context, o *os.File, fn func(*os.File) error) error {
	const newf = "new.yml"
	td, err := os.MkdirTemp("", "runn")
	if err != nil {
		return err
	}
	defer os.RemoveAll(td)
	tf, err := os.Create(filepath.Clean(filepath.Join(td, newf)))
	if err != nil {
		return err
	}
	defer func() {
		if err := tf.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	}()

	if err := fn(tf); err != nil {
		return err
	}

	opts := []runn.Option{
		runn.Book(tf.Name()),
		runn.Capture(capture.Runbook(td, capture.RunbookLoadDesc(true))),
		runn.GRPCNoTLS(flgs.GRPCNoTLS),
		runn.GRPCProtos(flgs.GRPCProtos),
		runn.GRPCImportPaths(flgs.GRPCImportPaths),
		runn.Scopes(scope.AllowReadParent),
	}
	oo, err := runn.New(opts...)
	if err != nil {
		return err
	}
	if err := oo.Run(ctx); err != nil {
		return err
	}

	entries, err := os.ReadDir(td)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.Name() != newf {
			b, err := os.ReadFile(filepath.Join(td, e.Name()))
			if err != nil {
				return err
			}
			if _, err := o.Write(b); err != nil {
				return err
			}
		}
	}

	return nil
}

func argsListFromStdin(in io.Reader) [][]string {
	var al [][]string
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		al = append(al, []string{scanner.Text()})
	}
	return al
}
