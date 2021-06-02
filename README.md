# Coyote

_Coyote_ is a test agent. It uses a yml configuration file with commands to setup
(stdin, env vars, etc) and run. It checks the output for errors and may further
search for the presence or absence of specific regular expressions. Finally it
creates a html report with the tests, their outputs and some statistics.

> Part of Landoop™ test suite

[![build status](https://img.shields.io/travis/Landoop/coyote/master.svg?style=flat-square)](https://travis-ci.org/Landoop/coyote) [![report card](https://img.shields.io/badge/report%20card-a%2B-ff3333.svg?style=flat-square)](http://goreportcard.com/report/Landoop/coyote) [![chat](https://img.shields.io/badge/join-%20chat-00BCD4.svg?style=flat-square)](https://slackpass.io/landoop-community)

## Installation

The only requirement is the [Go Programming Language](https://golang.org/dl), at least version **1.10+**.

```sh
$ go get -u github.com/landoop/coyote
```

> This command will install the Coyote in $PATH ([setup your $GOPATH/bin](https://github.com/golang/go/wiki/SettingGOPATH) if you didn't already).

### Running

Getting your tests ran by the `coyote -c` command.

```sh
$ coyote -c my-test.yml -c my-second-test.yml # or -c ./my-tests-folder to load all yaml tests from a particular folder
```

The above command will run against those tests described in the passed test files and will generate a rich report inside the `./coyote.html` template file before exit. The exit code of _coyote_ is the number of failed tests, up to 254 failed tests.
For 255 or more failed tests, the exit code will remain at 255.

The `coyote.html` template file [needs](https://github.com/Landoop/coyote/pull/4#issuecomment-372636427) a web server to be displayed correctly, you can use any web server to achieve this, like [iris](https://iris-go.com) or python's [httpserver](https://docs.python.org/3/library/http.server.html) module, it's up to you, i.e `cd ./my-tests-folder && python -m http.server 8000`, open a new browser tab and navigate to the <http://localhost:8000/coyote.html>.

The best example for understanding how to setup a coyote test, would be the
[kafka-tests.yaml](https://github.com/Landoop/coyote/blob/master/tests/kafka-tests.yml)
which we use to test our Kafka setup for the
[Landoop Boxes](https://docs.landoop.com/pages/your-box/).

<img src="https://storage.googleapis.com/wch/coyote.png" alt="coyote screenshot" type="image/png" width="900">

> Note that coyote stores the stderr and stdout of each command in memory, so
it isn't suitable for testing commands with huge outputs

### Examples

Sample entry in configuration yml file, overview of **workdir**, **nolog** and **partially match of the standard output**:

```yml
- name: Test 1
  entries:
    - name: Command 1
      command: ls /
      stdout:
       - match: ["my-folder", "my-file.txt"]
         partial: true

    - name: Command 2
      command: ls
      workdir: /home

    - command: ls /proc
      nolog: true
```

1. `name` is the test's group name,
2. `entries` contains the tests under that group, groups are seen separated in the report page (`coyote.html`),
3. `command` is the shell command that we want to test its output, it's required,
4. `match` under the `stdout` will compare the command's output with the slice of text inside it, you can add more than one expected output,
5. the `partial: true` under the `stdout` will tell the tester that we don't want to do an exact matching, just check if "my-folder" or "my-file.txt" is *part of the command's output*.
6. the `nolog` tells the tester that we don't want to log anything from that particular command in the `coyote`'s output.

#### timeout

Timeout will stop the command's execution when duration passed.

```yml
- name: Test 2
  entries:
    - name: Long running command
      command: cat
      timeout: 300s
      stdin: hello

```

#### skip

An option you may add to your groups or per command is `skip`. This option will
skip the test if set to (case insensitive) `true`. Please note that this isn't
a boolean option but rather a string.

The idea behind it is that you can have a test like this:

```yml
- name: test 1
  skip: _test1_
  entries:
   ...
   ...
- name: test 2
  skip: _test2_
  entries:
   ...
   ...
```



And then you can easily switch off parts of the test using sed or other tools.

## Versioning

Current: **v1.4.0**

Read more about Semantic Versioning 2.0.0

- http://semver.org/
- https://en.wikipedia.org/wiki/Software_versioning
- https://wiki.debian.org/UpstreamGuide#Releases_and_Versions

## License

Distributed under Apache 2.0, See [LICENSE](LICENSE) for more information.

