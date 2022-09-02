import asyncio

import tensorflow as tf
from flaap import taskstore_pb2
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

    handler = task_handler.TaskHandler()
    asyncio.run(handler._handle_task(task))

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

    handler = task_handler.TaskHandler()
    asyncio.run(handler._handle_task(task))

    result_value, result_value_type = value_serialization.deserialize_value(task.result)
    assert result_value == 5
