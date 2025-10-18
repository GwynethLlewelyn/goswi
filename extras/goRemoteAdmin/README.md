# goRemoteAdmin
## A very basic command-line tool to connect via XML-RPC to the remote admin console of OpenSimulator

It's still a work in progress.

Loosely based on Marcel Edward's [OpenSim-RemoteAdmin](https://github.com/MarcelEdward/OpenSim-RemoteAdmin/) package which was written in PHP (which I can read, unlike so many others), and having this cool idea of using a [RemoteAdmin.json](./RemoteAdmin.json) file to store all possible Remote Admin commands!

This is a standalone CLI tool just to do some basic testing.

You can use the 'help' command to get an idea of what commands are available, but here is a complete example:

`goRemoteAdmin -p averydifficulttoguesspassword --host http://my.opensim.server.tld:9150 admin_broadcast message "Look ma, no hands"`

## Future updates

If Marcel ever releases a new version of [`RemoteAdmin.json`](https://github.com/MarcelEdward/OpenSim-RemoteAdmin/blob/master/RemoteAdmin.json), make sure you get it!