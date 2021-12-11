# carctl

`carctl` is a tool to help you control your CODING Artifact Registry.

The full name of `carctl` is *CODING Artifacts Repository Control*.

Such as `migrate`, you can migrate your artifacts from local or a remote url
to a CODING Artifact Repository easily.


## Help

```shell
$ carctl

The CODING Artifact Registry Manager

Common actions for carctl:

- carctl login:      login to a CODING Artifact Registry
- carctl logout:     logout from a CODING Artifact Registry
- carctl migrate:    migrate artifacts from local or remote to a CODING Artifact Repository
- carctl pull:       pull artifacts from a CODING Artifact Repository to local (TODO)
- carctl push:       push artifacts from local to a CODING Artifact Repository (TODO)
- carctl search:     search for artifacts (TODO)
- carctl list:       list artifacts (TODO)

Usage:
  carctl [command]

Available Commands:
  help        Help about any command
  migrate     Migrate artifacts from anywhere to a CODING Artifact Repository.
  registry    login to or logout from a registry
  version     Print the CLI version

Flags:
  -h, --help      help for carctl
  -v, --verbose   Make the operation more talkative

Use "carctl [command] --help" for more information about a command.
```

## Commands

### Login

```shell
# input by interactive mode
$ carctl registry login <registry>
Username:
Password:
```

e.g.,

```shell
$ carctl registry login team-maven.pkg.coding.net
WARNING: Using --password via the CLI is insecure. Use --password-stdin.
Username: username
Password: 
Login Succeeded

# or
$ carctl registry login -u username -p password team-maven.pkg.coding.net
WARNING: Using --password via the CLI is insecure. Use --password-stdin.
Login Succeeded

# or
$ carctl registry login -u username team-maven.pkg.coding.net
WARNING: Using --password via the CLI is insecure. Use --password-stdin.
Password: 
Login Succeeded

# or
$ echo $PASSWORD | carctl registry login -u username --password-stdin
Login Succeeded
```


### Migrate

### Maven



