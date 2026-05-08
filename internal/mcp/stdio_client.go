package mcp

import (
	"context"
	"io"
	"os"
	"os/exec"
	"syscall"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
)

func newStdioMCPClientWithStderr(command string, env []string, args []string, ring io.Writer) (*mcpclient.Client, error) {
	cmdFunc := func(ctx context.Context, cmd string, e []string, a []string) (*exec.Cmd, error) {
		c := exec.CommandContext(ctx, cmd, a...)
		c.Env = append(os.Environ(), e...)
		// Setpgid puts the child in its own process group so a kill on the group
		// also reaps grandchildren. Without it, `npx <pkg>` exits but the Node
		// process it spawned hangs around as a zombie.
		c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
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
