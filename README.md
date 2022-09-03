# flaap
Federated Learning and Analytics Protocols (FLAAP).

Define the protocols and code for [Client Initiated Connections In TFF](https://docs.google.com/document/d/10rvJdXRtgVOYNU2cj-M4ycGLoAxI2m3BKcRJQtE9nY8/edit#heading=h.sw48ol3t02xj)

## Repository overview

[taskstore.proto](./protos/flaap/taskstore.proto) - Defines the protocol buffers and services
[go](./go) - A golang server providing a CRUD API for the [Task resource](./protos/flaap/taskstore.proto)

   * The current implementation of the task store is a file
[py/flaap/tff] Python module providing the code needed to support client initiated connections in TFF

   * [TaskStoreExecutor](py/flaap/tff/executors.py) - A TFF Executor that completes operations by creating Task resources in the task store and waiting for them to be completed.
   * [TaskHandler](py/flaap/tff/task_handler.py) - Processes a Task by using the TFEagerExecutor

## GoLang Protocol Buffer Libraries For TFF

The TFF Project doesn't generate GoLang versions of its client libraries
so this project generates them.
