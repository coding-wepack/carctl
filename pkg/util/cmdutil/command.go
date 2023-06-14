package cmdutil

import (
	"io"
	"os/exec"
	"runtime"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/coding-wepack/carctl/pkg/log"
	"github.com/coding-wepack/carctl/pkg/settings"
)

func PreRun(cmd *cobra.Command, args []string) {
	if settings.Verbose {
		// debug mode enable
		log.SetDebug()
	}
}

func Command(arg ...string) (result string, err error) {
	name, c := "/bin/bash", "-c"
	// 根据系统设定不同的命令name
	if runtime.GOOS == "windows" {
		name, c = "cmd", "/C"
	}
	arg = append([]string{c}, arg...)
	cmd := exec.Command(name, arg...)

	// 创建获取命令输出管道
	stderr, err := cmd.StderrPipe()
	if err != nil {
		err = errors.Wrapf(err, "failed to obtain stderr pipe for command")
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		err = errors.Wrapf(err, "failed to obtain stdout pipe for command")
		return
	}

	// 执行命令
	if err = cmd.Start(); err != nil {
		err = errors.Wrapf(err, "exec command is err")
		return
	}

	// 读取所有输出
	stdoutBytes, err := io.ReadAll(stdout)
	if err != nil {
		err = errors.Wrapf(err, "read all command stdout failed")
		return
	}

	// 读取错误输出
	stderrBytes, err := io.ReadAll(stderr)
	if err != nil {
		err = errors.Wrapf(err, "read all command stderr failed")
		return
	}

	if len(stdoutBytes) != 0 {
		result = string(stdoutBytes)
	}

	if err = cmd.Wait(); err != nil {
		return string(stderrBytes), errors.Wrapf(err, "wait failed")
	}

	return string(stdoutBytes), nil
}

// func ExecCmdAndOutput(c string) (output string, err error) {
// 	var cmd *exec.Cmd
// 	switch runtime.GOOS {
// 	case "windows":
// 		cmd = exec.Command("cmd.exe", "/c", c)
// 	case "linux", "darwin":
// 		cmd = exec.Command("bash", "-c", c)
// 	}
//
// 	//显示运行的命令
// 	stdout, err := cmd.StdoutPipe()
// 	if err != nil {
// 		log.Println("stdout pipe err,", err)
// 		return "", err
// 	}
// 	stderr, err := cmd.StderrPipe()
// 	if err != nil {
// 		log.Println("stderr pipe err,", err)
// 		return "", err
// 	}
// 	var wg sync.WaitGroup
// 	wg.Add(2)
// 	go read(context.Background(), &wg, stderr, nil)
// 	go read(context.Background(), &wg, stdout, &output)
// 	err = cmd.Start()
// 	if err != nil {
// 		log.Println("stdout pipe err,", err)
// 		return "", err
// 	}
// 	wg.Wait()
// 	_ = cmd.Wait()
// 	if !cmd.ProcessState.Success() {
// 		// 执行失败，返回错误信息
// 		return output, errors.New("<failed>")
// 	}
//
// 	return output, nil
// }
//
// func read(ctx context.Context, wg *sync.WaitGroup, std io.ReadCloser, output *string) {
// 	reader := bufio.NewReader(std)
// 	defer wg.Done()
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		default:
// 			readString, err := reader.ReadString('\n')
// 			if err != nil || err == io.EOF {
// 				return
// 			}
// 			fmt.Print(readString)
// 			if output != nil {
// 				*output += readString
// 			}
// 		}
// 	}
// }
//
// func ExecCmd(c string) error {
// 	_, err := ExecCmdAndOutput(c)
// 	return err
// }
