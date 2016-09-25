
Pushtart - The worlds easiest PaaS.
=======================================

Pushtart runs persistantly on any *nix box you own. You `git push` projects to it, and pushtart saves them on the box and runs the repositories' `startup.sh`. Its that simple!

There is also a simple (but fully featured) user management system, as well as the ability to set environment variables on each of your deployments (keep sensitive information out of your repositories!).

## Getting started

To get started, we need to build the program, run `pushtart make-config` to generate a configuration file, and setup a user with `pushtart make-user`. Then we are ready to run!

Assuming you have Go >1.6 installed on your system:

1. `git clone https://github.com/twitchyliquid64/pushtart`
2. `cd pushtart`
3. `export GOPATH=$PWD`
4. `go build`
5. `./pushtart make-config`
6. `./pushtart make-user --username bob --password hi --allow-ssh-password yes`
7. `./pushtart run`

You are now ready to `git push` your projects!

```
git remote add pushtart_server ssh://localhost:2022/test
git push pushtart_server master
```
:DD

#### Management interface

You can actually SSH into pushtart, and run the same commands you can on the unix command line (except make-config, and importing a SSH key).

`ssh <hostname> -p 2022`

Of course, you can change the port and bind-host in the configuration file.

#### Setting up a SSH key

Most sane people prefer to use SSH keys instead of passwords. To setup a key with an existing user, simply run this command before you start the server:

`./pushtart import-ssh-key --username <insert-pushtart-username-here> --pub-key-file ~/.ssh/id_rsa.pub `

If you need to add a user's SSH key when the server is running try this while logged into that computer as that user:

`cat ~/.ssh/id_rsa.pub | ssh <put-server-address-here> -p 2022 import-ssh-key --username <put-username-here>`

#### Terminology

Everything makes sense, except I decided to call a repository in pushtart a 'tart' :). Tarts can be in running or stopped states - this is all controllable through commands.

### USAGE

```
USAGE: pushtart <command> [--config <config file>] [command-specific-arguments...]
if no config file is specified, config.json will be used.
SSH server keys, user information, tart status, and other (normally external) information is stored in the config file.
Commands:
	run (Not available from SSH shell)
	make-config (Not available from SSH shell)
	import-ssh-key --username <username> [--pub-key-file <path-to-.pub-file>] (Not available from SSH shell)
	make-user --username <username [--password <password] [--name <name] [--allow-ssh-password yes/no]
	edit-user --username <username [--password <password] [--name <name] [--allow-ssh-password yes/no]
	ls-users
	ls-tarts
	start-tart --tart <pushURL>
	stop-tart --tart <pushURL>
	edit-tart --tart <pushURL>[--name <name>] [--set-env "<name>=<value>"] [--delete-env <name>]
```


### tartconfig files

Do you hate using pushtarts command line to perform configuration after you `git push`? You don't have to!

Keep all  configuration related to a tart within the tart's repository. Create and commit a file named `tartconfig`, where every line is a pushtart command. The commands in your file will be executed every push!


For example:

```
edit-tart --name "Test Tart"
extension --extension DNSServ --operation set-record --type A --domain crap.com --address 192.168.1.1 --ttl 100
```

_NB: You don't have to specify which tart (ie: --tart <pushURL>) like you do on the command line._

### Extensions

Continuing with the theme of making personal projects easier to develop and ship, there are a number of additional services available within pushtart which are technically out-of-scope, but exist for convienence.

Extensions cannot be turned on/off while the server is running, they _must_ be enabled/disabled before the server is started using the commands below.

#### DNSServ

When enabled, DNSServ provides a simple DNS server. Records can be managed via commands, or be automatically added by a tart (see documentation about the tartconfig file).

To enable DNSServ: `./pushtart extension --extension DNSServ --operation enable`

DNSServ can also act as an upstream (caching) DNS server - That way you can use it as your nameserver! To enable, run: `./pushtart extension --extension DNSServ --operation enable-recursion`

#### Managing records manually

```
./pushtart extension --extension DNSServ --operation set-record --type A --domain crap.com --address 192.168.1.1 --ttl 100
./pushtart extension --extension DNSServ --operation delete-record --domain crap.com
```


### TODO

 - [x] Lock configuration file (.lock file? when pushtart is running)
 - [x] Implement way to load a users ssh key when the server is running
 - [x] Implement access controls to prevent different users from touching tarts they didnt create
 - [x] Logging tart output to console
 - [x] Implement a live log using `ssh <server> logs`
 - [x] Implement a basic DNSserv extension to allow referencing tarts cleanly - maybe even make it a caching DNS server?
 - [ ] Implement a config file in the tart to allow it to specify its own config
 - [ ] Mechanism to set normal config parameters from the commandline (using reflection?)
 - [ ] Prevent tart management commands for tarts a user doesnt own
