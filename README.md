# Overview

## Supported Features

- bash (-b)
- go (-g)
- rust (-r)
- tekton (-t)
- yaml (-y)


## Install

```
git clone https://github.com/luigizuccarelli/bash-codegen
cd bash-codegen
./install
```

## Usage

```
# listed types are 
# -b (bash)
# -g (go)
# -r (rust)
# -t (tekton)
# -y (yaml)

# list yaml supported objects
cg help -y 

# use specific yaml object 
# in vim
# type shift :
# e golangci.yaml
cg golanglintci -y

```


## Credits

- k8s-examples.container-solutions.com
