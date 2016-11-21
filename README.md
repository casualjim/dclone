# D-clone

Helps you to duplicate or migrate running containers from 1 docker daemon to another.
 
## How it works

It uses the docker API to generate a command line to run a docker container just like the one that is already running.

```shell
dclone [container-name]
```

## Example

Imagine you started a container with:

```
docker run --name blah -dit --network testing -p 82:80 tutum/hello-world
```

Then running `dclone blah` would result in

```
docker run --interactive --tty --detach --name blah --network testing --port 82:80 tutum/hello-world
```