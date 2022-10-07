# flaap
Federated Learning and Analytics Protocols (FLAAP).


## TensorFlow Federated Extension

This repository contains an extension to [TensorFlow Federated](https://github.com/tensorflow/federated).

The extension allows the direction of communication to be reversed so that TFF workers(silos) can initiate
RPCs to the coordinator as opposed to relying on coordinators issuing RPCs to the workers. 

The work is based on two design documents although the implementation has diverged a bit

* [Client Initiated Connections In TFF](https://docs.google.com/document/d/10rvJdXRtgVOYNU2cj-M4ycGLoAxI2m3BKcRJQtE9nY8/edit#heading=h.sw48ol3t02xj)

* [Task Assignment](https://docs.google.com/document/d/1T8b8Ga_ORf283FeEsz1RshsxmIhJYkbzg2j2-_WS57w/edit#heading=h.tgf0yqghramm)

The solution consists of the following components:

* [taskstore.proto](./protos/flaap/taskstore.proto) - Defines the protocol buffers and services

   * Data is exchanged between coordinator and workers using a CRUD API for Task resources
   * Task resources wrap the data exchanged in [TFF's Executor Service](https://github.com/tensorflow/federated/blob/main/tensorflow_federated/proto/v0/executor.proto)

 * [Taskstore server](./go) - A golang server providing a CRUD API for the [Task resource](./protos/flaap/taskstore.proto)

   * The current implementation serializes the tasks to a file on durable storage for fault tolerance
   * The current implementation is neither scalable nor highly performant
   * There is also a crude CLI for interacting with the taskstore

 * [TaskStoreExecutor](py/flaap/tff/executors.py) - A TFF Executor that completes operations by creating Task resources in the task store and waiting for them to be completed.
 * [TaskHandler](py/flaap/tff/task_handler.py) - Processes a Task by using the TFEagerExecutor

 * [fed_average.py](https://github.com/jlewi/flaap/blob/main/py/flaap/testing/fed_average.py) a simple example of a TFF program
    that uses the Taskstore.

 * [e2e-test](go/pkg/testing/e2e/) - A go binary providing an E2E test. It starts all the relevant processes as subprocesses.

## GoLang Protocol Buffer Libraries For TFF

The TFF Project doesn't generate GoLang versions of its client libraries
so this project generates them.
