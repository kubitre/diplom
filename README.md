[![Actions Status](https://github.com/kubitre/diplom/workflows/CI_dev/badge.svg)](https://github.com/kubitre/diplom/actions?query=workflow%3ACI_dev)
[![codecov](https://codecov.io/gh/kubitre/diplom/branch/master/graph/badge.svg?token=EGI57ppVgU)](https://codecov.io/gh/kubitre/diplom) 
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/ec88d97d61344cea92d7f299d078c9c3)](https://www.codacy.com?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=kubitre/diplom&amp;utm_campaign=Badge_Grade)


# ExecutorTasksCandidateCodes
Diploma Project. This project executing tasks stages for manipulate candidates code

## Technologies
- go 1.13.5
- docker
- go git
- yaml

## Modules:
- core. This module contain main logic of pipeline running. In this cases runner should work with running same tasks, initiate configuring environment for any tasks by configuration module
- docker. This module work with docker api and can running containers, building images by dockerfile
- gitmod. this module work with git for clonning repository candidates
- parser. This module need for parsing yaml tasks specification derivet from portal
- report. This module work with main metrics which should get by any task

## Base Scenarious
1. Initiate runner and register that in portal
2. Setting up runner for current parallel worker can be delay any tasks
3. Executing task for aggregating candidate code by any stages which setup in portal company
4. Sending to portal reports with results of executing tasks

## Initiate runner and register that in portal
## Setting up runner for current parallel worker can be delay any tasks
## Executing task for aggregating candidate code by any stages which setup in portal company
## Secnding to portal reports with results of executing tasks

