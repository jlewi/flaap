import asyncio
from unittest import mock

import grpc
import tensorflow as tf
from flaap import taskstore_pb2
from flaap.tff import executors
from tensorflow_federated.python.core.impl.computation import computation_impl
from tensorflow_federated.python.core.impl.tensorflow_context import (
    tensorflow_computation,
)


# Patch the request object
@mock.patch("flaap.tff.executors._request")
def test_create_value(_request):
    channel = mock.MagicMock(spec=grpc.Channel)
    executor = executors.TaskStoreExecutor(channel=channel)

    # Construct the remotevalue task to be returned
    response = taskstore_pb2.CreateResponse()
    response.task.metadata.name = "returnedname"
    _request.return_value = response

    result = asyncio.run(executor.create_value(10, tf.int32))

    assert result.name == "returnedname"

    class CreateRequestMatcher:
        def __eq__(self, other):
            if other.task.metadata.name == "":
                return False
            return True

    _request.assert_called_once_with(mock.ANY, CreateRequestMatcher())


# Patch the request object
@mock.patch("flaap.networking.wait_for_task")
def test_compute(wait_for_task):
    @tensorflow_computation.tf_computation(tf.int32)
    def comp(x):
        return x + 1

    computation = computation_impl.ConcreteComputation.get_proto(comp)

    task = taskstore_pb2.Task()
    task.result.computation.MergeFrom(computation)
    wait_for_task.return_value = task

    channel = mock.MagicMock(spec=grpc.Channel)
    executor = executors.TaskStoreExecutor(channel=channel)
    result = asyncio.run(executor._compute("sometask"))

    assert result is not None
