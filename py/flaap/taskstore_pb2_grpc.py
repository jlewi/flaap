# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc

from flaap import taskstore_pb2 as flaap_dot_taskstore__pb2


class TasksServiceStub(object):
    """Tasks CRUD service
    """

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.Create = channel.unary_unary(
                '/flaap.v1alpha1.TasksService/Create',
                request_serializer=flaap_dot_taskstore__pb2.CreateRequest.SerializeToString,
                response_deserializer=flaap_dot_taskstore__pb2.CreateResponse.FromString,
                )
        self.Get = channel.unary_unary(
                '/flaap.v1alpha1.TasksService/Get',
                request_serializer=flaap_dot_taskstore__pb2.GetRequest.SerializeToString,
                response_deserializer=flaap_dot_taskstore__pb2.GetResponse.FromString,
                )
        self.Update = channel.unary_unary(
                '/flaap.v1alpha1.TasksService/Update',
                request_serializer=flaap_dot_taskstore__pb2.UpdateRequest.SerializeToString,
                response_deserializer=flaap_dot_taskstore__pb2.UpdateResponse.FromString,
                )
        self.Delete = channel.unary_unary(
                '/flaap.v1alpha1.TasksService/Delete',
                request_serializer=flaap_dot_taskstore__pb2.DeleteRequest.SerializeToString,
                response_deserializer=flaap_dot_taskstore__pb2.DeleteResponse.FromString,
                )


class TasksServiceServicer(object):
    """Tasks CRUD service
    """

    def Create(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Get(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Update(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Delete(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_TasksServiceServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'Create': grpc.unary_unary_rpc_method_handler(
                    servicer.Create,
                    request_deserializer=flaap_dot_taskstore__pb2.CreateRequest.FromString,
                    response_serializer=flaap_dot_taskstore__pb2.CreateResponse.SerializeToString,
            ),
            'Get': grpc.unary_unary_rpc_method_handler(
                    servicer.Get,
                    request_deserializer=flaap_dot_taskstore__pb2.GetRequest.FromString,
                    response_serializer=flaap_dot_taskstore__pb2.GetResponse.SerializeToString,
            ),
            'Update': grpc.unary_unary_rpc_method_handler(
                    servicer.Update,
                    request_deserializer=flaap_dot_taskstore__pb2.UpdateRequest.FromString,
                    response_serializer=flaap_dot_taskstore__pb2.UpdateResponse.SerializeToString,
            ),
            'Delete': grpc.unary_unary_rpc_method_handler(
                    servicer.Delete,
                    request_deserializer=flaap_dot_taskstore__pb2.DeleteRequest.FromString,
                    response_serializer=flaap_dot_taskstore__pb2.DeleteResponse.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'flaap.v1alpha1.TasksService', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


 # This class is part of an EXPERIMENTAL API.
class TasksService(object):
    """Tasks CRUD service
    """

    @staticmethod
    def Create(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/flaap.v1alpha1.TasksService/Create',
            flaap_dot_taskstore__pb2.CreateRequest.SerializeToString,
            flaap_dot_taskstore__pb2.CreateResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Get(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/flaap.v1alpha1.TasksService/Get',
            flaap_dot_taskstore__pb2.GetRequest.SerializeToString,
            flaap_dot_taskstore__pb2.GetResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Update(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/flaap.v1alpha1.TasksService/Update',
            flaap_dot_taskstore__pb2.UpdateRequest.SerializeToString,
            flaap_dot_taskstore__pb2.UpdateResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Delete(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/flaap.v1alpha1.TasksService/Delete',
            flaap_dot_taskstore__pb2.DeleteRequest.SerializeToString,
            flaap_dot_taskstore__pb2.DeleteResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)
