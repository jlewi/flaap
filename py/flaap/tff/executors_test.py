import asyncio
from unittest import mock

import grpc
import tensorflow as tf
import tensorflow_federated
from flaap import taskstore_pb2
from flaap.tff import executors
from tensorflow_federated.python.common_libs import py_typecheck
from tensorflow_federated.python.core.impl.computation import computation_impl
from tensorflow_federated.python.core.impl.tensorflow_context import (
    tensorflow_computation,
)
from tensorflow_federated.python.core.impl.types import computation_types

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
    @tensorflow_computation.tf_computation()
    def comp():
        return 20

    computation = computation_impl.ConcreteComputation.get_proto(comp)

    channel = mock.MagicMock(spec=grpc.Channel)
    executor = executors.TaskStoreExecutor(channel=channel)

    # Construct the remotevalue task to be returned
    response = taskstore_pb2.CreateResponse()
    response.task.metadata.name = "returnedname"

    fake = FakeRequestFn(response)
    executor._request_fn = fake

    # Call create_value to create the values for the computation
    function = asyncio.run(executor.create_value(computation))

    result = asyncio.run(executor.create_call(function))

    assert result.name == "returnedname"

    actual_task = executor._request_fn.request.task

    assert actual_task.metadata.name != ""
    assert actual_task.group_nonce == executor.group_nonce
    assert actual_task.input.HasField("function")
    assert not actual_task.input.HasField("argument")

    _, actual_comp_type = executor_serialization.deserialize_value(
        actual_task.input.function
    )

    py_typecheck.check_type(actual_comp_type, computation_types.FunctionType)


def test_create_call():
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

    # Call create_value to create the values for the computation and argument
    function = asyncio.run(executor.create_value(computation))
    argument = asyncio.run(executor.create_value(10, tf.int32))

    result = asyncio.run(executor.create_call(function, argument))

    assert result.name == "returnedname"

    actual_task = executor._request_fn.request.task

    assert actual_task.metadata.name != ""
    assert actual_task.group_nonce == executor.group_nonce
    assert actual_task.input.HasField("function")
    assert actual_task.input.HasField("argument")

    _, actual_comp_type = executor_serialization.deserialize_value(
        actual_task.input.function
    )

    py_typecheck.check_type(actual_comp_type, computation_types.FunctionType)

    _, actual_arg_type = executor_serialization.deserialize_value(
        actual_task.input.argument
    )
    py_typecheck.check_type(actual_arg_type, computation_types.TensorType)


# Patch the request object
@mock.patch("flaap.networking.wait_for_task")
def test_compute(wait_for_task):
    result, _ = executor_serialization.serialize_value(10, tf.int32)

    task = taskstore_pb2.Task()
    task.result.MergeFrom(result)
    wait_for_task.return_value = task

    channel = mock.MagicMock(spec=grpc.Channel)
    executor = executors.TaskStoreExecutor(channel=channel)
    result = asyncio.run(executor._compute("sometask"))

    assert result == 10
