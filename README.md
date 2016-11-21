# D-clone

Helps you to duplicate or migrate running containers from 1 docker daemon to another.
 
## How it works

It uses the docker API to generate a command line to run a docker container just like the one that is already running.

```shell
dclone [container-name]
```