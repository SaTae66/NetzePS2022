package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"satae66.dev/netzeps2022/cli"
)

type Command interface {
	Name() string

	Init([]string) error
	Exec() error
}

/*
/----------------------------------------------------------------------------------------------------------------------\
|                                                      CONSTANTS                                                       |
\----------------------------------------------------------------------------------------------------------------------/
*/

const flagErrorHandling = flag.ContinueOnError

/*
/----------------------------------------------------------------------------------------------------------------------\
|                                                     DEFAULT-CMD                                                      |
\----------------------------------------------------------------------------------------------------------------------/
*/

type DefaultCommand struct {
	destinationAddress string
	destinationPort    int

	localAddress string
	localPort    int

	maxPacketSize     int
	connectionTimeout int

	debug bool
}

func (cmd *DefaultCommand) SetDefaultFlags(fs *flag.FlagSet) {
	fs.StringVar(&cmd.destinationAddress, "rAddr", "localhost", "Remote IP-Address [default = localhost]")
	fs.IntVar(&cmd.destinationPort, "rPort", 6969, "Remote port [default = 6969]")

	fs.StringVar(&cmd.localAddress, "lAddr", "localhost", "Local IP-Address [default = localhost]")
	fs.IntVar(&cmd.localPort, "lPort", 6969, "Local port [default = 6969]")

	fs.IntVar(&cmd.maxPacketSize, "packetSize", 512, "Maximum size of a single packet [default = 512]")
	fs.IntVar(&cmd.connectionTimeout, "timeout", 10, "LastUpdated of the connection in seconds [default = 10]")

	fs.BoolVar(&cmd.debug, "debug", false, "Toggle debug mode (dumps log for every transmission)")
}

/*
/----------------------------------------------------------------------------------------------------------------------\
|                                                       SEND-CMD                                                       |
\----------------------------------------------------------------------------------------------------------------------/
*/

type SendCommand struct {
	fs *flag.FlagSet

	DefaultCommand
	filename string
}

func NewSendCommand() *SendCommand {
	cmd := &SendCommand{
		fs: flag.NewFlagSet("send", flagErrorHandling),
	}
	cmd.fs.StringVar(&cmd.filename, "filename", "", "The file to send")
	return cmd
}

func (cmd *SendCommand) Name() string {
	return "send"
}

func (cmd *SendCommand) Init(args []string) error {
	cmd.SetDefaultFlags(cmd.fs)

	err := cmd.fs.Parse(args)
	if err != nil {
		return err
	}

	if cmd.filename == "" {
		return errors.New("no file specified")
	}

	return nil
}

func (cmd *SendCommand) Exec() error {
	return nil
}

/*
/----------------------------------------------------------------------------------------------------------------------\
|                                                     RECEIVE-CMD                                                      |
\----------------------------------------------------------------------------------------------------------------------/
*/

type ReceiveCommand struct {
	fs *flag.FlagSet

	DefaultCommand
	outDir string
}

func NewReceiveCommand() *ReceiveCommand {
	cmd := &ReceiveCommand{
		fs: flag.NewFlagSet("receive", flagErrorHandling),
	}
	cmd.fs.StringVar(&cmd.outDir, "outDir", "", "The output directory")
	return cmd
}

func (cmd *ReceiveCommand) Name() string {
	return "receive"
}

func (cmd *ReceiveCommand) Init(args []string) error {
	cmd.SetDefaultFlags(cmd.fs)
	return cmd.fs.Parse(args)
}

func (cmd *ReceiveCommand) Exec() error {
	return nil
}

/*
/----------------------------------------------------------------------------------------------------------------------\
|                                                         MAIN                                                         |
\----------------------------------------------------------------------------------------------------------------------/
*/

func handleCommand(args []string) error {
	if len(args) < 1 {
		return errors.New("no command specified")
	}

	commands := []Command{
		NewSendCommand(),
		NewReceiveCommand(),
	}

	selectedCommand := args[0]

	for _, cmd := range commands {
		if cmd.Name() == selectedCommand {
			err := cmd.Init(args[1:])
			if err != nil {
				return err
			}
			return cmd.Exec()
		}
	}

	return fmt.Errorf("unknown command: %s", selectedCommand)
}

var log *bufio.Writer

func main() {
	logFile, err := os.OpenFile("go_log.txt", os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	log = bufio.NewWriter(logFile)

	remoteAddr := &net.UDPAddr{
		IP: net.ParseIP("localhost"),
		//IP:   net.ParseIP("10.168.239.159"),
		Port: 6969,
		Zone: "",
	}

	// Receiver
	r, err := NewReceiver(1406, 10, 100000, "./down/", remoteAddr)
	if err != nil {
		panic(err)
	}

	/*
		cleaner, err := core.NewTransmissionCleaner(1, 10, &r.transmissions)
		if err != nil {
			panic(err)
		}
		go cleaner.Start()

	*/

	// CLI
	x, err := cli.NewCliWorker(1, &r.transmissions)
	if err != nil {
		panic(err)
	}

	commandIn := make(chan string, 10)
	go x.Start(commandIn)
	go func() {
		for {
			//TODO: handle commands
			<-commandIn
		}
	}()

	// Receiver
	errorChannel := make(chan error, 10)
	r.Start(errorChannel)
	go func() {
		for {
			fmt.Printf("%s\n", <-errorChannel)
		}
	}()

	t, err := NewTransmitter(1406, 10)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 10; i++ {
		go func() {
			f, err := os.Open("file.test")
			if err != nil {
				panic(err)
			}
			err = t.SendFileTo(f, remoteAddr)
			if err != nil {
				panic(err)
			}
		}()
	}

	fin := make(chan bool, 1)
	<-fin
}
