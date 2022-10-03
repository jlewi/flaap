import asyncio
from concurrent import futures
from unittest import mock

import grpc
import portpicker
import tensorflow as tf
from flaap import conditions, taskstore_pb2, taskstore_pb2_grpc
from flaap.tff import task_handler
from tensorflow_federated.proto.v0 import executor_pb2
from tensorflow_federated.python.core.impl.computation import computation_impl
from tensorflow_federated.python.core.impl.tensorflow_context import (
    tensorflow_computation,
)
from tensorflow_federated.python.core.impl.types import computation_types


def test_handle_task_create_value():
    @tensorflow_computation.tf_computation
    def comp():
        return 1000

    computation = computation_impl.ConcreteComputation.get_proto(comp)

    task = taskstore_pb2.Task()

    value = executor_pb2.Value(computation=computation)
    create_request = executor_pb2.CreateValueRequest(value=value)
    task.metadata.name = "somefunc"
    task.input.create_value = create_request.SerializeToString()

    handler = task_handler.TaskHandler(mock.MagicMock())
    asyncio.run(handler._handle_task(task))

    assert conditions.get(task, conditions.SUCCEEDED) == taskstore_pb2.TRUE

    # Ensure the value is stored
    stored = handler._wrapper._values["somefunc"]
    assert stored.type_signature == computation_types.FunctionType(None, tf.int32)


def test_handle_task_create_call():
    @tensorflow_computation.tf_computation
    def comp():
        return 1000

    handler = task_handler.TaskHandler(mock.MagicMock())

    # Embed the function in the wrapper
    asyncio.run(
        handler._wrapper.create_value(
            "somefunc", comp, computation_types.FunctionType(None, tf.int32)
        )
    )

    task = taskstore_pb2.Task()

    call_request = executor_pb2.CreateCallRequest(
        function_ref=executor_pb2.ValueRef(id="somefunc")
    )
    task.metadata.name = "result"
    task.input.create_call = call_request.SerializeToString()

    asyncio.run(handler._handle_task(task))

    assert conditions.get(task, conditions.SUCCEEDED) == taskstore_pb2.TRUE

    # Ensure the value is stored
    result = asyncio.run(handler._wrapper.get_value("result").compute())
    assert result == 1000


def test_handle_task_create_call_with_arg():
    @tensorflow_computation.tf_computation(tf.int32)
    def comp(x):
        return x + 2

    handler = task_handler.TaskHandler(mock.MagicMock())

    # Embed the function in the wrapper
    asyncio.run(
        handler._wrapper.create_value(
            "somefunc", comp, computation_types.FunctionType(tf.int32, tf.int32)
        )
    )

    # Embed the argument in the wrapper
    asyncio.run(handler._wrapper.create_value("somearg", 10, tf.int32))
    task = taskstore_pb2.Task()

    call_request = executor_pb2.CreateCallRequest(
        function_ref=executor_pb2.ValueRef(id="somefunc"),
        argument_ref=executor_pb2.ValueRef(id="somearg"),
    )
    task.metadata.name = "result"
    task.input.create_call = call_request.SerializeToString()

    asyncio.run(handler._handle_task(task))

    assert conditions.get(task, conditions.SUCCEEDED) == taskstore_pb2.TRUE

    # Ensure the value is stored
    result = asyncio.run(handler._wrapper.get_value("result").compute())
    assert result == 12


class _TasksServicer(taskstore_pb2_grpc.TasksService):
    """Implement the TasksServicer for poll and handle test"""

    def __init__(self):
        # Number of times get called
        self._get_call_index = 0
        self._tasks = {}

        self._list_requests = []
        self._list_responses = []

    def Get(self, request, context):
        self._get_call_index += 1
        t = self._get_tasks[self._get_call_index - 1]
        response = taskstore_pb2.GetResponse(task=t)
        return response

    def List(self, request, context):
        # Save the list request for verification in the test
        index = len(self._list_requests)
        self._list_requests.append(request)
        response = self._list_responses[index]

        return response

    def Create(self, request, context):
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Create not implemented")
        return taskstore_pb2.CreateResponse()

    def Delete(self, request, context):
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Delete not implemented")
        return taskstore_pb2.DeleteResponse()

    def Update(self, request, context):
        self._tasks[request.task.metadata.name] = request.task
        # Store worker_id for verification
        self._update_worker_id = request.worker_id
        return taskstore_pb2.UpdateResponse(task=request.task)


# Patch the handle task function so that we don't actually try to process tasks using TFEager
@mock.patch.object(task_handler.TaskHandler, "_handle_task")
def test_poll_and_handle_task_success(handle_task_fn):
    """This test verifies we can successfully poll for a task and handle it."""
    port = portpicker.pick_unused_port()

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    server.add_insecure_port("[::]:{}".format(port))

    servicer = _TasksServicer()

    # Add two tasks. The second of which will have lower
    # group_index. This way we verify we select the task with lower group_index
    task = taskstore_pb2.Task()
    task.metadata.name = "task1"
    task.group_index = 10

    task2 = taskstore_pb2.Task()
    task2.metadata.name = "task2"
    task2.group_index = 1

    list_response = taskstore_pb2.ListResponse()
    list_response.items.append(task)
    list_response.items.append(task2)
    servicer._list_responses.append(list_response)

    taskstore_pb2_grpc.add_TasksServiceServicer_to_server(servicer, server)
    server.start()

    handled_task = taskstore_pb2.Task()
    handled_task.MergeFrom(task)
    conditions.set(handled_task, conditions.SUCCEEDED, taskstore_pb2.TRUE)
    handle_task_fn.return_value = handled_task

    with grpc.insecure_channel(f"localhost:{port}") as channel:
        tasks_stub = taskstore_pb2_grpc.TasksServiceStub(channel)
        handler = task_handler.TaskHandler(tasks_stub)

        result = asyncio.run(handler._poll_and_handle_task())
        assert result == "task2"

        # Ensure worker_id is set correctly
        list_request = servicer._list_requests[0]
        assert list_request.worker_id == handler._worker_id
        assert list_request.done == False

        # Make sure update request contains worker_id
        assert servicer._update_worker_id == handler._worker_id
        # Make sure test is correctly maked as done
        actual_task = servicer._tasks["task1"]
        assert conditions.get(actual_task, conditions.SUCCEEDED) == taskstore_pb2.TRUE


# Patch the handle task function so that we don't actually try to process tasks using TFEager
@mock.patch.object(task_handler.TaskHandler, "_handle_task")
def test_poll_and_handle_task_no_tasks(handle_task_fn):
    """This test verifies we can successfully poll for a task and deal with no tasks."""
    port = portpicker.pick_unused_port()

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    server.add_insecure_port("[::]:{}".format(port))

    servicer = _TasksServicer()
    taskstore_pb2_grpc.add_TasksServiceServicer_to_server(servicer, server)
    server.start()

    servicer._list_responses = [taskstore_pb2.ListResponse()]

    with grpc.insecure_channel(f"localhost:{port}") as channel:
        tasks_stub = taskstore_pb2_grpc.TasksServiceStub(channel)
        handler = task_handler.TaskHandler(tasks_stub)
        handler._polling_interval = 0.1
        result = asyncio.run(handler._poll_and_handle_task())
        assert result == ""

        # Ensure worker_id is set correctly
        list_request = servicer._list_requests[0]
        assert list_request.worker_id == handler._worker_id
        assert list_request.done == False
