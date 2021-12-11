# carctl

`carctl` is a tool to help you control your CODING Artifact Registry.

The full name of `carctl` is *CODING Artifacts Repository Control*.

Such as `migrate`, you can migrate artifacts from a local directory or a remote url
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

Migrate local `~/.m2/repository` to remote maven repository:

```shell
$ carctl migrate maven --src=/home/juan/.m2/swagger-core-repository --dst=http://codingcorp-maven.pkg.nh4ivfk.dev.coding.io/repository/registry/swagger-repository/
2021-12-11 21:28:32.136	INFO	Stat source repository ...
2021-12-11 21:28:32.138	INFO	Check authorization of the registry
2021-12-11 21:28:32.143	INFO	Scanning repository ...
2021-12-11 21:28:32.148	INFO	Successfully to scan the repository	{"groups": 1, "artifacts": 1, "versions": 1, "files": 4}
2021-12-11 21:28:32.148	INFO	Begin to migrate ...
2021-12-11 21:28:32.148	INFO	Put file:	{"file": "/home/juan/.m2/swagger-core-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.jar", "url": "http://codingcorp-maven.pkg.nh4ivfk.dev.coding.io/repository/registry/swagger-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.jar"}
2021-12-11 21:28:32.830	INFO	Put file:	{"file": "/home/juan/.m2/swagger-core-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.jar.sha1", "url": "http://codingcorp-maven.pkg.nh4ivfk.dev.coding.io/repository/registry/swagger-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.jar.sha1"}
2021-12-11 21:28:32.970	INFO	Put file:	{"file": "/home/juan/.m2/swagger-core-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.pom", "url": "http://codingcorp-maven.pkg.nh4ivfk.dev.coding.io/repository/registry/swagger-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.pom"}
2021-12-11 21:28:33.198	INFO	Put file:	{"file": "/home/juan/.m2/swagger-core-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.pom.sha1", "url": "http://codingcorp-maven.pkg.nh4ivfk.dev.coding.io/repository/registry/swagger-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.pom.sha1"}
2021-12-11 21:28:33.382	INFO	End to migrate.	{"duration": "1.234179809s", "total": 4, "succeededCount": 4, "failedCount": 0, "skippedCount": 0}
```

You can use `-v` or `--verbose` flag to see more info:

```shell
$ carctl migrate maven --src=/home/juan/.m2/swagger-core-repository --dst=http://codingcorp-maven.pkg.nh4ivfk.dev.coding.io/repository/registry/swagger-core-repository/ -v
2021-12-11 21:31:09.612	INFO	Stat source repository ...
2021-12-11 21:31:09.620	INFO	Check authorization of the registry
2021-12-11 21:31:09.626	DEBUG	Auth config	{"host": "codingcorp-maven.pkg.nh4ivfk.dev.coding.io", "username": "coding-coding", "password": "coding123"}
2021-12-11 21:31:09.626	INFO	Scanning repository ...
2021-12-11 21:31:09.645	INFO	Successfully to scan the repository	{"groups": 1, "artifacts": 1, "versions": 1, "files": 4}
2021-12-11 21:31:09.645	INFO	Repository Info:
+--------------------+--------------------+-------------------+-----------------------------+
|      GROUP ID      |    ARTIFACT ID     |      VERSION      |            FILE             |
+--------------------+--------------------+-------------------+-----------------------------+
| io.swagger.core.v3 | swagger-core       | 2.1.2             | swagger-core-2.1.2.jar      |
+                    +                    +                   +-----------------------------+
|                    |                    |                   | swagger-core-2.1.2.jar.sha1 |
+                    +                    +                   +-----------------------------+
|                    |                    |                   | swagger-core-2.1.2.pom      |
+                    +                    +                   +-----------------------------+
|                    |                    |                   | swagger-core-2.1.2.pom.sha1 |
+--------------------+--------------------+-------------------+-----------------------------+
|  TOTAL GROUPS: 1   | TOTAL ARTIFACTS: 1 | TOTAL VERSIONS: 1 |       TOTAL FILES: 4        |
+--------------------+--------------------+-------------------+-----------------------------+
2021-12-11 21:31:09.645	INFO	Begin to migrate ...
2021-12-11 21:31:09.645	INFO	Put file:	{"file": "/home/juan/.m2/swagger-core-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.jar", "url": "http://codingcorp-maven.pkg.nh4ivfk.dev.coding.io/repository/registry/swagger-core-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.jar"}
2021-12-11 21:31:10.299	INFO	Successfully migrated:	{"file": "/home/juan/.m2/swagger-core-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.jar"}
2021-12-11 21:31:10.299	INFO	Put file:	{"file": "/home/juan/.m2/swagger-core-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.jar.sha1", "url": "http://codingcorp-maven.pkg.nh4ivfk.dev.coding.io/repository/registry/swagger-core-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.jar.sha1"}
2021-12-11 21:31:10.514	INFO	Successfully migrated:	{"file": "/home/juan/.m2/swagger-core-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.jar.sha1"}
2021-12-11 21:31:10.514	INFO	Put file:	{"file": "/home/juan/.m2/swagger-core-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.pom", "url": "http://codingcorp-maven.pkg.nh4ivfk.dev.coding.io/repository/registry/swagger-core-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.pom"}
2021-12-11 21:31:10.731	INFO	Successfully migrated:	{"file": "/home/juan/.m2/swagger-core-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.pom"}
2021-12-11 21:31:10.731	INFO	Put file:	{"file": "/home/juan/.m2/swagger-core-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.pom.sha1", "url": "http://codingcorp-maven.pkg.nh4ivfk.dev.coding.io/repository/registry/swagger-core-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.pom.sha1"}
2021-12-11 21:31:10.878	INFO	Successfully migrated:	{"file": "/home/juan/.m2/swagger-core-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.pom.sha1"}
2021-12-11 21:31:10.878	INFO	End to migrate.	{"duration": "1.232229131s", "total": 4, "succeededCount": 4, "failedCount": 0, "skippedCount": 0}
```

