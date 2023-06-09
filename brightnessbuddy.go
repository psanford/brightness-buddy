package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/godbus/dbus/v5"
)

func main() {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		log.Fatalf("Failed to connect to session bus: %s", err)
	}
	defer conn.Close()

	if err = conn.AddMatchSignal(
		dbus.WithMatchPathNamespace("/org/xfce"),
		dbus.WithMatchInterface("org.xfce.ScreenSaver"),
	); err != nil {
		log.Fatalf("Failed to connect to add signal match: %s", err)
	}

	signals := make(chan *dbus.Signal, 10)
	conn.Signal(signals)
	log.Print("brightness buddy start")

	for {
		select {
		case message := <-signals:
			log.Printf("ScreenSaver event: %+v\n", message)

			if message.Name != "org.xfce.ScreenSaver.ActiveChanged" {
				continue
			}

			if len(message.Body) < 1 {
				continue
			}

			screenSaverActive, ok := message.Body[0].(bool)

			if !ok {
				continue
			}

			if screenSaverActive {
				continue
			}

			log.Printf("Fiddling with brightness")
			out, err := exec.Command(helper(), "--get-brightness").CombinedOutput()
			if err != nil {
				log.Printf("get brightness err: %s", err)
				continue
			}

			brightnessStr := strings.TrimSpace(string(out))

			curBrightness, err := strconv.Atoi(brightnessStr)
			if err != nil {
				log.Printf("get brightness int err: %s", err)
				continue
			}

			setBrightness(curBrightness)
			if err != nil {
				panic(err)
			}
		}
	}
}

func helper() string {
	var helperRef = "/run/current-system/sw/sbin/xfpm-power-backlight-helper"
	s, err := os.Readlink(helperRef)
	if err != nil {
		panic(err)
	}

	return s
}

func setBrightness(n int) error {
	log.Println("run command: ", "pkexec", helper(), "--set-brightness", strconv.Itoa(n))
	out, err := exec.Command("pkexec", helper(), "--set-brightness", strconv.Itoa(n)).CombinedOutput()
	if err != nil {
		return fmt.Errorf("set brightness err: %w %s", err, out)
	}
	return nil
}
