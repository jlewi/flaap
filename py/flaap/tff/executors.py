import uuid
import weakref
from typing import Mapping

import grpc
import tensorflow_federated
from absl import logging
from flaap import networking, taskstore_pb2, taskstore_pb2_grpc
from tensorflow_federated.proto.v0 import executor_pb2
from tensorflow_federated.python.common_libs import py_typecheck, tracing
from tensorflow_federated.python.core.impl.executors import (
    executor_base,
    executor_value_base,
    executors_errors,
)
from tensorflow_federated.python.core.impl.types import computation_types, placements

# N.B looks like value_serialization gets moved to executor_serialization in 0.34
if tensorflow_federated.__version__ < "0.34.0":
    from tensorflow_federated.python.core.impl.executors import value_serialization

    executor_serialization = value_serialization
else:
    from tensorflow_federated.python.core.impl.executors import executor_serialization


class TaskStoreExecutor(executor_base.Executor):
    """A TensorFlow federated executor that creates tasks in the task stor.

    Inspired by the Remote Executor:
    https://github.com/tensorflow/federated/blob/54ae7836c593746e3dd9a3ccfe74f61d46005c5c/tensorflow_federated/python/core/impl/executors/remote_executor.py

    Unlike the remote executor it doesn't batch requests to delete remote values but maybe it should.
    """

    def __init__(self, device=None, channel=None):
        """Creates a new instance of this executor.
        Args:
          channel: An instance of `grpc.Channel` to use for communication with the
            task store.
        Raises:
          TypeError: if arguments are of the wrong types.
        """
        py_typecheck.check_type(channel, grpc.Channel)

        logging.debug("Creating new TaskStoreExecutor")

        self._channel_status = False

        def _channel_status_callback(channel_connectivity: grpc.ChannelConnectivity):
            self._channel_status = channel_connectivity

        channel.subscribe(_channel_status_callback, try_to_connect=True)

        # We need to keep a reference to the channel around to prevent the Python
        # object from being GC'ed and the callback above from no-op'ing.
        self._channel = channel
        self._stub = taskstore_pb2_grpc.TasksServiceStub(channel)

        # Allow injection for unittesting.
        self._request_fn = _request

        # Generate a random nonce used to identify all the tasks created by this executor.
        # This will be used to ensure they all get assigned to the same client and a single
        # client doesn't claim more then its set of tasks.
        self._group_nonce = uuid.uuid4().hex

    @property
    def group_nonce(self):
        return self._group_nonce

    @property
    def is_ready(self) -> bool:
        return self._channel_status == grpc.ChannelConnectivity.READY

    def close(self):
        logging.info(
            "TODO(flaap/3) clearing tasks in the taskstore on close is not implemented."
        )

    def _dispose(self, name: str):
        """Dispose of the corresponding task."""
        delete_request = taskstore_pb2.DeleteRequest()
        delete_request.name = name
        self._request_fn(self._stub.Delete, delete_request)

    @tracing.trace(span=True)
    def set_cardinalities(
        self, cardinalities: Mapping[placements.PlacementLiteral, int]
    ):
        # TODO(jeremy): What should we do this for request? Just ignore it? It doesn't look like its
        # imlemented for EagerTFExecutor so maybe we don't need it for this one either
        # https://github.com/tensorflow/federated/blob/6fa4137d5485c119560a6363c705b76aa29237b1/tensorflow_federated/python/core/impl/executors/eager_tf_executor.py#L584
        #
        # I blieve a placement is a mapping from a placement e.g. CLIENTS or SERVER to integer. I believe the integer
        # is how many of that placement this executor handles; e.g. CLIENTS->N means this executor represents N clients.
        # Presumably N > 1 implies the executor knows how to handle more than 1 client. In this case, how would it know which computations
        # go to which executor? I suspect eager_tf_executor doesn't implement set_cardinalities because it is always a "leaf" executor
        # and there is a 1:1 mapping from eager_tf_executors to clients.
        raise NotImplementedError(
            "set_cardinalities isn't implemented for TaskStoreExecutor"
        )
        serialized_cardinalities = executor_serialization.serialize_cardinalities(
            cardinalities
        )
        request = executor_pb2.SetCardinalitiesRequest(
            cardinalities=serialized_cardinalities
        )

        _request(self._stub.SetCardinalities, request)

    @tracing.trace(span=True)
    async def create_value(self, value, type_spec=None):
        """Create value creates the value in the executor"""
        # Creating the value is a nullop. The value is returned
        # so that the caller will pass it along to create_call where it will
        # be serialized as part of the Task.
        return TaskInputValue(value, type_spec)

    @tracing.trace(span=True)
    async def create_call(self, comp, arg=None):
        """create_call creates a task containing the computation with the provided argument.

        Creating the task in the taskstore schedules it for execution.

        Args:
          computation: A value representing the AST to be run
          arg: Optional the value to be passed to the computation
        """

        py_typecheck.check_type(comp, TaskInputValue)
        # Comp needs to represent a function type as it is supposed to define the operations
        # to be run
        comp_proto, type_spec = executor_serialization.serialize_value(
            comp.value, comp.type_spec
        )
        py_typecheck.check_type(type_spec, computation_types.FunctionType)

        task = taskstore_pb2.Task()
        task.metadata.name = uuid.uuid4().hex
        task.input.function = comp_proto.SerializeToString()
        task.group_nonce = self._group_nonce

        if arg is not None:
            arg_proto, _ = executor_serialization.serialize_value(
                arg.value, arg.type_spec
            )
            task.input.argument = arg_proto.SerializeToString()

        create_task_request = taskstore_pb2.CreateRequest(task=task)

        response = self._request_fn(self._stub.Create, create_task_request)
        py_typecheck.check_type(response, taskstore_pb2.CreateResponse)
        return TaskValue(response.task.metadata.name, type_spec.result, self)

    @tracing.trace(span=True)
    async def create_struct(self, elements):
        # Currently, create_struct isn't implemented because it doesn't seem like this would be needed in
        # a scenario where workers call into the control plane to request and asynchronously process tasks.
        # In this asynchronous mode it doesn't seem like it makes sense to embed a struct in a remote worker
        # and then in a subsequent create a computation that references it. It seems like this should all be
        # bundled into a single task to be claimed and processed by the worker.
        raise NotImplementedError(
            "create_struct isn't supported with asynchronous execution via the task store"
        )

    @tracing.trace(span=True)
    async def create_selection(self, source, index):
        # Currently, create_selection isn't implemented because it doesn't seem like this would be needed in
        # a scenario where workers call into the control plane to request and asynchronously process tasks.
        # In this asynchronous mode it doesn't seem like it makes sense to create a selection in a remote worker
        # and try to asynchronously process it.
        raise NotImplementedError(
            "create_selection isn't supported with asynchronous execution via the task store"
        )

    @tracing.trace(span=True)
    async def _compute(self, name):
        """Compute waits for a given task to complete and then returns its value"""
        task = networking.wait_for_task(self._stub, name)
        value_pb = executor_pb2.Value()
        value_pb.ParseFromString(task.result)
        value, _ = executor_serialization.deserialize_value(value_pb)
        return value


