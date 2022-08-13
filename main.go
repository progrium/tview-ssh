package main

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/terminfo"
	"github.com/gliderlabs/ssh"
	"github.com/rivo/tview"
)

func main() {
	ssh.Handle(func(sess ssh.Session) {
		screen, err := NewSessionScreen(sess)
		if err != nil {
			fmt.Fprintln(sess.Stderr(), "unable to create screen:", err)
			return
		}

		// tview says we don't have to do this
		// when using SetScreen, but it lies
		if err := screen.Init(); err != nil {
			fmt.Fprintln(sess.Stderr(), "unable to init screen:", err)
			return
		}

		app := tview.NewApplication().SetScreen(screen).EnableMouse(true)

		modal := tview.NewModal().
			SetText("Do you want to quit the application?").
			AddButtons([]string{"Quit", "Cancel"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Quit" {
					app.Stop()
				}
			})

		app.SetRoot(modal, false)
		if err := app.Run(); err != nil {
			fmt.Fprintln(sess.Stderr(), err)
			return
		}

		sess.Exit(0)
	})

	log.Fatal(ssh.ListenAndServe(":2222", nil, ssh.HostKeyFile("/Users/progrium/.ssh/id_rsa")))
}

func NewSessionScreen(s ssh.Session) (tcell.Screen, error) {
	pi, ch, ok := s.Pty()
	if !ok {
		return nil, errors.New("no pty requested")
	}
	ti, err := terminfo.LookupTerminfo(pi.Term)
	if err != nil {
		return nil, err
	}
	screen, err := tcell.NewTerminfoScreenFromTtyTerminfo(&tty{
		Session: s,
		size:    pi.Window,
		ch:      ch,
	}, ti)
	if err != nil {
		return nil, err
	}
	return screen, nil
}

type tty struct {
	ssh.Session
	size     ssh.Window
	ch       <-chan ssh.Window
	resizecb func()
	mu       sync.Mutex
}

func (t *tty) Start() error {
	go func() {
		for win := range t.ch {
			t.size = win
			t.notifyResize()
		}
	}()
	return nil
}

func (t *tty) Stop() error {
	return nil
}

func (t *tty) Drain() error {
	return nil
}

func (t *tty) WindowSize() (width int, height int, err error) {
	return t.size.Width, t.size.Height, nil
}

func (t *tty) NotifyResize(cb func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.resizecb = cb
}

func (t *tty) notifyResize() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.resizecb != nil {
		t.resizecb()
	}
}
