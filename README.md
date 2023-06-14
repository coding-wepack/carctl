# carctl

`carctl` is a tool to help you control your CODING Artifact Registry.

The full name of `carctl` is *CODING Artifacts Repository Control*.

Such as `migrate`, you can migrate artifacts from a local directory or a remote url
to a CODING Artifact Repository easily.

`migrate` now supports:
- JFrog Artifactory: `generic`、`docker`、`maven` and `npm`.
- Nexus: `maven`、`pypi` and `composer`.
- Local Repository: `maven`
- Repository settings like proxy source list.

## Installation

### cURL & wget

```shell
# Format: 'https://coding-public-generic.pkg.coding.net/registry/disk/carctl/(linux|darwin|windows)/(amd64|arm64)/carctl?version=latest'
# e.g.,
$ curl -fL 'https://coding-public-generic.pkg.coding.net/registry/disk/carctl/linux/amd64/carctl?version=latest' -o carctl
# or wget
$ wget 'https://coding-public-generic.pkg.coding.net/registry/disk/carctl/linux/amd64/carctl?version=latest' -O carctl
# for MacOS
$ wget 'https://coding-public-generic.pkg.coding.net/registry/disk/carctl/darwin/amd64/carctl?version=latest' -O carctl
# MacOS ARM64
$ wget 'https://coding-public-generic.pkg.coding.net/registry/disk/carctl/darwin/arm64/carctl?version=latest' -O carctl

$ chmod +x carctl
$ sudo mv carctl /usr/local/bin

# validate
$ carctl
```


## Help

```shell
$ carctl

The CODING Artifact Registry Manager

Common actions for carctl:

- carctl login:      login to a CODING Artifact Registry
- carctl logout:     logout from a CODING Artifact Registry
- carctl repo:       handle and control artifact repository
- carctl migrate:    migrate artifacts from local or remote to a CODING Artifact Repository

Usage:
  carctl [command]

Available Commands:
  help        Help about any command
  repo        The repo command can handle and control artifact repository.
  logout      Logout from a registry
  migrate     Migrate artifacts from anywhere to a CODING Artifact Repository.
  repo        The repo command can handle and control artifact repository.
  version     print the CLI version

Flags:
  -h, --help      help for carctl
  -v, --verbose   Make the operation more talkative

Use "carctl [command] --help" for more information about a command.
```

## Commands

### Login

```shell
# input by interactive mode
$ carctl login <registry>
Username:
Password:
```

e.g.,

```shell
$ carctl login team-maven.pkg.coding.net
WARNING: Using --password via the CLI is insecure. Use --password-stdin.
Username: username
Password: 
Login Succeeded

# or
$ carctl login -u username -p password team-maven.pkg.coding.net
WARNING: Using --password via the CLI is insecure. Use --password-stdin.
Login Succeeded

# or
$ carctl login -u username team-maven.pkg.coding.net
WARNING: Using --password via the CLI is insecure. Use --password-stdin.
Password: 
Login Succeeded

# or
$ echo $PASSWORD | carctl login -u username --password-stdin
Login Succeeded
```


### Logout

```shell
$ carctl logout <registry>
```

e.g.,

```shell
$ carctl logout team-maven.pkg.coding.net
Removing login credentials for team-maven.pkg.coding.net
```


### Migrate

#### Maven

Migrate your maven repository to a remote maven repository:

```shell
$ carctl migrate maven --src=/home/juan/.m2/swagger-repository --dst=http://codingcorp-maven.pkg.coding.com/repository/registry/overridable-maven-migrate/   
2021-12-13 16:35:30.067	INFO	Stat source repository ...
2021-12-13 16:35:30.067	INFO	Check authorization of the registry
2021-12-13 16:35:30.067	INFO	Scanning repository ...
2021-12-13 16:35:30.067	INFO	Successfully to scan the repository	{"groups": 3, "artifacts": 16, "versions": 16, "files": 56}
2021-12-13 16:35:30.067	INFO	Begin to migrate ...
Pushing: Done! [==============================================================================] 56 / 56  100 %
2021-12-13 16:35:39.742	INFO	End to migrate.	{"duration": "9.674504559s", "succeededCount": 56, "skippedCount": 0, "failedCount": 0}
```

You can use `-v` or `--verbose` flag to see more info:

```shell
$ carctl migrate maven --src=/home/juan/.m2/swagger-repository --dst=http://codingcorp-maven.pkg.coding.com/repository/registry/overridable-maven-migrate/ -v
2021-12-13 16:33:58.526	INFO	Stat source repository ...
2021-12-13 16:33:58.529	INFO	Check authorization of the registry
2021-12-13 16:33:58.531	DEBUG	Auth config	{"host": "codingcorp-maven.pkg.coding.com", "username": "username", "password": "password"}
2021-12-13 16:33:58.532	INFO	Scanning repository ...
2021-12-13 16:33:58.532	INFO	Successfully to scan the repository	{"groups": 3, "artifacts": 16, "versions": 16, "files": 56}
2021-12-13 16:33:58.532	INFO	Repository Info:
+----------------------+-----------------------------+--------------------+---------------------------------------------+
|       GROUP ID       |         ARTIFACT ID         |      VERSION       |                    FILE                     |
+----------------------+-----------------------------+--------------------+---------------------------------------------+
| io.swagger.core.v3   | swagger-annotations         | 2.1.2              | swagger-annotations-2.1.2.jar               |
+                      +                             +                    +---------------------------------------------+
# --- snip ---
```

Migrate your nexus maven repository to a remote maven repository:

```shell
$ carctl migrate maven --src-type=nexus --src=http://localhost:8081/repository/maven-test/ --src-username=admin --src-password=admin123 --dst=http://codingcorp-maven.pkg.coding.com/repository/registry/overridable-maven-migrate/ 
```
