#!/bin/bash

docker run -d --name=dev-consul --net=host -e CONSUL_BIND_INTERFACE=eno2 -p 8500:8500  consul