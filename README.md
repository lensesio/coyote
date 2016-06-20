A simple test agent. To build:

    go generate
    go build

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
      command: |
        sqoop-import --connect jdbc:mysql://mariadb.landoop.com/northwind
                     --username data --password base
                     --table Orders --m 1
      timeout: 300s
```
