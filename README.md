# docker-lu project

## Introduction

docker-lu is a small GO program that adapts container files, /etc/passwd, and /etc/group with docker host local user UID & GID.

## The use case is:

When a user executes a container as himself on his file system, the container can create files.
The created files must not be owned by a user ID that is not the user himself.

The following example shows correct behavior.

```bash
[me@localhost tmp ] ll
total 0
[me@localhost tmp ] docker run -it -v $(pwd):/home/me -w /home/me -u $(id -u):$(id -g) --rm alpine sh -c "echo blabla > test.txt"
[me@localhost tmp ] ll
total 4
-rw-r--r-- 1 me me 7 Jul  5 16:14 test.txt
```

This command works.

However, using a more advanced command like `git` can result in an error such as `fatal: unable to look up current user in the passwd file: no such user`

```bash
[me@localhost tmp ] docker run --rm -e http_proxy -e https_proxy -e no_proxy -it -u 1001:1001 forjdevops/jenkins git clone https://github.com/forj-oss/jenkins-install-inits /tmp/jenkins-install-inits
Cloning into '/tmp/jenkins-install-inits'...
remote: Counting objects: 423, done.
remote: Total 423 (delta 0), reused 0 (delta 0), pack-reused 422
Receiving objects: 100% (423/423), 340.18 KiB | 0 bytes/s, done.
Resolving deltas: 100% (171/171), done.
fatal: unable to look up current user in the passwd file: no such user
Unexpected end of command stream
```

You cannot replace the `-u $(id -u):$(id -g)` by `-u $(id -un):$(id -gn)` or you may get following error: 
`docker: Error response from daemon: linux spec user: unable to find user me: no matching entries in passwd file.`

Why are we having this issue? The user's uid or username is not recognized in the container. This means the uid or username is not in the container passwd file.

## How to resolve it?

To resolve this, we need to ensure that the UID/GID given is listed in `/etc/passwd` and `/etc/group` in the container.

You can fix it easily with a basic bash script to run as root:

```bash
sed -i 's/\(devops:x:\)1000:1000/\1'"$1"':'"$2"'/g' /etc/passwd
sed -i 's/\(devops:x:\)1000/\1'"$2"'/g' /etc/group
```

If `sed` has been installed you can use that code.

If `sed` is not installed, you have to :
- install it
- create a script with those 2 `sed` commands, and check parameters.

Using `docker-lu` globally can do exactly this.

Why did we created a GO program instead of a script to do this?

- you do not need install anything other than docker-lu
- you do not need to create a script to test parameters, docker-lu will handle it
- you limit the risk of breaking your container if your script incorrectly updates those files (bad error handling)
- docker-lu refuses to work outside a container and as non-container root.

## How to use docker-lu

You should use docker-lu if all following conditions are true

- if you add `-u` to `docker run`
- if your container has a local FS mounted (`-v /local/fs:/data`)
- if your container writes to/updates the mounted path with the user rights given (`-u`)
- if the user uid is not recognized in the /etc/passwd of the container
- if you REALLY need to be in the /etc/passwd (because of git, npm, go, etc...)

If your use case is confirmed, do the following:

### Use case 1 - Dockerfile

You can add `docker-lu` in your docker image, call it as root during the entrypoint and then become that user with `su -` or any equivalent command.

1. Add `ADD https://github.com/forj-oss/docker-lu/releases/download/0.1/docker-lu /usr/local/bin/docker-lu`
2. Add `RUN chmod +x /usr/local/bin/docker-lu`
3. Add a call to docker-lu in the entrypoint. ex: `ENTRYPOINT [ "/usr/local/bin/entrypoint.sh" ]`

    ... assume jenkins exists as user in the image ...
    
    ```bash
    [...]
    # entrypoint.sh with docker-lu
    /usr/local/bin/docker-lu "jenkins" $UID "jenkins" $GID

    ```
4. Run it

    ```bash
    docker run -it --name test --rm -e UID=$(id -u) -e GID=$(id -g) <image> <tool> <parameters>
    ```

### Use case 2 - docker run as daemon

If you run your container as daemon and execute commands through `docker exec`

1. Download `docker-lu`. 

    `wget -O ~/bin/docker-lu https://github.com/forj-oss/docker-lu/releases/download/0.1/docker-lu`

2. Give executable rights

    `chmod +x ~/bin/docker-lu`
3. Start your container as docker daemon: 
    
    ex: it will use `developer` as declared user.
    
    `docker run -u $(id -u):$(id -g) --name test -it -d alpine sh -c 'adduser developer -D;cat'`
4. copy docker-lu: `docker copy ~/bin/docker-lu test:/usr/bin/docker-lu`
5. run it : `docker exec -u 0 -it test docker-lu developer $(id -u) developer $(id -g)`


check it : `docker exec -it test id`

## Build the project

If you want to contribute to this project (bug/enhancement), feel free to create a PR or just ask through issues.

This section explains how to build it.

*Small IDE experience*:

If you want to debug through a IDE, it works great from Visual Studio Code under linux.
I never tested in other OS.

I was using idea, but not it is unusable, because they built a non free IDE called goglang. 
I was not able to debug test files... From VSC, both work and are free (Opensource)

### First time

```bash
mkdir -p ~/go/src
git clone https://github.com/forj-oss/docker-lu.git
source build-env.sh ~/go
create-go-build-env.sh
glide i
go build
```

You can add --sudo if your docker is runnable from sudo only

**NOTE**: ~/go is your GOPATH. Change ~/go or update .be-gopath to a different path if your project is cloned somewhere else.

### Next time

```bash
cd ~/go/src/docker-lu
build-env
glide i
go build
```

### troubleshoot

- Running `glide` or `go` return following error. How to fix it?

    ```bash
    ++ sudo docker run --rm -i -t -v /home/larsonsh/src/forj:/go -w /go/src/docker-lu -u 10001 docker-lu-go-env /usr/bin/glide init
    Unable to find image 'docker-lu-go-env:latest' locally
    docker: Error response from daemon: repository docker-lu-go-env not found: does not exist or no pull access.
    See 'docker run --help'.
    ++ _be_restore_debug
    ++ [[ '' != x ]]
    ++ set +x
    ```

    Run create-go-build-env.sh

- Running `source build-env.sh` return following error. How to fix it?

    ```bash
    $ source build-env.sh
    Loading module go ...
    Using docker directly. (no sudo)
    Got permission denied while trying to connect to the Docker daemon socket at unix:///var/run/docker.sock: Get http://%2Fvar%2Frun%2Fdocker.sock/v1.29/version: dial unix /var/run/docker.sock: connect: permission denied
    ```

    source it with --sudo or fix the docker socket group.
    `source build-env.sh --sudo`

- Running `source build-env.sh` return following error. How to fix it?

    ```bash
    $ source build-env.sh
    Loading module go ...
    Using docker directly. (no sudo)
    Missing GOPATH. Please set it, or define it in your local personal '.be-gopath' file
    ```
    You missed providing the GOPATH setup

    Call `source build-env.ss ~/go`

Forj Team
