import asyncio
from concurrent import futures
from unittest import mock

import grpc
import portpicker
import tensorflow as tf
from flaap import conditions, taskstore_pb2, taskstore_pb2_grpc
from flaap.tff import task_handler
from tensorflow_federated.python.core.impl.computation import computation_impl
from tensorflow_federated.python.core.impl.executors import value_serialization
from tensorflow_federated.python.core.impl.tensorflow_context import (
    tensorflow_computation,
)


def test_handle_task():
    @tensorflow_computation.tf_computation
    def comp():
        return 1000

    computation = computation_impl.ConcreteComputation.get_proto(comp)

    task = taskstore_pb2.Task()
    task.input.function.computation.MergeFrom(computation)

    handler = task_handler.TaskHandler(mock.MagicMock())
    asyncio.run(handler._handle_task(task))

    assert conditions.get(task, conditions.SUCCEEDED) == taskstore_pb2.TRUE

    result_value, result_value_type = value_serialization.deserialize_value(task.result)
    assert result_value == 1000


def test_handle_task_with_arg():
    @tensorflow_computation.tf_computation(tf.int32)
    def comp(x):
        return x + 2

    computation = computation_impl.ConcreteComputation.get_proto(comp)

    task = taskstore_pb2.Task()
    task.input.function.computation.MergeFrom(computation)

    argument, _ = value_serialization.serialize_value(3, tf.int32)
    task.input.argument.MergeFrom(argument)

    handler = task_handler.TaskHandler(mock.MagicMock())
    asyncio.run(handler._handle_task(task))
    assert conditions.get(task, conditions.SUCCEEDED) == taskstore_pb2.TRUE

    result_value, result_value_type = value_serialization.deserialize_value(task.result)
    assert result_value == 5


class _TasksServicer(taskstore_pb2_grpc.TasksService):
    """Implement the TasksServicer for poll and handle test"""

    def __init__(self):
        # Number of times get called
        self._get_call_index = 0
        self._tasks = {}

        self._list_requests = []

    def Get(self, request, context):
        self._get_call_index += 1
        t = self._get_tasks[self._get_call_index - 1]
        response = taskstore_pb2.GetResponse(task=t)
        return response

    def List(self, request, context):
        # Save the list request for verification in the test
        self._list_requests.append(request)
        response = taskstore_pb2.ListResponse()

        for _, t in self._tasks.items():
            response.items.append(t)
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
        return taskstore_pb2.UpdateResponse(task=request.task)


# Patch the handle task function so that we don't actually try to process tasks using TFEager
@mock.patch.object(task_handler.TaskHandler, "_handle_task")
def test_poll_and_handle_task_success(handle_task_fn):
    """This test verifies we can successfully poll for a task and handle it."""
    port = portpicker.pick_unused_port()

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    server.add_insecure_port("[::]:{}".format(port))

    servicer = _TasksServicer()
    task = taskstore_pb2.Task()
    task.metadata.name = "task1"

    servicer._tasks[task.metadata.name] = task

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
        assert result == 1

        # Ensure worker_id is set correctly
        list_request = servicer._list_requests[0]
        assert list_request.worker_id == handler._worker_id
        assert list_request.done == False

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

    with grpc.insecure_channel(f"localhost:{port}") as channel:
        tasks_stub = taskstore_pb2_grpc.TasksServiceStub(channel)
        handler = task_handler.TaskHandler(tasks_stub)
        handler._polling_interval = 0.1
        result = asyncio.run(handler._poll_and_handle_task())
        assert result == 0

        # Ensure worker_id is set correctly
        list_request = servicer._list_requests[0]
        assert list_request.worker_id == handler._worker_id
        assert list_request.done == False
