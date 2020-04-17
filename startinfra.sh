#!/bin/bash

docker run -d --name=dev-consul --net=host -e CONSUL_BIND_INTERFACE=eno2 -e CONSUL_HTTP_AUTH=kubitre:password -p 8500:8500  consul