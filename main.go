package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"satae66.dev/netzeps2022/cli"
	"satae66.dev/netzeps2022/core"
)

type Command interface {
	Name() string

	Init([]string) error
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
	localAddress string
	localPort    int

	maxPacketSize     int
	connectionTimeout int
}

func (cmd *DefaultCommand) SetDefaultFlags(fs *flag.FlagSet) {
	fs.StringVar(&cmd.localAddress, "lAddr", "127.0.0.1", "Listen IP-Address [default = 127.0.0.1]")
	fs.IntVar(&cmd.localPort, "lPort", 6969, "Listen port [default = 6969]")

	fs.IntVar(&cmd.maxPacketSize, "packetSize", 512, "Maximum size of each packet [default = 512]")
	fs.IntVar(&cmd.connectionTimeout, "timeout", 10, "Timeout of the connection in seconds [default = 10]")
}

/*
/----------------------------------------------------------------------------------------------------------------------\
|                                                       SEND-CMD                                                       |
\----------------------------------------------------------------------------------------------------------------------/
*/

type SendCommand struct {
	fs *flag.FlagSet

	DefaultCommand

	destinationAddress string
	destinationPort    int
	filename           string
}

func NewSendCommand() *SendCommand {
	cmd := &SendCommand{
		fs: flag.NewFlagSet("send", flagErrorHandling),
	}

	cmd.fs.StringVar(&cmd.destinationAddress, "rAddr", "localhost", "Remote IP-Address [default = localhost]")
	cmd.fs.IntVar(&cmd.destinationPort, "rPort", 6969, "Remote port [default = 6969]")
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

	cmd.fs.StringVar(&cmd.outDir, "outDir", ".", "The output directory")
	return cmd
}

func (cmd *ReceiveCommand) Name() string {
	return "receive"
}

func (cmd *ReceiveCommand) Init(args []string) error {
	cmd.SetDefaultFlags(cmd.fs)
	return cmd.fs.Parse(args)
}

/*
/----------------------------------------------------------------------------------------------------------------------\
|                                                         MAIN                                                         |
\----------------------------------------------------------------------------------------------------------------------/
*/

var log *bufio.Writer
var errorLog *bufio.Writer

func init() {
	logFile, err := os.Create("measure_log.txt")
	if err != nil {
		panic(err)
	}
	log = bufio.NewWriter(logFile)

	errorFile, err := os.Create("error_log.txt")
	if err != nil {
		panic(err)
	}
	errorLog = bufio.NewWriter(errorFile)
}

func main() {
	var err error

	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Printf("%s", "not enough arguments")
		os.Exit(-1)
	}

	selectedCommand := args[0]
	args = args[1:]

	switch selectedCommand {
	case "send":
		cmd := NewSendCommand()
		err = cmd.Init(args)
		if err != nil {
			fmt.Printf("%v", err)
			os.Exit(-1)
		}
		err = startSender(cmd)
		if err != nil {
			fmt.Printf("%v", err)
			os.Exit(-1)
		}
		break
	case "receive":
		cmd := NewReceiveCommand()
		err = cmd.Init(args)
		if err != nil {
			fmt.Printf("%v", err)
			os.Exit(-1)
		}
		err = startReceiver(cmd)
		if err != nil {
			fmt.Printf("%v", err)
			os.Exit(-1)
		}
	default:
		err = fmt.Errorf("undefined command %q", selectedCommand)
		os.Exit(-1)
	}

	fin := make(chan bool, 1)
	<-fin
}

func startReceiver(cmd *ReceiveCommand) error {
	lIp := cmd.localAddress
	lPort := cmd.localPort
	maxPacketSize := cmd.maxPacketSize
	netTimeout := cmd.connectionTimeout
	outPath := cmd.outDir

	ip := net.ParseIP(lIp)
	if ip == nil {
		return fmt.Errorf("ip %q could not be parsed", lIp)
	}

	lAddr := &net.UDPAddr{
		IP:   ip,
		Port: lPort,
	}

	// Receiver
	//TODO: remove buffer
	r, err := NewReceiver(maxPacketSize, netTimeout, 100000, outPath, lAddr)
	if err != nil {
		return err
	}

	cleaner, err := core.NewTransmissionCleaner(1, 10, &r.transmissions)
	if err != nil {
		return err
	}
	go cleaner.Start()

	// CLI
	ui, err := cli.NewCliWorker(1, &r.transmissions)
	if err != nil {
		return err
	}
	go ui.Start()

	// Receiver
	errorChannel := make(chan error, 10)
	r.Start(errorChannel)
	go func() {
		for {
			_, _ = fmt.Fprintf(errorLog, "%s\n", <-errorChannel)
		}
	}()

	return nil
}

func startSender(cmd *SendCommand) error {
	return fmt.Errorf("%v", "sender not implemented yet")
}
