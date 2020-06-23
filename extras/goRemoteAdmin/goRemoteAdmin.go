/**
*	Sub-project to connect to the OpenSim Remote Admin console via XML-RPC. This ‘extra’ will act as a CLI application 
*	to send commands from the shell to the remote admin console, basically to test things; it uses the excellent idea 
*	from https://github.com/MarcelEdward/OpenSim-RemoteAdmin/ (a PHP tool) which reads all valid XML-RPC commands from a JSON file.
**/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"io/ioutil"
//	"github.com/kolo/xmlrpc"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	var (
		language, remoteAdminFile string
		verboseMode bool
	)

	app := &cli.App{
		EnableBashCompletion: true,
		Flags: []cli.Flag {
			&cli.StringFlag{
				Name:			"lang",
				Value:			"english",
				Usage:			"language for the greeting",
				Destination:	&language,
			},
			&cli.StringFlag{
				Name:			"config",
				Value:			"RemoteAdmin.json",
				Aliases:		[]string{"c"},
				Usage:			"Load JSON directives from `FILE`",
				Destination:	&remoteAdminFile,
			},
			&cli.BoolFlag{
				Name:			"view",
				Value:			false,
				Aliases:		[]string{"w"},
				Usage:			"View result of JSON directives of `FILE` (verbose mode)",
				Destination:	&verboseMode,
				DefaultText:	"boolean",
			},
		},
		Commands: []*cli.Command{
			{
				Name:	 "complete",
				Aliases: []string{"c"},
				Usage:	 "complete a task on the list",
				Action:	 func(c *cli.Context) error {
					sometask := "nothing"
					if c.NArg() > 0 {
						sometask = c.Args().First()
					}
					fmt.Println("completing", sometask)
					return nil
				},
			},
			{
				Name:	 "add",
				Aliases: []string{"a"},
				Usage:	 "add a task to the list",
				Action:	 func(c *cli.Context) error {
					sometask := "nothing"
					if c.NArg() > 0 {
						sometask = c.Args().First()
					}
					fmt.Println("adding", sometask)		
					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			name := "someone"
			if c.NArg() > 0 {
				name = c.Args().First()
			}
			if language == "spanish" {
				fmt.Println("Hola", name)
			} else {
				fmt.Println("Hello", name)
			}
			// Open our jsonFile
			jsonFile, oneErr := os.Open(remoteAdminFile)
			// if we os.Open returns an error then handle it
			if oneErr != nil {
				fmt.Println(oneErr)
			}
			fmt.Println("Successfully Opened ", remoteAdminFile)
			// defer the closing of our jsonFile so that we can parse it later on
			defer jsonFile.Close()
		
			byteValue, _ := ioutil.ReadAll(jsonFile)
		
			var result map[string]interface{}
			json.Unmarshal([]byte(byteValue), &result)
			if verboseMode {
				fmt.Printf("%+v", result) 
			}
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
	log.Fatal(err)
	}
}
