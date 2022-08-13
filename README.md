# tview-ssh
Example using [tcell](https://github.com/gdamore/tcell)+[tview](https://github.com/rivo/tview) over SSH using [gliderlabs/ssh](https://github.com/gliderlabs/ssh) without allocating a PTY or creating a subprocess. There is a little bit of glue, but maybe not enough for a library? Plus it's probably incomplete. Here is what it looks like to make an SSH server that shows a modal when you connect:

```golang
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
```

If you try this, change the hostkey file to something that works for you.