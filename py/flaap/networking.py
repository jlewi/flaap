import logging
import socket
from contextlib import closing

import grpc
import tenacity
from flaap import conditions, taskstore_pb2


# TODO(jeremy): Should we just use the portpicker library like TFF
# https://github.com/tensorflow/federated/blob/e1fef6445f6ff6cc3f2d5c16eb9c66fac8704f66/tensorflow_federated/python/core/impl/executors/remote_executor_test.py#L24
def find_free_port():
    """Find a free port"""
    with closing(socket.socket(socket.AF_INET, socket.SOCK_STREAM)) as s:
        s.bind(("", 0))
        s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        return s.getsockname()[1]


class NotDoneException(Exception):
    pass


def _is_retryable(exc):
    if isinstance(exc, NotDoneException):
        return True
    return _is_retryable_grpc_error(exc)


@tenacity.retry(
    retry=tenacity.retry_if_exception(_is_retryable),
    stop=tenacity.stop_after_delay(300),
)
def wait_for_task(stub, task_name):
    """Wait for task waits for the task to complete.

    Args:
      stub: The TaskServicesStub
      task_name: The name of the task to wait for

    Returns:
      Task:  Final task
    """
    logging.debug("Getting task =%s", task_name)
    request = taskstore_pb2.GetRequest(name=task_name)
    response = stub.Get(request)

    if conditions.is_done(response.task):
        return response.task
    else:
        raise NotDoneException()


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
