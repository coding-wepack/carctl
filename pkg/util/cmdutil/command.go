package cmdutil

import (
	"bufio"
	"context"
	"io"
	stdLog "log"
	"os/exec"
	"runtime"
	"sync"

	"github.com/coding-wepack/carctl/pkg/log"
	"github.com/coding-wepack/carctl/pkg/settings"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func PreRun(cmd *cobra.Command, args []string) {
	if settings.Verbose {
		// debug mode enable
		log.SetDebug()
	}
}

func Command(c string) (output string, err error) {
	if settings.Verbose {
		log.Debug(c)
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd.exe", "/c", c)
	case "linux", "darwin":
		cmd = exec.Command("bash", "-c", c)
	}

	// 显示运行的命令
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", errors.Wrapf(err, "get stdout pipe failed")
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdLog.Println("stderr pipe err,", err)
		return "", errors.Wrapf(err, "get stderr pipe failed")
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go read(context.Background(), &wg, stderr, nil)
	go read(context.Background(), &wg, stdout, &output)
	err = cmd.Start()
	if err != nil {
		return "", errors.Wrapf(err, "exec cmd failed")
	}
	wg.Wait()
	_ = cmd.Wait()
	if !cmd.ProcessState.Success() {
		// 执行失败，返回错误信息
		return output, errors.New("failed")
	}
	return output, nil
}

func read(ctx context.Context, wg *sync.WaitGroup, std io.ReadCloser, output *string) {
	reader := bufio.NewReader(std)
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			readString, err := reader.ReadString('\n')
			if err != nil || err == io.EOF {
				return
			}
			if output != nil {
				*output += readString
			}
		}
	}
}
