# carctl

`carctl` is a tool to help you control your CODING Artifact Registry.

The full name of `carctl` is *CODING Artifacts Repository Control*.

Such as `migrate`, you can migrate artifacts from a local directory or a remote url
to a CODING Artifact Repository easily.


## Installation

### cURL & wget

```shell
$ curl -fL 'https://coding-public-generic.pkg.coding.net/registry/disk/carctl/linux/amd64/carctl?version=latest' -o carctl
# or wget
$ wget 'https://coding-public-generic.pkg.coding.net/registry/disk/carctl/linux/amd64/carctl?version=latest' -O carctl

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

Migrate your maven repository to a remote maven repository:

```shell
$ carctl migrate maven --src=/home/juan/.m2/swagger-repository --dst=http://codingcorp-maven.pkg.nh4ivfk.dev.coding.io/repository/registry/overridable-maven-migrate/   
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
$ carctl migrate maven --src=/home/juan/.m2/swagger-repository --dst=http://codingcorp-maven.pkg.nh4ivfk.dev.coding.io/repository/registry/overridable-maven-migrate/ -v
2021-12-13 16:33:58.526	INFO	Stat source repository ...
2021-12-13 16:33:58.529	INFO	Check authorization of the registry
2021-12-13 16:33:58.531	DEBUG	Auth config	{"host": "codingcorp-maven.pkg.nh4ivfk.dev.coding.io", "username": "coding-coding", "password": "coding123"}
2021-12-13 16:33:58.532	INFO	Scanning repository ...
2021-12-13 16:33:58.532	INFO	Successfully to scan the repository	{"groups": 3, "artifacts": 16, "versions": 16, "files": 56}
2021-12-13 16:33:58.532	INFO	Repository Info:
+----------------------+-----------------------------+--------------------+---------------------------------------------+
|       GROUP ID       |         ARTIFACT ID         |      VERSION       |                    FILE                     |
+----------------------+-----------------------------+--------------------+---------------------------------------------+
| io.swagger.core.v3   | swagger-annotations         | 2.1.2              | swagger-annotations-2.1.2.jar               |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-annotations-2.1.2.jar.sha1          |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-annotations-2.1.2.pom               |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-annotations-2.1.2.pom.sha1          |
+                      +-----------------------------+                    +---------------------------------------------+
|                      | swagger-core                |                    | swagger-core-2.1.2.jar                      |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-core-2.1.2.jar.sha1                 |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-core-2.1.2.pom                      |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-core-2.1.2.pom.sha1                 |
+                      +-----------------------------+                    +---------------------------------------------+
|                      | swagger-models              |                    | swagger-models-2.1.2.jar                    |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-models-2.1.2.jar.sha1               |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-models-2.1.2.pom                    |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-models-2.1.2.pom.sha1               |
+                      +-----------------------------+                    +---------------------------------------------+
|                      | swagger-project             |                    | swagger-project-2.1.2.pom                   |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-project-2.1.2.pom.sha1              |
+----------------------+-----------------------------+--------------------+---------------------------------------------+
| io.swagger.parser.v3 | swagger-parser              | 2.0.20             | swagger-parser-2.0.20.jar                   |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-parser-2.0.20.jar.sha1              |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-parser-2.0.20.pom                   |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-parser-2.0.20.pom.sha1              |
+                      +-----------------------------+                    +---------------------------------------------+
|                      | swagger-parser-core         |                    | swagger-parser-core-2.0.20.jar              |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-parser-core-2.0.20.jar.sha1         |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-parser-core-2.0.20.pom              |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-parser-core-2.0.20.pom.sha1         |
+                      +-----------------------------+                    +---------------------------------------------+
|                      | swagger-parser-project      |                    | swagger-parser-project-2.0.20.pom           |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-parser-project-2.0.20.pom.sha1      |
+                      +-----------------------------+                    +---------------------------------------------+
|                      | swagger-parser-v2-converter |                    | swagger-parser-v2-converter-2.0.20.jar      |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-parser-v2-converter-2.0.20.jar.sha1 |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-parser-v2-converter-2.0.20.pom      |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-parser-v2-converter-2.0.20.pom.sha1 |
+                      +-----------------------------+                    +---------------------------------------------+
|                      | swagger-parser-v3           |                    | swagger-parser-v3-2.0.20.jar                |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-parser-v3-2.0.20.jar.sha1           |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-parser-v3-2.0.20.pom                |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-parser-v3-2.0.20.pom.sha1           |
+----------------------+-----------------------------+--------------------+---------------------------------------------+
| io.swagger           | swagger-annotations         | 1.6.1              | swagger-annotations-1.6.1.jar               |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-annotations-1.6.1.jar.sha1          |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-annotations-1.6.1.pom               |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-annotations-1.6.1.pom.sha1          |
+                      +-----------------------------+--------------------+---------------------------------------------+
|                      | swagger-compat-spec-parser  | 1.0.51             | swagger-compat-spec-parser-1.0.51.jar       |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-compat-spec-parser-1.0.51.jar.sha1  |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-compat-spec-parser-1.0.51.pom       |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-compat-spec-parser-1.0.51.pom.sha1  |
+                      +-----------------------------+--------------------+---------------------------------------------+
|                      | swagger-core                | 1.6.1              | swagger-core-1.6.1.jar                      |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-core-1.6.1.jar.sha1                 |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-core-1.6.1.pom                      |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-core-1.6.1.pom.sha1                 |
+                      +-----------------------------+                    +---------------------------------------------+
|                      | swagger-models              |                    | swagger-models-1.6.1.jar                    |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-models-1.6.1.jar.sha1               |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-models-1.6.1.pom                    |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-models-1.6.1.pom.sha1               |
+                      +-----------------------------+--------------------+---------------------------------------------+
|                      | swagger-parser              | 1.0.51             | swagger-parser-1.0.51.jar                   |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-parser-1.0.51.jar.sha1              |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-parser-1.0.51.pom                   |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-parser-1.0.51.pom.sha1              |
+                      +-----------------------------+                    +---------------------------------------------+
|                      | swagger-parser-project      |                    | swagger-parser-project-1.0.51.pom           |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-parser-project-1.0.51.pom.sha1      |
+                      +-----------------------------+--------------------+---------------------------------------------+
|                      | swagger-project             | 1.6.1              | swagger-project-1.6.1.pom                   |
+                      +                             +                    +---------------------------------------------+
|                      |                             |                    | swagger-project-1.6.1.pom.sha1              |
+----------------------+-----------------------------+--------------------+---------------------------------------------+
|   TOTAL GROUPS: 3    |     TOTAL ARTIFACTS: 16     | TOTAL VERSIONS: 16 |               TOTAL FILES: 56               |
+----------------------+-----------------------------+--------------------+---------------------------------------------+
2021-12-13 16:33:58.542	INFO	Begin to migrate ...
Pushing: Done! [==============================================================================] 56 / 56  100 %
2021-12-13 16:34:08.780	INFO	End to migrate.	{"duration": "10.237994014s", "succeededCount": 56, "skippedCount": 0, "failedCount": 0}
2021-12-13 16:34:08.780	INFO	Migrate result:
+---------------------------------------------------------+---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                        ARTIFACT                         |                                                                 PATH                                                                  |  RESULT   |
+---------------------------------------------------------+---------------------------------------------------------------------------------------------------------------------------------------+-----------+
| io.swagger.core.v3:swagger-annotations:2.1.2            | /home/juan/.m2/swagger-repository/io/swagger/core/v3/swagger-annotations/2.1.2/swagger-annotations-2.1.2.jar                          | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/core/v3/swagger-annotations/2.1.2/swagger-annotations-2.1.2.jar.sha1                     | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/core/v3/swagger-annotations/2.1.2/swagger-annotations-2.1.2.pom                          | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/core/v3/swagger-annotations/2.1.2/swagger-annotations-2.1.2.pom.sha1                     | Succeeded |
+---------------------------------------------------------+---------------------------------------------------------------------------------------------------------------------------------------+-----------+
| io.swagger.core.v3:swagger-core:2.1.2                   | /home/juan/.m2/swagger-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.jar                                        | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.jar.sha1                                   | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.pom                                        | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/core/v3/swagger-core/2.1.2/swagger-core-2.1.2.pom.sha1                                   | Succeeded |
+---------------------------------------------------------+---------------------------------------------------------------------------------------------------------------------------------------+-----------+
| io.swagger.core.v3:swagger-models:2.1.2                 | /home/juan/.m2/swagger-repository/io/swagger/core/v3/swagger-models/2.1.2/swagger-models-2.1.2.jar                                    | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/core/v3/swagger-models/2.1.2/swagger-models-2.1.2.jar.sha1                               | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/core/v3/swagger-models/2.1.2/swagger-models-2.1.2.pom                                    | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/core/v3/swagger-models/2.1.2/swagger-models-2.1.2.pom.sha1                               | Succeeded |
+---------------------------------------------------------+---------------------------------------------------------------------------------------------------------------------------------------+-----------+
| io.swagger.core.v3:swagger-project:2.1.2                | /home/juan/.m2/swagger-repository/io/swagger/core/v3/swagger-project/2.1.2/swagger-project-2.1.2.pom                                  | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/core/v3/swagger-project/2.1.2/swagger-project-2.1.2.pom.sha1                             | Succeeded |
+---------------------------------------------------------+---------------------------------------------------------------------------------------------------------------------------------------+-----------+
| io.swagger.parser.v3:swagger-parser-core:2.0.20         | /home/juan/.m2/swagger-repository/io/swagger/parser/v3/swagger-parser-core/2.0.20/swagger-parser-core-2.0.20.pom                      | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/parser/v3/swagger-parser-core/2.0.20/swagger-parser-core-2.0.20.pom.sha1                 | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/parser/v3/swagger-parser-core/2.0.20/swagger-parser-core-2.0.20.jar                      | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/parser/v3/swagger-parser-core/2.0.20/swagger-parser-core-2.0.20.jar.sha1                 | Succeeded |
+---------------------------------------------------------+---------------------------------------------------------------------------------------------------------------------------------------+-----------+
| io.swagger.parser.v3:swagger-parser-project:2.0.20      | /home/juan/.m2/swagger-repository/io/swagger/parser/v3/swagger-parser-project/2.0.20/swagger-parser-project-2.0.20.pom.sha1           | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/parser/v3/swagger-parser-project/2.0.20/swagger-parser-project-2.0.20.pom                | Succeeded |
+---------------------------------------------------------+---------------------------------------------------------------------------------------------------------------------------------------+-----------+
| io.swagger.parser.v3:swagger-parser-v2-converter:2.0.20 | /home/juan/.m2/swagger-repository/io/swagger/parser/v3/swagger-parser-v2-converter/2.0.20/swagger-parser-v2-converter-2.0.20.pom.sha1 | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/parser/v3/swagger-parser-v2-converter/2.0.20/swagger-parser-v2-converter-2.0.20.jar      | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/parser/v3/swagger-parser-v2-converter/2.0.20/swagger-parser-v2-converter-2.0.20.jar.sha1 | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/parser/v3/swagger-parser-v2-converter/2.0.20/swagger-parser-v2-converter-2.0.20.pom      | Succeeded |
+---------------------------------------------------------+---------------------------------------------------------------------------------------------------------------------------------------+-----------+
| io.swagger.parser.v3:swagger-parser-v3:2.0.20           | /home/juan/.m2/swagger-repository/io/swagger/parser/v3/swagger-parser-v3/2.0.20/swagger-parser-v3-2.0.20.jar.sha1                     | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/parser/v3/swagger-parser-v3/2.0.20/swagger-parser-v3-2.0.20.pom                          | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/parser/v3/swagger-parser-v3/2.0.20/swagger-parser-v3-2.0.20.pom.sha1                     | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/parser/v3/swagger-parser-v3/2.0.20/swagger-parser-v3-2.0.20.jar                          | Succeeded |
+---------------------------------------------------------+---------------------------------------------------------------------------------------------------------------------------------------+-----------+
| io.swagger.parser.v3:swagger-parser:2.0.20              | /home/juan/.m2/swagger-repository/io/swagger/parser/v3/swagger-parser/2.0.20/swagger-parser-2.0.20.pom.sha1                           | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/parser/v3/swagger-parser/2.0.20/swagger-parser-2.0.20.pom                                | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/parser/v3/swagger-parser/2.0.20/swagger-parser-2.0.20.jar.sha1                           | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/parser/v3/swagger-parser/2.0.20/swagger-parser-2.0.20.jar                                | Succeeded |
+---------------------------------------------------------+---------------------------------------------------------------------------------------------------------------------------------------+-----------+
| io.swagger:swagger-annotations:1.6.1                    | /home/juan/.m2/swagger-repository/io/swagger/swagger-annotations/1.6.1/swagger-annotations-1.6.1.jar                                  | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/swagger-annotations/1.6.1/swagger-annotations-1.6.1.jar.sha1                             | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/swagger-annotations/1.6.1/swagger-annotations-1.6.1.pom                                  | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/swagger-annotations/1.6.1/swagger-annotations-1.6.1.pom.sha1                             | Succeeded |
+---------------------------------------------------------+---------------------------------------------------------------------------------------------------------------------------------------+-----------+
| io.swagger:swagger-compat-spec-parser:1.0.51            | /home/juan/.m2/swagger-repository/io/swagger/swagger-compat-spec-parser/1.0.51/swagger-compat-spec-parser-1.0.51.jar                  | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/swagger-compat-spec-parser/1.0.51/swagger-compat-spec-parser-1.0.51.jar.sha1             | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/swagger-compat-spec-parser/1.0.51/swagger-compat-spec-parser-1.0.51.pom                  | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/swagger-compat-spec-parser/1.0.51/swagger-compat-spec-parser-1.0.51.pom.sha1             | Succeeded |
+---------------------------------------------------------+---------------------------------------------------------------------------------------------------------------------------------------+-----------+
| io.swagger:swagger-core:1.6.1                           | /home/juan/.m2/swagger-repository/io/swagger/swagger-core/1.6.1/swagger-core-1.6.1.jar                                                | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/swagger-core/1.6.1/swagger-core-1.6.1.jar.sha1                                           | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/swagger-core/1.6.1/swagger-core-1.6.1.pom.sha1                                           | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/swagger-core/1.6.1/swagger-core-1.6.1.pom                                                | Succeeded |
+---------------------------------------------------------+---------------------------------------------------------------------------------------------------------------------------------------+-----------+
| io.swagger:swagger-models:1.6.1                         | /home/juan/.m2/swagger-repository/io/swagger/swagger-models/1.6.1/swagger-models-1.6.1.jar                                            | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/swagger-models/1.6.1/swagger-models-1.6.1.jar.sha1                                       | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/swagger-models/1.6.1/swagger-models-1.6.1.pom                                            | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/swagger-models/1.6.1/swagger-models-1.6.1.pom.sha1                                       | Succeeded |
+---------------------------------------------------------+---------------------------------------------------------------------------------------------------------------------------------------+-----------+
| io.swagger:swagger-parser-project:1.0.51                | /home/juan/.m2/swagger-repository/io/swagger/swagger-parser-project/1.0.51/swagger-parser-project-1.0.51.pom                          | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/swagger-parser-project/1.0.51/swagger-parser-project-1.0.51.pom.sha1                     | Succeeded |
+---------------------------------------------------------+---------------------------------------------------------------------------------------------------------------------------------------+-----------+
| io.swagger:swagger-parser:1.0.51                        | /home/juan/.m2/swagger-repository/io/swagger/swagger-parser/1.0.51/swagger-parser-1.0.51.jar.sha1                                     | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/swagger-parser/1.0.51/swagger-parser-1.0.51.pom                                          | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/swagger-parser/1.0.51/swagger-parser-1.0.51.pom.sha1                                     | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/swagger-parser/1.0.51/swagger-parser-1.0.51.jar                                          | Succeeded |
+---------------------------------------------------------+---------------------------------------------------------------------------------------------------------------------------------------+-----------+
| io.swagger:swagger-project:1.6.1                        | /home/juan/.m2/swagger-repository/io/swagger/swagger-project/1.6.1/swagger-project-1.6.1.pom.sha1                                     | Succeeded |
+                                                         +---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                                                         | /home/juan/.m2/swagger-repository/io/swagger/swagger-project/1.6.1/swagger-project-1.6.1.pom                                          | Succeeded |
+---------------------------------------------------------+---------------------------------------------------------------------------------------------------------------------------------------+-----------+
|                          TOTAL                          |                                                                  56                                                                   |            
+---------------------------------------------------------+---------------------------------------------------------------------------------------------------------------------------------------+-----------+
```

Migrate your nexus maven repository to a remote maven repository:

```shell
$ carctl migrate maven --src-type=nexus3 --src=http://localhost:8081/repository/maven-test/ --src-username=admin --src-password=admin123 --dst=http://codingcorp-maven.pkg.nh4ivfk.dev.coding.io/repository/registry/overridable-maven-migrate/ 
```
