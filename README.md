# Coyote Tester #
Part of Landoop™ test suite.


Coyote is a simple test agent. It takes instructions from a yml configuration
file with commands to setup (stdin, env vars) and run. It checks the output
for errors and may further search for the presence or absence of specific
regular expressions.

It creates a html report with the tests, their outputs and some statistics.

To build:

    go generate
    go build

To execute:

    ./coyote-tester -c conf.yml

_Notice:_ coyote stores the stderr and stdout of each command in memory, so it
isn't suitable for testing commands with huge outputs.

Sample entry in configuration yml file:

```
- name: Test 1
  entries:
    - name: Command 1
      command: ls /

    - name: Command 2
      command: ls
      workdir: /home

    - command: ls /proc
      nolog: true
```

Advanced options:

```
- name: Test 2
  entries:
    - name: Long running command
      command: cat
      timeout: 300s
      stdin: hello

```
