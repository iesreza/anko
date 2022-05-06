package watcher

import (
	"fmt"
	"os/exec"
	"strings"
	"github.com/GuilhermeCaruso/anko/internal/banner"
)

// InitApp is responsible for initializing the configured application using anko.yaml
// to configure the primary language and processes.
var cmd *exec.Cmd
func (wc *Watcher) InitApp() {
	

	lang, err := GetLanguage(wc.Language)

	wc.selectedLanguage = lang

	if err != nil {
		banner.Error(err.Error())
	}

	if wc.selectedLanguage.ExecCmd == "" {
		cmd = exec.Command(wc.selectedLanguage.ExecPath, wc.AppPath)
	} else {
		cmd = exec.Command(wc.selectedLanguage.ExecPath, wc.selectedLanguage.ExecCmd, wc.AppPath)

	}
	stdout, err := cmd.StdoutPipe()

	cmd.Stderr = cmd.Stdout

	if err != nil {
		banner.Error(err.Error())
		wc.DoneChan <- true
	}

	if err = cmd.Start(); err != nil {

		banner.Error(err.Error())
		wc.DoneChan <- true
	}

	for {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)

		if !strings.Contains(string(tmp), "signal: killed") {
			fmt.Print(string(tmp))
		}

		if err != nil {
			stdout.Close()
			banner.Restarting()
			break

		}
	}
}

func (wc *Watcher) resetApp() {
	if cmd == nil || cmd.Process == nil{
		go wc.InitApp()
		return
	}

	var err = cmd.Process.Kill()
	if err != nil {
		banner.Error(err.Error())
		return
	}

	go wc.InitApp()
	
}

// AppController is the main channel used for control anko actions
// like reset, start and other future options
func (wc *Watcher) AppController() {
	openDispacher := true
	go wc.InitApp()
	for {
		select {
		case action := <-wc.DispatcherChan:
			switch action {
			case ACT_INIT:
				go wc.InitApp()
				wc.IsOpen = &openDispacher
			case ACT_RESET:
				wc.resetApp()
				wc.IsOpen = &openDispacher
			}
		}
	}
}
