import asyncio
from unittest import mock

import grpc
import tensorflow as tf
import tensorflow_federated
from flaap import taskstore_pb2
from flaap.tff import executors
from tensorflow_federated.proto.v0 import executor_pb2
from tensorflow_federated.python.common_libs import py_typecheck
from tensorflow_federated.python.core.impl.computation import computation_impl
from tensorflow_federated.python.core.impl.tensorflow_context import (
    tensorflow_computation,
)
from tensorflow_federated.python.core.impl.types import computation_types
import pytest

# N.B looks like value_serialization gets moved to executor_serialization in 0.34
if tensorflow_federated.__version__ < "0.34.0":
    from tensorflow_federated.python.core.impl.executors import value_serialization

    executor_serialization = value_serialization
else:
    from tensorflow_federated.python.core.impl.executors import executor_serialization


class FakeRequestFn:
    """Create a fake to inject for the request_fn"""

    def __init__(self, response=None):
        self.request = None
        self.response = response

    def __call__(self, func, request):
        self.request = request
        return self.response


def test_create_call_no_arg():
    channel = mock.MagicMock(spec=grpc.Channel)
    executor = executors.TaskStoreExecutor(channel=channel)

    # Construct the remotevalue task to be returned
    response = taskstore_pb2.CreateResponse()
    response.task.metadata.name = "returnedname"

    fake = FakeRequestFn(response)
    executor._request_fn = fake

    # Create a TaskValue to represent the remort function.
    function_type = computation_types.FunctionType(computation_types.TensorType(tf.int32),computation_types.TensorType(tf.int32))
    function = executors.TaskValue("functask", function_type, executor)

    result = asyncio.run(executor.create_call(function))

    assert result.name == "returnedname"

    actual_task = executor._request_fn.request.task

    assert actual_task.metadata.name != ""
    assert actual_task.group_nonce == executor.group_nonce
    assert len(actual_task.input.create_call) > 0

    request = executor_pb2.CreateCallRequest()
    request.ParseFromString(actual_task.input.create_call)
    
    assert request.function_ref.id == "functask"


def test_create_call():
    channel = mock.MagicMock(spec=grpc.Channel)
    executor = executors.TaskStoreExecutor(channel=channel)

    # Construct the remotevalue task to be returned
    response = taskstore_pb2.CreateResponse()
    response.task.metadata.name = "returnedname"

    fake = FakeRequestFn(response)
    executor._request_fn = fake

     # Create a TaskValue to represent the remort function.
    function_type = computation_types.FunctionType(computation_types.TensorType(tf.int32),computation_types.TensorType(tf.int32))
    function = executors.TaskValue("functask", function_type, executor)
    argument = executors.TaskValue("argument", computation_types.TensorType(tf.int32), executor)

    result = asyncio.run(executor.create_call(function, argument))

    assert result.name == "returnedname"

    actual_task = executor._request_fn.request.task

    assert actual_task.metadata.name != ""
    assert actual_task.group_nonce == executor.group_nonce
    assert len(actual_task.input.create_call) > 0

    request = executor_pb2.CreateCallRequest()
    request.ParseFromString(actual_task.input.create_call)
    
    assert request.function_ref.id == "functask"
    assert request.argument_ref.id == "argument"

# TODO(jeremy): Need a test for create_value
def test_create_value():
    # Test that we can create a value representing a computation    
    @tensorflow_computation.tf_computation(tf.int32)
    def comp(x):
        return x + 1

    computation = computation_impl.ConcreteComputation.get_proto(comp)

    channel = mock.MagicMock(spec=grpc.Channel)
    executor = executors.TaskStoreExecutor(channel=channel)

    # Construct the remotevalue task to be returned
    response = taskstore_pb2.CreateResponse()
    response.task.metadata.name = "returnedname"

    fake = FakeRequestFn(response)
    executor._request_fn = fake
    
    result = asyncio.run(executor.create_value(computation))

    assert result.name == "returnedname"

    actual_task = executor._request_fn.request.task

    assert actual_task.metadata.name != ""
    assert actual_task.group_nonce == executor.group_nonce
    assert len(actual_task.input.create_value) > 0

    request = executor_pb2.CreateValueRequest()
    request.ParseFromString(actual_task.input.create_value)

    _, actual_comp_type = executor_serialization.deserialize_value(request.value)

    py_typecheck.check_type(actual_comp_type, computation_types.FunctionType)


# Patch the request object
@mock.patch("flaap.networking.wait_for_task")
def test_compute(wait_for_task):
    channel = mock.MagicMock(spec=grpc.Channel)
    executor = executors.TaskStoreExecutor(channel=channel)

    # Construct the remotevalue task to be returned
    # The task should contain the result of the computation in this case an integer
    result, _ = executor_serialization.serialize_value(10, tf.int32)

    # Response is the task that will be returned by the call to create the task
    # that _compute issues.
    response = taskstore_pb2.CreateResponse()
    response.task.metadata.name = "returnedname"
    response.task.output.compute = result.SerializeToString()

    fake = FakeRequestFn(response)
    executor._request_fn = fake

    # We also need to set the value for the wait_for_task mock
    wait_for_task.return_value = response.task

    # Output should be the actual materialized value
    actual = asyncio.run(executor._compute("sometask"))
    assert actual == 10
