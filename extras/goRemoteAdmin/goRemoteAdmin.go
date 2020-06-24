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
	"github.com/kolo/xmlrpc"
	"github.com/urfave/cli/v2"
	"os"
	"reflect"
	"time"
)

func main() {
	var (
		remoteAdminFile string = "RemoteAdmin.json"
		verboseMode bool = false // just to make sure!!
		opensimServerURL string
		xmlrpcRawCommandsJSON map[string]interface{}
		username, password string
	)

	// I have no idea why this isn't the default! Note: it doesn't work anyway
	cli.VersionFlag = &cli.BoolFlag{
		Name:			"version",
		Aliases: 		[]string{"V"},
		Usage:			"Print the version",
	}

	app := &cli.App{
//		Name: "goRemoteAdmin",	// the default is fine
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Gwyneth Llewelyn",
				Email: "gwyneth.llewelyn@gwynethllewelyn.net",
			},
		},
//		HelpName: "goRemoteAdmin",	// the default is fine
		Usage: "Access OpenSimulator Remote Admin console via XML-RPC calls",
		UsageText: "Run it from the shell with a few parameters to send commands.\nYou need to have a copy of a valid RemoteAdmin.json file with all the XML-RPC commands\nknown to OpenSimulator. You can get a copy from https://github.com/MarcelEdward/OpenSim-RemoteAdmin/blob/master/RemoteAdmin.json\n(which served as inspiration for this command)",
		EnableBashCompletion: true,
		HideVersion: false,
		Flags: []cli.Flag {
			&cli.StringFlag{
				Name:			"config",
				Value:			"RemoteAdmin.json",
				Aliases:		[]string{"c"},
				Usage:			"Load JSON directives from `FILE`",
				Destination:	&remoteAdminFile,
			},
			&cli.StringFlag{
				Name:			"host",
				Value:			"http://127.0.0.1:9000",
				Aliases:		[]string{"H"},
				Usage:			"`URL` to OpenSimulator instance",
				Destination:	&opensimServerURL,
				DefaultText:	"localhost, port 9000",
			},
			&cli.StringFlag{
				Name:			"user",
				Value:			"",
				Aliases:		[]string{"u", "username"},
				Usage:			"`Username` for XML-RPC call",
				Destination:	&username,
				DefaultText:	"none, unsafe!",
			},
			&cli.StringFlag{
				Name:			"password",
				Value:			"",
				Aliases:		[]string{"p", "pass", "secret"},
				Usage:			"`Password` or 'secret' for XML-RPC call",
				Destination:	&username,
				DefaultText:	"none, unsafe!",
			},
			&cli.BoolFlag{
				Name:			"verbose",
				Value:			false,
				Aliases:		[]string{"w", "v", "debug"},
				Usage:			"Verbose/debug mode (shows lots of info)",
				Destination:	&verboseMode,
				DefaultText:	"boolean",
			},
		},
		CommandNotFound: func(c *cli.Context, command string) {
			fmt.Fprintf(c.App.Writer, "The command %q doesn't seem to be valid for OpenSimulator Remote Admin.\n", command)
		},
		Action: func(c *cli.Context) error {
			var oneRawCommand, oneRawProperty, oneRawSubProperty map[string]interface{}
			var ok bool
			if verboseMode && xmlrpcRawCommandsJSON != nil {
				//fmt.Printf("%#v", xmlrpcRawCommandsJSON)
				fmt.Println("--- Command list dump below ---")
				for rawCommand, rawCommandData := range xmlrpcRawCommandsJSON {
				   	fmt.Println("\n\n====================\nCommand", rawCommand, ":\n====================")
				   	oneRawCommand, ok = rawCommandData.(map[string]interface{})
				   	if ok {
						for property, rawData := range oneRawCommand {
							fmt.Println("Property:", property, "\nData:")
							// see if data has more slices/maps inside
							if reflect.TypeOf(rawData).Kind() != reflect.String {
								oneRawProperty, ok = rawData.(map[string]interface{})
								if ok {
									for subProperty, rawSubPropertyData := range oneRawProperty {
										fmt.Println("\t", subProperty, ">")
										// see if this subproperty has more slices/maps inside... (we go another level deeper!)
										if reflect.TypeOf(rawSubPropertyData).Kind() != reflect.String {
											oneRawSubProperty, ok = rawSubPropertyData.(map[string]interface{})
											if ok {
												for subSubProperty, rawSubSubPropertyData := range oneRawSubProperty {
													fmt.Println("\t\t", subSubProperty, ":>", rawSubSubPropertyData)
												}
											}			
										} else {
											fmt.Println("\t\t", rawSubPropertyData)
										}
									}
								}
							} else { // rawData is just a string
								fmt.Println("\t", rawData)
							}
							fmt.Println("")
						} 
					} else {
						fmt.Println("(couldn't extract properties)")
					}
				}
			}
			return nil
		},
	}
	
