from flaap.tff import stateful_executor

from tensorflow_federated.python.core.impl.executors import (
    eager_tf_executor,
    value_serialization,
)
from tensorflow_federated.python.core.impl.tensorflow_context import tensorflow_computation
from tensorflow_federated.python.core.impl.computation import computation_impl
from tensorflow_federated.python.core.impl.types import computation_types

import asyncio
import tensorflow as tf

def test_call():  
  # Run a simple test. Verify that we can properly embed a simple
  # computation and then compute the value using the TFEager executor
  wrapper = stateful_executor.StatefulWrapper(target_executor=eager_tf_executor.EagerTFExecutor())

  @tensorflow_computation.tf_computation
  def comp():
    return 1000

  comp_proto = computation_impl.ConcreteComputation.get_proto(comp)
  val = asyncio.run(
      wrapper.create_value("somefunc", comp_proto,
                           computation_types.FunctionType(None, tf.int32)))

  assert wrapper._values["somefunc"] == 1000