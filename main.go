package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
	"snmpwalk/snmp"
	"strconv"
	"time"
)

type cfg struct {
	Version   string
	Community string
	Timeout   int
	Repeats   int
}

var Cfg cfg

func init() {
}

func main() {
	app()
}

func app() {

	app := &cli.App{
		Name:      "snmpt",
		Usage:     "Cli tool for work with SNMP",
		UsageText: "./snmpt [global options] command [command options] [arguments...]",
		ArgsUsage: "",
		Version:   "0.0.1",
		Commands: []*cli.Command{
			walkCommand(),
			getCommand(),
			setCommand(),
			walkBulkCommand(),
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "snmpversion",
				Usage:       "Snmp version. Variants: v1, v2c",
				Value:       "2c",
				Destination: &Cfg.Version,
				Aliases:     []string{"sv"},
			},
			&cli.StringFlag{
				Name:        "community",
				Usage:       "Community",
				Value:       "public",
				Destination: &Cfg.Community,
				Aliases:     []string{"c"},
			},
			&cli.IntFlag{
				Name:        "timeout",
				Value:       3,
				Usage:       "Timeout of snmp request",
				Destination: &Cfg.Timeout,
				Aliases:     []string{"t"},
			},
			&cli.IntFlag{
				Name:        "repeats",
				Value:       5,
				Usage:       "Timeout of snmp repeats",
				Destination: &Cfg.Repeats,
				Aliases:     []string{"r"},
			},
		},
		BashComplete: nil,
		Before:       nil,
		After:        nil,
		Action:       nil,
		CommandNotFound: func(c *cli.Context, command string) {
			fmt.Fprintf(c.App.Writer, "Command %q not found.\nType ./snmpt --help for list all supported commands\n", command)
		},
		OnUsageError:           nil,
		Compiled:               time.Time{},
		Copyright:              "SNMP tool for switcher-core",
		CustomAppHelpTemplate:  "",
		UseShortOptionHandling: false,
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf(`%v
`, err.Error())
		os.Exit(1)
	}
}

func getCommand() *cli.Command {
	return &cli.Command{
		Name:        "get",
		Usage:       "get",
		Description: "send get request",
		HelpName:    "get",
		Action: func(c *cli.Context) error {
			IP := c.Args().Get(0)
			err, SNMP := getSNMP(IP)
			if IP == "" {
				return fmt.Errorf("IP is required argument for walk")
			}
			if err != nil {
				return err
			}
			oid := c.Args().Get(1)
			if oid == "" {
				return fmt.Errorf("OID is required argument for walk")
			}
			err, response := SNMP.Get(c.Args().Get(1))
			if err != nil {
				return err
			}
			bts, err := json.Marshal(response)
			if err != nil {
				return err
			}
			bts, _ = prettyprint(bts)
			fmt.Println(string(bts))
			return nil
		},
	}
}

func setCommand() *cli.Command {
	return &cli.Command{
		Name:        "set",
		Usage:       "set",
		Description: "send set request",
		HelpName:    "set",
		Action: func(c *cli.Context) error {
			IP := c.Args().Get(0)
			err, SNMP := getSNMP(IP)
			if IP == "" {
				return fmt.Errorf("IP is required argument for walk")
			}
			if err != nil {
				return err
			}
			oid := c.Args().Get(1)
			if oid == "" {
				return fmt.Errorf("OID is required argument for walk")
			}
			tp := c.Args().Get(2)
			var value interface{}
			var typeFormated string
			if tp == "" {
				return fmt.Errorf("Type is required for set")
			}
			switch tp {
			case "i":
				intVar, err := strconv.Atoi(c.Args().Get(3))
				if err != nil {
					return fmt.Errorf("Incorrect set value. Must be a number")
				}
				value = intVar
				typeFormated = "Integer"
				break
			case "s":
				value = c.Args().Get(3)
				typeFormated = "OctetString"
				break
			default:
				return fmt.Errorf("Only supported type Integer(i) and OctetString(s)")
			}

			err, response := SNMP.Set(c.Args().Get(1), typeFormated, value)
			if err != nil {
				return err
			}
			bts, err := json.Marshal(response)
			if err != nil {
				return err
			}
			bts, _ = prettyprint(bts)
			fmt.Println(string(bts))
			return nil
		},
	}
}

func walkBulkCommand() *cli.Command {
	return &cli.Command{
		Name:        "bulk-walk",
		Usage:       "bulk-walk",
		Description: "send bulk-walk request",
		HelpName:    "bulk-walk",
		Action: func(c *cli.Context) error {
			IP := c.Args().Get(0)
			err, SNMP := getSNMP(IP)
			if IP == "" {
				return fmt.Errorf("IP is required argument for walk")
			}
			if err != nil {
				return err
			}
			oid := c.Args().Get(1)
			if oid == "" {
				return fmt.Errorf("OID is required argument for walk")
			}
			err, response := SNMP.WalkBulk(c.Args().Get(1))
			if err != nil {
				return err
			}
			bts, err := json.Marshal(response)
			if err != nil {
				return err
			}
			bts, _ = prettyprint(bts)
			fmt.Println(string(bts))
			return nil
		},
	}
}

func walkCommand() *cli.Command {
	return &cli.Command{
		Name:        "walk",
		Usage:       "walk",
		Description: "send walk request",
		HelpName:    "walk",
		Action: func(c *cli.Context) error {
			IP := c.Args().Get(0)
			err, SNMP := getSNMP(IP)
			if IP == "" {
				return fmt.Errorf("IP is required argument for walk")
			}
			if err != nil {
				return err
			}
			oid := c.Args().Get(1)
			if oid == "" {
				return fmt.Errorf("OID is required argument for walk")
			}
			err, response := SNMP.Walk(c.Args().Get(1))
			if err != nil {
				return err
			}
			bts, err := json.Marshal(response)
			if err != nil {
				return err
			}
			bts, _ = prettyprint(bts)
			fmt.Println(string(bts))
			return nil
		},
	}
}

func getSNMP(ip string) (error, *snmp.Snmp) {
	var vers snmp.SnmpVersion
	switch Cfg.Version {
	case "1":
		vers = snmp.Version1
		break
	case "2c":
		vers = snmp.Version2c
		break
	}
	return snmp.Connect(snmp.InitStruct{
		Version:    vers,
		TimeoutSec: time.Duration(Cfg.Timeout) * time.Second,
		Repeats:    Cfg.Repeats,
		Ip:         ip,
		Community:  Cfg.Community,
	})
}

//dont do this, see above edit
func prettyprint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}
