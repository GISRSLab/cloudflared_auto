package client

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"

	"github.com/rivo/tview"
)

const servicePrefix = "cloudflared-"

func installService(name string, desc string, hostname string, app *tview.Application, logView *tview.TextView) error {
	serviceName := servicePrefix + name

	// 查找一个大于3000的未使用端口
	port, err1 := findAvailablePort()
	if err1 != nil {
		go func() {
			app.QueueUpdateDraw(func() {
				logView.SetText(err1.Error())
			})
		}()
		return err1
	}

	// 构建服务执行路径
	execPath := fmt.Sprintf("cloudflared access rdp --hostname %s --url rdp://localhost:%d", hostname, port)

	cmd := exec.Command("sc", "create", serviceName, "binpath=", execPath, "start=auto")

	err := cmd.Run()
	// 创建服务
	if err != nil {
		if err.Error() == "exit status 5" {
			return errors.New("可能是由于权限问题导致错误，请用管理员权限运行软件, exit status 5")
		}
		return err
	}

	// 启动服务
	err = exec.Command("sc", "start", serviceName).Run()
	if err != nil {
		return err
	}

	// 异步更新界面
	go func() {
		app.QueueUpdateDraw(func() {
			logView.SetText(fmt.Sprintf("\nService %s installed and started successfully\nYour RDP is running at localhost:%d\n", serviceName, port))
		})
	}()

	return nil
}

func findAvailablePort() (int, error) {
	for port := 3000; port < 65535; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			ln.Close()
			return port, nil
		}
	}
	return 0, errors.New("No port available!")
}

func listService() ([]string, error) {
	// 创建第一个命令：sc query
	cmd1 := exec.Command("sc", "query")

	// 创建第二个命令：findstr /R "cloudflared-"
	cmd2 := exec.Command("findstr", "/R", "cloudflared-")

	// 创建缓冲区来捕获第一个命令的输出
	var stdout1 bytes.Buffer
	cmd1.Stdout = &stdout1

	// 将第一个命令的输出作为第二个命令的输入
	cmd2.Stdin = &stdout1

	// 创建缓冲区来捕获第二个命令的输出
	var stdout2 bytes.Buffer
	cmd2.Stdout = &stdout2

	// 执行第一个命令
	err := cmd1.Run()
	if err != nil {
		if err.Error() == "exit status 5" {
			return make([]string, 0), errors.New("可能是由于权限问题导致错误，请用管理员权限运行软件, exit status 5")
		}
		return make([]string, 0), err
	}

	// 执行第二个命令
	err = cmd2.Run()
	if err != nil {
		if err.Error() == "exit status 5" {
			return make([]string, 0), errors.New("可能是由于权限问题导致错误，请用管理员权限运行软件, exit status 5")
		}
		return make([]string, 0), err
	}

	// 处理第二个命令的输出，提取服务名称
	output := stdout2.String()
	lines := strings.Split(output, "\n")
	var serviceNames []string

	// 正则表达式匹配服务名称
	re := regexp.MustCompile(`SERVICE_NAME:\s*(cloudflared-[^ ]+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				serviceNames = append(serviceNames, matches[1])
			}
		}
	}
	return serviceNames, nil

}

func deleteService(name string) error {
	cmd1 := exec.Command("sc", "stop", name)
	err1 := cmd1.Run()
	if err1 != nil {
		if err1.Error() == "exit status 5" {
			return errors.New("可能是由于权限问题导致错误，请用管理员权限运行软件, exit status 5")
		}
		return err1
	}

	cmd2 := exec.Command("sc", "delete", name)
	err2 := cmd2.Run()
	if err2 != nil {
		if err2.Error() == "exit status 5" {
			return errors.New("可能是由于权限问题导致错误，请用管理员权限运行软件, exit status 5")
		}
		return err2
	}
	return nil
}
