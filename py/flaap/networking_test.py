from concurrent import futures

import grpc
import portpicker
import pytest
from flaap import networking, taskstore_pb2, taskstore_pb2_grpc


@pytest.mark.parametrize(
    "conditions,expected",
    [
        ({"succeeded": taskstore_pb2.TRUE}, True),
        ({"succeeded": taskstore_pb2.FALSE}, True),
        ({"succeeded": taskstore_pb2.UNKNOWN}, False),
        ({"other": taskstore_pb2.UNKNOWN}, False),
        ({"other": taskstore_pb2.TRUE}, False),
    ],
)
def test_is_done(conditions, expected):
    task = taskstore_pb2.Task()
    for k, v in conditions.items():
        condition = taskstore_pb2.Condition()
        condition.type = k
        condition.status = v
        task.status.conditions.append(condition)

    assert networking.task_is_done(task) == expected


class _TasksServicer(taskstore_pb2_grpc.TasksService):
    """Implement the TasksServicer for the wait_for_task_test"""

    def __init__(self):
        # Number of times get called
        self._get_call_index = 0
        # Tasks to return on each call to get task
        self._get_tasks = []

    def Get(self, request, context):
        self._get_call_index += 1
        t = self._get_tasks[self._get_call_index - 1]
        response = taskstore_pb2.GetResponse(task=t)
        return response

    def Create(self, request, context):
        raise NotImplementedError()

    def Delete(self, request, context):
        raise NotImplementedError()

    def Update(self, request, context):
        raise NotImplementedError()


def test_wait_for_task():
    port = portpicker.pick_unused_port()

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    server.add_insecure_port("[::]:{}".format(port))

    servicer = _TasksServicer()
    t = taskstore_pb2.Task()
    t.metadata.name = "task1"
    servicer._get_tasks.append(t)

    t = taskstore_pb2.Task()
    t.metadata.name = "task1"
    t.status.conditions.append(
        taskstore_pb2.Condition(type="Succeeded", status=taskstore_pb2.TRUE)
    )
    servicer._get_tasks.append(t)

    taskstore_pb2_grpc.add_TasksServiceServicer_to_server(servicer, server)
    server.start()

    with grpc.insecure_channel(f"localhost:{port}") as channel:
        stub = taskstore_pb2_grpc.TasksServiceStub(channel)
        task = networking.wait_for_task(stub, "task1")

    assert task.metadata.name == "task1"
