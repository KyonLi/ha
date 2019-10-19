package adb

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type Config struct {
	IP    string
	Port  string
}

type Target struct {
	config      Config
	isConnected bool
}

func NewTarget(config Config) Target {
	return Target{config: config}
}

func (t *Target) Reconnect() error {
	log.Println("reconnecting...")
	cmd := exec.Command("adb", "disconnect")
	//log.Printf("run: %s %s\n", cmd.Path, cmd.Args)
	_, err := cmd.CombinedOutput()
	//log.Printf("out: %s\n", out)
	if err != nil {
		t.isConnected = false
		log.Println("connect failed")
		return err
	}

	cmd = exec.Command("adb", "connect", fmt.Sprintf("%s:%s", t.config.IP, t.config.Port))
	//log.Printf("run: %s %s\n", cmd.Path, cmd.Args)
	_, err = cmd.CombinedOutput()
	//log.Printf("out: %s\n", out)
	if err != nil {
		t.isConnected = false
		log.Println("connect failed")
		return err
	}
	t.isConnected = true
	return nil
}

func (t *Target) IsAwake() (bool, error) {
	log.Println("detecting wake status...")
	cmd := exec.Command("adb", "shell", "dumpsys power | grep mWakefulness=")
	//log.Printf("run: %s %s\n", cmd.Path, cmd.Args)
	out, err := cmd.CombinedOutput()
	//log.Printf("out: %s\n", out)
	outStr := string(out)
	if err != nil {
		if strings.Contains(outStr, "no devices") || strings.Contains(outStr, "device offline") {
			err := t.Reconnect()
			if err != nil {
				return false, err
			}
			return t.IsAwake()
		}
		return false, err
	}
	if strings.Contains(outStr, "Awake") {
		return true, nil
	} else if strings.Contains(outStr, "Asleep") {
		return false, nil
	}
	return false, fmt.Errorf("unknown status")
}

func (t *Target) SendKey(code KeyCode) error {
	log.Printf("sending keycode %d...", code)
	cmd := exec.Command("adb", "shell", fmt.Sprintf("input keyevent %d", code))
	//log.Printf("run: %s %s\n", cmd.Path, cmd.Args)
	out, err := cmd.CombinedOutput()
	//log.Printf("out: %s\n", out)
	outStr := string(out)
	if err != nil {
		if strings.Contains(outStr, "no devices") || strings.Contains(outStr, "device offline") {
			err := t.Reconnect()
			if err != nil {
				return err
			}
			return t.SendKey(code)
		}
		log.Println("send failed")
		return err
	}
	return nil
}