# TODO(jeremy): should we generalize this?
@tracing.trace(span=True)
def _request(rpc_func, request):
    """Populates trace context and reraises gRPC errors with retryable info."""
    with tracing.wrap_rpc_in_trace_context():
        try:
            return rpc_func(request)
        except grpc.RpcError as e:
            if _is_retryable_grpc_error(e):
                logging.info("Received retryable gRPC error: %s", e)
                raise executors_errors.RetryableError(e)
            else:
                raise


def _is_retryable_grpc_error(error):
    """Predicate defining what is a retryable gRPC error."""
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
    }
    return isinstance(error, grpc.RpcError) and error.code() not in non_retryable_errors


class TaskInputValue(executor_value_base.ExecutorValue):
    """TaskInputValue is a wrapper around TFF values used for the output of create_value.

    Executors expect to call create_value to embed values in the executor. Those functions
    are expected to return subclasses of ExecutorValue which can then be passed along
    to the create_call.

    For the taskstore executor this just stores the arguments to create_value so
    that serialization can happen when create_call is invoked.
    """

    def __init__(self, value, type_spec):
        self.value = value
        self.type_spec = type_spec

    async def compute(self):
        raise NotImplementedError(
            "Compute is not expected to be called on TaskInputValue"
        )

    @property
    def type_signature(self):
        return self.type_spec


class TaskValue(executor_value_base.ExecutorValue):
    """Represents a Task to be computed.

    Inspired by: https://github.com/tensorflow/federated/blob/54ae7836c593746e3dd9a3ccfe74f61d46005c5c/tensorflow_federated/python/core/impl/executors/remote_executor.py#L163

    The main purpose of this is to handle cleanup following the pattern used by the
    RemoteExecutor. When the executor creates a Task in the taskstore to
    represent a value it returns an instance of RemoteValueTask. When that
    RemoteValueTask python object is deleted this should trigger a delete
    request for the task.
    """

    def __init__(self, name: str, type_spec, executor):
        """Creates the value.
        Args:
          name: Name of the created task
          type_spec: An instance of `computation_types.Type`.
          executor: The executor that created this value.
        """
        py_typecheck.check_type(type_spec, computation_types.Type)
        py_typecheck.check_type(executor, TaskStoreExecutor)
        self._name = name
        self._executor = executor
        self._type_signature = type_spec

        # Clean up the value and the memory associated with it on the remote
        # worker when no references to it remain.
        def finalizer(task_name, executor):
            executor._dispose(task_name)  # pylint: disable=protected-access

        weakref.finalize(self, finalizer, name, executor)

    @property
    def type_signature(self):
        return self._type_signature

    @tracing.trace(span=True)
    async def compute(self):
        return await self._executor._compute(
            self._name
        )  # pylint: disable=protected-access

    @property
    def name(self):
        return self._name
