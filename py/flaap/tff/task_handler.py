import asyncio
import logging
import uuid

import fire
import grpc
import tenacity
from flaap import conditions, taskstore_pb2, taskstore_pb2_grpc
from tensorflow_federated.python.common_libs import tracing
from tensorflow_federated.python.core.impl.executors import (
    eager_tf_executor,
    value_serialization,
)


class TaskHandler:
    """Task handler evaluates Tasks using the TFEagerExecutor"""

    def __init__(self, tasks_stub):
        """Create a task handler

        Args:
          tasks_stub: GRPC stub for the taskstore CRUD API.
        """
        self._executor = eager_tf_executor.EagerTFExecutor()
        self._event_loop = asyncio.new_event_loop()
        # How often to sleep waiting for tasks; in seconds
        self._polling_interval = 30
        self._tasks_stub = tasks_stub
        # Generate a UUID used to identify the worker. This is required for task assignment.
        self._worker_id = uuid.uuid4().hex

    async def _handle_task(self, task):
        """Handle a task.

        A task is expected to hold a CreateValueRequest. With the gRPC executor service
        the client would issue a CreateValueRequest to create the value in the executor
        and then issue a Compute request to get the value of that request. _handle_task
        effectively bundles that functionality.

        For reference: Here's a link to the remote executor service.
        https://github.com/tensorflow/federated/blob/a6506385def304c260a424b29e5b34c6d905760e/tensorflow_federated/python/core/impl/executors/executor_service.py#L223

        """
        logging.info("Handling task: %s", task.metadata.name)
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

        conditions.set(task, conditions.SUCCEEDED, taskstore_pb2.TRUE)
        return task

    async def _poll_and_handle_task(self):
        """Poll for any available tasks and handle them if they are available

        Returns:
           num_tasks: Number of tasks successfully handled. Should be 1 if one task is processed
             and zero otherwise.
        """
        # TODO(https://github.com/jlewi/flaap/issues/7):
        request = taskstore_pb2.ListRequest()
        # Return unfinished tasks for this worker
        request.done = False
        request.worker_id = self._worker_id
        response = _run_rpc(self._tasks_stub.List, request)

        if len(response.items) == 0:
            logging.info("No available tass")
            await asyncio.sleep(self._polling_interval)
            return 0

        logging.info("Recieved %s tasks", len(response.items))

        # TODO(jeremy): We should catch retryable exceptions. As we identify exceptions
        # that we should retry on we should add them to a try/catch block.
        task = response.items[0]

        new_task = await self._handle_task(task)

        update_request = taskstore_pb2.UpdateRequest()
        update_request.task.MergeFrom(new_task)

        # TODO(jeremy): We should add appropriate error and retry handling.
        response = _run_rpc(self._tasks_stub.Update, update_request)
        return 1

    async def run(self):
        """Periodically poll the taskstore for tasks and process them"""
        logging.info("Starting polling loop")
        while True:
            await self._poll_and_handle_task()


def _is_retryable(error):
    """Predicate defining what is a retryable gRPC error."""
    # N.B. this is a guess of which errors should be retryable. This might require some tweaking
    # We might also need different values for different methods.
    non_retryable_errors = {
        grpc.StatusCode.INVALID_ARGUMENT,
        grpc.StatusCode.NOT_FOUND,
        grpc.StatusCode.ALREADY_EXISTS,
        grpc.StatusCode.PERMISSION_DENIED,
        grpc.StatusCode.FAILED_PRECONDITION,
        grpc.StatusCode.ABORTED,
        grpc.StatusCode.OUT_OF_RANGE,
        grpc.StatusCode.UNIMPLEMENTED,
        grpc.StatusCode.DATA_LOSS,
        grpc.StatusCode.UNAUTHENTICATED,
        grpc.StatusCode.UNKNOWN,
    }
    return isinstance(error, grpc.RpcError) and error.code() not in non_retryable_errors


# TODO(jeremy): should we generalize this?
@tenacity.retry(
    retry=tenacity.retry_if_exception(_is_retryable),
    stop=tenacity.stop_after_delay(300),
)
@tracing.trace(span=True)
def _run_rpc(rpc_func, request):
    """Populates trace context and reraises gRPC errors with retryable info."""
    with tracing.wrap_rpc_in_trace_context():
        return rpc_func(request)


class Runner:
    @staticmethod
    def run(taskstore):
        """Run the taskhandler.

        Args:
          taskstore: host:port of the gRPC service for the Taskstore API
        """
        logging.info("Connecting to taskstore: %s", taskstore)
        with grpc.insecure_channel(taskstore) as channel:
            tasks_stub = taskstore_pb2_grpc.TasksServiceStub(channel)
            handler = TaskHandler(tasks_stub)
            logging.info("Starting worker task handling loop")
            asyncio.run(handler.run())


if __name__ == "__main__":
    logging.basicConfig(
        level=logging.INFO,
        format=("%(levelname)s|%(asctime)s" "|%(pathname)s|%(lineno)d| %(message)s"),
        datefmt="%Y-%m-%dT%H:%M:%S",
    )
    logging.getLogger().setLevel(logging.INFO)
    fire.Fire(Runner)
