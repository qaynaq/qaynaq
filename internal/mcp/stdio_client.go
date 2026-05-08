package mcp

import (
	"context"
	"io"
	"os"
	"os/exec"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
)

func newStdioMCPClientWithStderr(command string, env []string, args []string, ring io.Writer) (*mcpclient.Client, error) {
	cmdFunc := func(ctx context.Context, cmd string, e []string, a []string) (*exec.Cmd, error) {
		c := exec.CommandContext(ctx, cmd, a...)
		c.Env = append(os.Environ(), e...)
		setProcessGroup(c)
		return c, nil
	}

	stdioTransport := transport.NewStdioWithOptions(
		command,
		env,
		args,
		transport.WithCommandFunc(cmdFunc),
	)

	if err := stdioTransport.Start(context.Background()); err != nil {
		return nil, err
	}

	if ring != nil {
		go func() {
			_, _ = io.Copy(ring, stdioTransport.Stderr())
		}()
	}

	return mcpclient.NewClient(stdioTransport), nil
}
