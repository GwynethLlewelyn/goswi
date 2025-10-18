/**
*	Sub-project to connect to the OpenSim Remote Admin console via XML-RPC. This ‘extra’ will act as a CLI application
*	to send commands from the shell to the remote admin console, basically to test things; it uses the excellent idea
*	from https://github.com/MarcelEdward/OpenSim-RemoteAdmin/ (a PHP tool) which reads all valid XML-RPC commands from a JSON file.
**/

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"os"
	"reflect"
	"time"

	"github.com/earthboundkid/versioninfo/v2" // mostly to get the git version of this build!
	// "github.com/kolo/xmlrpc"
	_ "github.com/joho/godotenv/autoload" // gets value from OPENSIM_REMOTE_PASSWORD optionally from .env
	"github.com/urfave/cli/v3"            // main CLI engine
)

// No harm is done having just one context, which is simoly the background.
var ctx = context.Background()

type XmlRpcParameter struct {
	Parameter string `xmlrpc:"parameter"`
	Value     string `xmlrpc:"string"`
}

func main() {
	var (
		remoteAdminFile       string = "RemoteAdmin.json"
		verboseMode           bool   = false // just to make sure!!
		opensimServerURL      string
		xmlrpcRawCommandsJSON map[string]any
		password              string
	)

	// Check if we have the remote password on environment.
	// Note that we are using godotenv/autoload to automatically retrieve .env
	// and merge with the existing environment.
	password = os.Getenv("OPENSIM_REMOTE_PASSWORD")

	cmd := &cli.Command{
		Name: os.Args[0], // use whatever the compiled version says it uses
		Authors: []any{
			&mail.Address{Name: "Gwyneth Llewelyn", Address: "gwyneth.llewelyn@gwynethllewelyn.net"},
		},
		//		HelpName: "goRemoteAdmin",	// the default is fine
		Usage: "Access OpenSimulator Remote Admin console via XML-RPC calls",
		UsageText: `Run it from the shell with a few parameters to send commands.
You need to have a copy of a valid RemoteAdmin.json file with all the XML-RPC commands
known to OpenSimulator. You can get a copy from <https://github.com/MarcelEdward/OpenSim-RemoteAdmin/blob/master/RemoteAdmin.json>
(which served as inspiration for this command)
` + func() string {
			if len(password) < 5 {
				return "\nSet `OPENSIM_REMOTE_PASSWORD` to avoid passing it on the command line (inside `.env` is fine)"
			}
			return "\n`OPENSIM_REMOTE_PASSWORD` set to [..." + password[len(password)-4:] + "]"
		}(),
		EnableShellCompletion: true,
		HideVersion:           false,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Value:       "RemoteAdmin.json",
				Aliases:     []string{"c"},
				Usage:       "Load JSON directives from `FILE`",
				Destination: &remoteAdminFile,
			},
			&cli.StringFlag{
				Name:        "host",
				Value:       "http://127.0.0.1:9000",
				Aliases:     []string{"H", "url"},
				Usage:       "`URL` to OpenSimulator instance",
				Destination: &opensimServerURL,
				DefaultText: "localhost, port 9000",
			},
			/*
				&cli.StringFlag{
					Name:			"user",
					Value:			"",
					Aliases:		[]string{"u", "username"},
					Usage:			"`Username` for XML-RPC call",
					Destination:	&username,
					DefaultText:	"none, unsafe!",
				},
			*/
			&cli.StringFlag{
				Name: "password",
				Value: func() string {
					if len(password) < 5 {
						return ""
					}
					return password
				}(),
				Aliases:     []string{"p"},
				Usage:       "Access `password` or 'secret' for XML-RPC call",
				Destination: &password,
				DefaultText: func() string {
					if len(password) < 5 {
						return "none, unsafe — set `OPENSIM_REMOTE_PASSWORD` instead"
					}
					return "set from OPENSIM_REMOTE_PASSWORD=[..." + password[len(password)-4:] + "]"
				}(),
				Action: func(ctx context.Context, c *cli.Command, value string) error {
					if len(password) < 5 {
						return errors.New("empty or too small password")
					}
					return nil
				},
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Value:       false,
				Aliases:     []string{"w", "v", "debug"},
				Usage:       "Verbose/debug mode (shows lots of info)",
				Destination: &verboseMode,
				DefaultText: "false",
			},
			&cli.BoolFlag{
				Name:    "version",
				Aliases: []string{"V"},
				Usage:   "Print the version",
				Action: func(ctx context.Context, c *cli.Command, value bool) error {
					fmt.Printf("VERSION:\n%s\n", versioninfo.Short())
					return nil
				},
			},
		},
		CommandNotFound: func(ctx context.Context, cmd *cli.Command, command string) {
			fmt.Fprintf(cmd.Writer, "The command %q doesn't seem to be valid for OpenSimulator Remote Admin.\n", command)
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if len(os.Args) < 2 {
				return errors.New("empty command line; try --help to get a list of available commands")
			}

			var oneRawCommand, oneRawProperty, oneRawSubProperty map[string]any
			var ok bool
			if verboseMode && xmlrpcRawCommandsJSON != nil {
				//fmt.Printf("%#v", xmlrpcRawCommandsJSON)
				fmt.Println("--- Command list dump below ---")
				for rawCommand, rawCommandData := range xmlrpcRawCommandsJSON {
					fmt.Println("\n\n====================\nCommand", rawCommand, ":\n====================")
					oneRawCommand, ok = rawCommandData.(map[string]any)
					if ok {
						for property, rawData := range oneRawCommand {
							fmt.Println("Property:", property, "\nData:")
							// see if data has more slices/maps inside
							if reflect.TypeOf(rawData).Kind() != reflect.String {
								oneRawProperty, ok = rawData.(map[string]any)
								if ok {
									for subProperty, rawSubPropertyData := range oneRawProperty {
										fmt.Println("\t", subProperty, ">")
										// see if this subproperty has more slices/maps inside... (we go another level deeper!)
										if reflect.TypeOf(rawSubPropertyData).Kind() != reflect.String {
											oneRawSubProperty, ok = rawSubPropertyData.(map[string]any)
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
		if xmlrpcRawCommandsJSON == nil || cmd == nil {
			fmt.Println("Please place the JSON Remote Admin file on the path\n(you can get it from here: https://github.com/MarcelEdward/OpenSim-RemoteAdmin/blob/master/RemoteAdmin.json)\nand try again")
			os.Exit(2)	// ENOENT
		}
	*/

	cmd.Copyright = "Licensed as CC-BY " + formatAsYear(time.Now()) + " by " + cmd.Authors[0].(*mail.Address).Name + ". Few rights reserved."

	// add xmlrpc commands on demand!!
	if remoteAdminFile == "" {
		if verboseMode {
			fmt.Println("Empty JSON config file name, file not loaded, and no XML-RPC commands will be available")
		}
	} else {
		// Open our jsonFile
		// if we os.Open returns an error then handle it
		if jsonFile, err := os.Open(remoteAdminFile); err == nil {
			if verboseMode {
				fmt.Println("Successfully Opened", remoteAdminFile)
			}
			// defer the closing of our jsonFile so that we can parse it later on
			defer jsonFile.Close()

			if byteValue, err := io.ReadAll(jsonFile); err == nil {
				json.Unmarshal([]byte(byteValue), &xmlrpcRawCommandsJSON)
			} else {
				if verboseMode {
					fmt.Println("Could not load", remoteAdminFile)
				}
			}
		} else {
			if verboseMode {
				fmt.Println("Could not open file", remoteAdminFile)
			}
		}

		// now try to add the commands to the App
		if xmlrpcRawCommandsJSON == nil {
			fmt.Println("No JSON config file specified or file invalid; limited functionality applies")
		} else {
			//		First attempt: just add commands, if they don't work, tough!
			//		var oneRawCommand, oneRawProperty, oneRawSubProperty map[string]interface{}
			var usage string
			if verboseMode {
				fmt.Println("--- Loading commands ---")
			}
			for rawCommand, rawCommandData := range xmlrpcRawCommandsJSON {
				// there should be one
				oneRawCommand, ok := rawCommandData.(map[string]any)
				if !ok {
					if verboseMode {
						fmt.Print("x")
					}
					break // try next command...
				}

				if rawUsage, ok := oneRawCommand["description"]; ok {
					usage = fmt.Sprintf("%v", rawUsage)
				} else {
					usage = "(no description)"
				}

				var newCommand cli.Command
				newCommand = cli.Command{
					Name:  rawCommand,
					Usage: usage,
					Action: func(ctx context.Context, cmd *cli.Command) error {
						fmt.Println("Sending", cmd.Name, "with", cmd.Args(), "to", opensimServerURL)
						if !cmd.Args().Present() {
							return fmt.Errorf("No arguments found for command %q - aborting!\n", rawCommand)
						}

						// Build the XML manually, since it's too hard to do it using a library...
						var xmlrpcRequest bytes.Buffer
						xmlrpcRequest.WriteString(`<?xml version="1.0"?>
<methodCall>
	<methodName>` + cmd.Name + `</methodName>
		<params>
			<param>
				<value>
					<struct>
						<member>
							<name>password</name>
							<value><string>` + password + `</string></value>
						</member>`)
						for i := 0; i < cmd.NArg()-1; i += 2 {
							xmlrpcRequest.WriteString("\n\t\t\t\t\t\t<member>\n\t\t\t\t\t\t\t<name>" + cmd.Args().Get(i) + "</name>\n")
							xmlrpcRequest.WriteString("\t\t\t\t\t\t\t<value>" + cmd.Args().Get(i+1) + "</value>\n\t\t\t\t\t\t</member>\n")
						}
						xmlrpcRequest.WriteString(`					</struct>
				</value>
			</param>
		</params>
</methodCall>`)
						//							return fmt.Errorf("Could not initialise buffer for request; error was: %q\n", err.Error())

						// should do some sanitation here

						fmt.Printf("Request: %v\n", xmlrpcRequest.String())

						var client http.Client

						if req, err := http.NewRequest("POST", opensimServerURL, &xmlrpcRequest); err != nil {
							return fmt.Errorf("Could not initialise request with error: %q\n", err.Error())
						} else {
							req.Header.Add("Content-type", "text/xml")
							req.Header.Add("Connection", "close")
							resp, err := client.Do(req)

							fmt.Printf("Response: %v\n", resp)
							// decode resp to get valid XML, etc.
							return err
						}
						/*
							// It's hard to extract the parameters to do what I want! This is using "github.com/kolo/xmlrpc"
							//  loop through parameter/value pairs
							var xmlrpcRequest []XmlRpcParameter
							// add password first
							xmlrpcRequest = append(xmlrpcRequest, XmlRpcParameter{
									Parameter:	"password",
									Value:		password,
								})
							for i := 2; i < c.NArg() - 1; i+=2 {
								xmlrpcRequest = append(xmlrpcRequest, XmlRpcParameter{
									Parameter:	c.Args().Get(i),
									Value:		c.Args().Get(i+1),
								})
							}

							fmt.Printf("Request: %v\n", xmlrpcRequest)
							if client, err := xmlrpc.NewClient(opensimServerURL, nil); err != nil {
								return fmt.Errorf("Could not create a new client to %q - aborting!\n", opensimServerURL)
							} else {
								var response xmlrpc.Response
								err := client.Call(c.Command.Name, &xmlrpcRequest, &response)
								fmt.Printf("Response: %v\n", response)
								return err
							}
							return nil
						*/
					},
				}

				cmd.Commands = append(cmd.Commands, &newCommand)
				if verboseMode {
					fmt.Print(".")
				}
			}
			if verboseMode {
				fmt.Println("\nAll found commands loaded.")
			}
		}
	}

	if err := cmd.Run(ctx, os.Args); err != nil {
		fmt.Println(err)
		os.Exit(22) // EINVAL
	}
}

// formatAsYear extracts the year drom a string and returns it as a valid string.
func formatAsYear(t time.Time) string {
	year, _, _ := t.Date()
	return fmt.Sprintf("%d", year)
}
