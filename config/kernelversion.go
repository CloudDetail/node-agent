package config

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
)

func KernelBlow317() bool {
	var Kernel317 bool
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	// 解析版本号
	kernelVersionText := strings.TrimSpace(string(output))
	kernelVersion, err := strconv.Atoi(strings.Split(kernelVersionText, ".")[0])
	if err != nil {
		panic(err)
	}

	// 比较版本
	if kernelVersion >= 3 && kernelVersion >= 17 {
		Kernel317 = false
	} else {
		Kernel317 = true
	}
	log.Printf("kernel version: %s", kernelVersionText)
	log.Printf("Kernel317: %s", strconv.FormatBool(Kernel317))
	return Kernel317

}
