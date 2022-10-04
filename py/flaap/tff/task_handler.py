import asyncio
import logging
import uuid

import fire
import grpc
import tenacity
from flaap import conditions, taskstore_pb2, taskstore_pb2_grpc
from flaap.tff import stateful_wrapper
from tensorflow_federated.proto.v0 import executor_pb2
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
        self._wrapper = stateful_wrapper.StatefulWrapper(
            target_executor=eager_tf_executor.EagerTFExecutor()
        )
        # TODO(jeremy): We don't seem to be using this right now but maybe we should?
        self._event_loop = asyncio.new_event_loop()
        # How often to sleep waiting for tasks; in seconds
        self._polling_interval = 30
        self._tasks_stub = tasks_stub
        # Generate a UUID used to identify the worker. This is required for task assignment.
        self._worker_id = uuid.uuid4().hex

    async def _handle_task(self, task):
        """Handle a task.

        For reference: Here's a link to the remote executor service.
        https://github.com/tensorflow/federated/blob/a6506385def304c260a424b29e5b34c6d905760e/tensorflow_federated/python/core/impl/executors/executor_service.py#L223

        """
        logging.info("Handling task: %s", task.metadata.name)

        if task.input.HasField("create_value"):
            request = executor_pb2.CreateValueRequest()
            request.ParseFromString(task.input.create_value)

            value, value_type = value_serialization.deserialize_value(request.value)
            logging.info("Create value %s", task.metadata.name)
            await self._wrapper.create_value(task.metadata.name, value, value_type)

        elif task.input.HasField("create_call"):
            request = executor_pb2.CreateCallRequest()
            request.ParseFromString(task.input.create_call)
            comp_name = request.function_ref.id
            arg_name = None
            if request.HasField("argument_ref"):
                arg_name = request.argument_ref.id
            logging.info(
                "Create call %s; comp=%s arg=%s",
                task.metadata.name,
                comp_name,
                arg_name,
            )
            await self._wrapper.create_call(task.metadata.name, comp_name, arg_name)
        elif task.input.HasField("create_struct"):
            request = executor_pb2.CreateStructRequest()
            request.ParseFromString(task.input.create_struct)

            logging.info(
                "Create struct %s",
                task.metadata.name,
            )

            await self._wrapper.create_struct(task.metadata.name, request.element)
        elif task.input.HasField("compute"):
            request = executor_pb2.ComputeRequest()
            request.ParseFromString(task.input.compute)

            # Get the value requested in the compute request
            eager_value = self._wrapper.get_value(request.value_ref.id)
            result = await eager_value.compute()
            value_proto, _ = value_serialization.serialize_value(
                result, eager_value.type_signature
            )

            # Store the output in the task
            call_response = executor_pb2.ComputeResponse(value=value_proto)

            task.output.compute = call_response.SerializeToString()

        else:
            raise NotImplementedError(
                "Code for handling task of this type is not implemented"
            )

        conditions.set(task, conditions.SUCCEEDED, taskstore_pb2.TRUE)
        return task

    async def _poll_and_handle_task(self):
        """Poll for any available tasks and handle them if they are available

        Returns:
           task_name: Name of task successfully processed if task is successfull processed and empty string otherwise.
        """
        # TODO(https://github.com/jlewi/flaap/issues/7):
        request = taskstore_pb2.ListRequest()
        # Return unfinished tasks for this worker
        request.done = False
        request.worker_id = self._worker_id
        response = _run_rpc(self._tasks_stub.List, request)

        if len(response.items) == 0:
            logging.info("No available tasks")
            await asyncio.sleep(self._polling_interval)
            return ""

        logging.info("Recieved %s tasks", len(response.items))

        # TODO(jeremy): We should catch retryable exceptions. As we identify exceptions
        # that we should retry on we should add them to a try/catch block.

        # Select the task with the smallest index
        task = response.items[0]
        for other in response.items[1:]:
            if other.group_index == task.group_index:
                logging.error(
                    "Task %s and %s had the same group_index; this shouldn't happen",
                    other.metadata.name,
                    task.metadata.name,
                )
                # Try to degrade gracefully but who know what will happen
                continue
            if other.group_index < task.group_index:
                task = other

        new_task = await self._handle_task(task)

        update_request = taskstore_pb2.UpdateRequest()
        update_request.task.MergeFrom(new_task)
        update_request.worker_id = self._worker_id

        # TODO(jeremy): We should add appropriate error and retry handling.
        response = _run_rpc(self._tasks_stub.Update, update_request)
        return task.metadata.name

    async def run(self):
        """Periodically poll the taskstore for tasks and process them"""
        logging.info("Starting polling loop; worker_id=%s", self._worker_id)
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
