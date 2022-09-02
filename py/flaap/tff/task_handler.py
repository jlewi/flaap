import asyncio

from tensorflow_federated.python.core.impl.executors import (
    eager_tf_executor,
    value_serialization,
)


class TaskHandler:
    """Task handler evaluates Tasks using the TFEagerExecutor"""

    def __init__(self):
        self._executor = eager_tf_executor.EagerTFExecutor()
        self._event_loop = asyncio.new_event_loop()

    async def _handle_task(self, task):
        """Handle a task.

        A task is expected to hold a CreateValueRequest. With the gRPC executor service
        the client would issue a CreateValueRequest to create the value in the executor
        and then issue a Compute request to get the value of that request. _handle_task
        effectively bundles that functionality.

        For reference: Here's a link to the remote executor service.
        https://github.com/tensorflow/federated/blob/a6506385def304c260a424b29e5b34c6d905760e/tensorflow_federated/python/core/impl/executors/executor_service.py#L223

        """
        # Desiralize the value proto into a value and type.
        function_value, function_value_type = value_serialization.deserialize_value(
            task.input.function
        )

        eager_value = await self._executor.create_value(
            function_value, function_value_type
        )
        eager_arg = None
        if task.input.HasField("argument"):
            arg_value, arg_type = value_serialization.deserialize_value(
                task.input.argument
            )

            eager_arg = await self._executor.create_value(arg_value, arg_type)
        eager_result = await self._executor.create_call(eager_value, eager_arg)

        # Invoke compute in order to get the actual value.
        result_val = await eager_result.compute()
        val_type = eager_result.type_signature
        value_proto, _ = value_serialization.serialize_value(result_val, val_type)
        task.result.MergeFrom(value_proto)
        return task
