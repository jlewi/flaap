from concurrent import futures

import grpc
import portpicker
from flaap import networking, taskstore_pb2, taskstore_pb2_grpc


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
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Create not implemented")
        return taskstore_pb2.CreateResponse()

    def Delete(self, request, context):
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Delete not implemented")
        return taskstore_pb2.DeleteResponse()

    def Update(self, request, context):
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Update not implemented")
        return taskstore_pb2.UpdateResponse()


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