/*
	if xmlrpcRawCommandsJSON == nil || app == nil {
		log.Fatal("Please place the JSON Remote Admin file on the path\n(you can get it from here: https://github.com/MarcelEdward/OpenSim-RemoteAdmin/blob/master/RemoteAdmin.json)\nand try again")	
} */

	app.Copyright = "Licensed as CC-BY " + formatAsYear(time.Now()) + " by " + app.Authors[0].String() + ". Few rights reserved."

	// add xmlrpc commands on demand!!
	if remoteAdminFile == "" {
		if verboseMode { fmt.Println("Empty JSON config file name, file not loaded, and no XML-RPC commands will be available") }
	} else {
			// Open our jsonFile
			// if we os.Open returns an error then handle it
			if jsonFile, err := os.Open(remoteAdminFile); err == nil {
				if verboseMode { fmt.Println("Successfully Opened", remoteAdminFile) }
				// defer the closing of our jsonFile so that we can parse it later on
				defer jsonFile.Close()
			
				if byteValue, err := ioutil.ReadAll(jsonFile); err == nil {
					json.Unmarshal([]byte(byteValue), &xmlrpcRawCommandsJSON)
				} else {
					if verboseMode { fmt.Println("Could not load", remoteAdminFile) }
				}
			} else {
				if verboseMode { fmt.Println("Could not open file", remoteAdminFile) }
			}
		
		// now try to add the commands to the App
		if xmlrpcRawCommandsJSON == nil {
			fmt.Println("No JSON config file specified or file invalid; limited functionality applies")
		} else {
	//		First attempt: just add commands, if they don't work, tough!
	//		var oneRawCommand, oneRawProperty, oneRawSubProperty map[string]interface{}
			var usage string
			if verboseMode { fmt.Println("--- Loading commands ---") }
			for rawCommand, rawCommandData := range xmlrpcRawCommandsJSON {
				// there should be one
				oneRawCommand, ok := rawCommandData.(map[string]interface{})	
				if !ok {
					if verboseMode { fmt.Print("x") }
					break	// try next command...
				}
				
				if rawUsage, ok := oneRawCommand["description"]; ok {
					usage = fmt.Sprintf("%v", rawUsage)
				} else {
					usage = "(no description)"
				}
				
				var newCommand cli.Command
				newCommand = cli.Command{
					Name:	 rawCommand,
					Usage:	 usage,
					Action:	 func(c *cli.Context) error {
						fmt.Println("Sending", rawCommand, "with", c.Args(), "to", opensimServerURL)
						client, _ := xmlrpc.NewClient(opensimServerURL, nil)
						result := struct{
							Version string `xmlrpc:"version"`
						}{}
						client.Call(rawCommand, nil, &result)
						fmt.Printf("Result: %v\n", result) // Version: 4.2.7+
						return nil
					},
				}
				
				app.Commands = append(app.Commands, &newCommand)
				if verboseMode { fmt.Print(".") }
			}
			if verboseMode { fmt.Println("\nAll found commands loaded.") }
		}
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// formatAsYear extracts the year drom a string and returns it as a valid string.
func formatAsYear(t time.Time) string {
	year, _, _ := t.Date()
	return fmt.Sprintf("%d", year)
}
