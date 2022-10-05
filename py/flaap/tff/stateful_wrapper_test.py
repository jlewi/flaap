import asyncio

import tensorflow as tf
import tensorflow_federated
from flaap.tff import stateful_wrapper
from tensorflow_federated.proto.v0 import executor_pb2
from tensorflow_federated.python.core.impl.computation import computation_impl
from tensorflow_federated.python.core.impl.executors import (
    eager_tf_executor,
    value_serialization,
)
from tensorflow_federated.python.core.impl.tensorflow_context import (
    tensorflow_computation,
)
from tensorflow_federated.python.core.impl.types import computation_types

# N.B looks like value_serialization gets moved to executor_serialization in 0.34
if tensorflow_federated.__version__ < "0.34.0":
    from tensorflow_federated.python.core.impl.executors import value_serialization

    executor_serialization = value_serialization
else:
    pass

# N.B. The unittest for task_handler provides some additional coverage because it also
# invokes the stateful_wrapper.


def test_call():
    # Run a simple test. Verify that we can properly embed a simple
    # computation and then compute the value using the TFEager executor
    wrapper = stateful_wrapper.StatefulWrapper(
        target_executor=eager_tf_executor.EagerTFExecutor()
    )

    @tensorflow_computation.tf_computation
    def comp():
        return 1000

    comp_proto = computation_impl.ConcreteComputation.get_proto(comp)
    asyncio.run(
        wrapper.create_value(
            "somefunc", comp_proto, computation_types.FunctionType(None, tf.int32)
        )
    )

    asyncio.run(wrapper.create_call("result", "somefunc"))
    result = asyncio.run(wrapper.get_value("result").compute())
    assert result == 1000


def test_call_with_arg():
    # Run a simple test. Verify that we can properly embed a simple
    # computation and then compute the value using the TFEager executor
    wrapper = stateful_wrapper.StatefulWrapper(
        target_executor=eager_tf_executor.EagerTFExecutor()
    )

    @tensorflow_computation.tf_computation(tf.int32)
    def comp(x):
        return 1000 + x

    computation_impl.ConcreteComputation.get_proto(comp)
    asyncio.run(
        wrapper.create_value(
            "somefunc", comp, computation_types.FunctionType(tf.int32, tf.int32)
        )
    )

    asyncio.run(wrapper.create_value("somearg", 10, tf.int32))

    asyncio.run(wrapper.create_call("result", "somefunc", "somearg"))
    result = asyncio.run(wrapper.get_value("result").compute())
    assert result == 1010


def test_create_struct_named():
    # Test create struct with named fields
    wrapper = stateful_wrapper.StatefulWrapper(
        target_executor=eager_tf_executor.EagerTFExecutor()
    )

    asyncio.run(wrapper.create_value("someval", 10, tf.int32))

    value_ref = executor_pb2.ValueRef(id="someval")
    elements = [executor_pb2.CreateStructRequest.Element(name="a", value_ref=value_ref)]
    asyncio.run(wrapper.create_struct("result", elements))
    result = asyncio.run(wrapper.get_value("result").compute())
    assert result["a"] == 10


def test_create_struct():
    # Test create struct with unnamed fields
    wrapper = stateful_wrapper.StatefulWrapper(
        target_executor=eager_tf_executor.EagerTFExecutor()
    )

    asyncio.run(wrapper.create_value("someval", 10, tf.int32))

    value_ref = executor_pb2.ValueRef(id="someval")
    elements = [executor_pb2.CreateStructRequest.Element(value_ref=value_ref)]
    asyncio.run(wrapper.create_struct("result", elements))
    result = asyncio.run(wrapper.get_value("result").compute())
    assert result[0] == 10
