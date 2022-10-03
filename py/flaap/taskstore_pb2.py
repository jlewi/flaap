# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: flaap/taskstore.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x15\x66laap/taskstore.proto\x12\x0e\x66laap.v1alpha1\"\x96\x01\n\x08Metadata\x12\x0c\n\x04name\x18\x01 \x01(\t\x12\x34\n\x06labels\x18\x02 \x03(\x0b\x32$.flaap.v1alpha1.Metadata.LabelsEntry\x12\x17\n\x0fresourceVersion\x18\x03 \x01(\t\x1a-\n\x0bLabelsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01\"J\n\tCondition\x12\x0c\n\x04type\x18\x01 \x01(\t\x12/\n\x06status\x18\x02 \x01(\x0e\x32\x1f.flaap.v1alpha1.StatusCondition\";\n\nTaskStatus\x12-\n\nconditions\x18\x01 \x03(\x0b\x32\x19.flaap.v1alpha1.Condition\"\xfc\x01\n\x04Task\x12\x13\n\x0b\x61pi_version\x18\x01 \x01(\t\x12\x0c\n\x04kind\x18\x02 \x01(\t\x12*\n\x08metadata\x18\x03 \x01(\x0b\x32\x18.flaap.v1alpha1.Metadata\x12(\n\x05input\x18\x04 \x01(\x0b\x32\x19.flaap.v1alpha1.TaskInput\x12*\n\x06output\x18\x05 \x01(\x0b\x32\x1a.flaap.v1alpha1.TaskOutput\x12*\n\x06status\x18\x06 \x01(\x0b\x32\x1a.flaap.v1alpha1.TaskStatus\x12\x0e\n\x06result\x18\x07 \x01(\x0c\x12\x13\n\x0bgroup_nonce\x18\x08 \x01(\t\"\xc4\x01\n\tTaskInput\x12\x10\n\x08\x66unction\x18\x01 \x01(\x0c\x12\x10\n\x08\x61rgument\x18\x02 \x01(\x0c\x12\x16\n\x0c\x63reate_value\x18\x04 \x01(\x0cH\x00\x12\x15\n\x0b\x63reate_call\x18\x05 \x01(\x0cH\x00\x12\x17\n\rcreate_struct\x18\x06 \x01(\x0cH\x00\x12\x1a\n\x10\x63reate_selection\x18\x07 \x01(\x0cH\x00\x12\x11\n\x07\x63ompute\x18\x08 \x01(\x0cH\x00\x12\x11\n\x07\x64ispose\x18\t \x01(\x0cH\x00\x42\t\n\x07request\">\n\nTaskOutput\x12\x11\n\x07\x63ompute\x18\x05 \x01(\x0cH\x00\x12\x11\n\x07\x64ispose\x18\x06 \x01(\x0cH\x00\x42\n\n\x08response\"3\n\rCreateRequest\x12\"\n\x04task\x18\x01 \x01(\x0b\x32\x14.flaap.v1alpha1.Task\"4\n\x0e\x43reateResponse\x12\"\n\x04task\x18\x01 \x01(\x0b\x32\x14.flaap.v1alpha1.Task\"\x1a\n\nGetRequest\x12\x0c\n\x04name\x18\x01 \x01(\t\"1\n\x0bGetResponse\x12\"\n\x04task\x18\x01 \x01(\x0b\x32\x14.flaap.v1alpha1.Task\".\n\x0bListRequest\x12\x11\n\tworker_id\x18\x01 \x01(\t\x12\x0c\n\x04\x64one\x18\x02 \x01(\x08\"3\n\x0cListResponse\x12#\n\x05items\x18\x01 \x03(\x0b\x32\x14.flaap.v1alpha1.Task\"\x7f\n\x08Selector\x12?\n\x0cmatch_labels\x18\x01 \x03(\x0b\x32).flaap.v1alpha1.Selector.MatchLabelsEntry\x1a\x32\n\x10MatchLabelsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01\"F\n\rUpdateRequest\x12\"\n\x04task\x18\x01 \x01(\x0b\x32\x14.flaap.v1alpha1.Task\x12\x11\n\tworker_id\x18\x02 \x01(\t\"4\n\x0eUpdateResponse\x12\"\n\x04task\x18\x01 \x01(\x0b\x32\x14.flaap.v1alpha1.Task\"\x1d\n\rDeleteRequest\x12\x0c\n\x04name\x18\x01 \x01(\t\"\x10\n\x0e\x44\x65leteResponse\"\x0f\n\rStatusRequest\"\x95\x02\n\x0eStatusResponse\x12O\n\x11group_assignments\x18\x01 \x03(\x0b\x32\x34.flaap.v1alpha1.StatusResponse.GroupAssignmentsEntry\x12\x45\n\x0ctask_metrics\x18\x02 \x03(\x0b\x32/.flaap.v1alpha1.StatusResponse.TaskMetricsEntry\x1a\x37\n\x15GroupAssignmentsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01\x1a\x32\n\x10TaskMetricsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\x05:\x02\x38\x01\"\xb7\x01\n\nStoredData\x12#\n\x05tasks\x18\x01 \x03(\x0b\x32\x14.flaap.v1alpha1.Task\x12K\n\x11group_assignments\x18\x02 \x03(\x0b\x32\x30.flaap.v1alpha1.StoredData.GroupAssignmentsEntry\x1a\x37\n\x15GroupAssignmentsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01*3\n\x0fStatusCondition\x12\x0b\n\x07UNKNOWN\x10\x00\x12\t\n\x05\x46\x41LSE\x10\x01\x12\x08\n\x04TRUE\x10\x02\x32\xc1\x03\n\x0cTasksService\x12I\n\x06\x43reate\x12\x1d.flaap.v1alpha1.CreateRequest\x1a\x1e.flaap.v1alpha1.CreateResponse\"\x00\x12@\n\x03Get\x12\x1a.flaap.v1alpha1.GetRequest\x1a\x1b.flaap.v1alpha1.GetResponse\"\x00\x12\x43\n\x04List\x12\x1b.flaap.v1alpha1.ListRequest\x1a\x1c.flaap.v1alpha1.ListResponse\"\x00\x12I\n\x06Update\x12\x1d.flaap.v1alpha1.UpdateRequest\x1a\x1e.flaap.v1alpha1.UpdateResponse\"\x00\x12I\n\x06\x44\x65lete\x12\x1d.flaap.v1alpha1.DeleteRequest\x1a\x1e.flaap.v1alpha1.DeleteResponse\"\x00\x12I\n\x06Status\x12\x1d.flaap.v1alpha1.StatusRequest\x1a\x1e.flaap.v1alpha1.StatusResponse\"\x00\x42+Z)github.com/jlewi/flaap/go/protos/v1alpha1b\x06proto3')

_STATUSCONDITION = DESCRIPTOR.enum_types_by_name['StatusCondition']
StatusCondition = enum_type_wrapper.EnumTypeWrapper(_STATUSCONDITION)
UNKNOWN = 0
FALSE = 1
TRUE = 2


_METADATA = DESCRIPTOR.message_types_by_name['Metadata']
_METADATA_LABELSENTRY = _METADATA.nested_types_by_name['LabelsEntry']
_CONDITION = DESCRIPTOR.message_types_by_name['Condition']
_TASKSTATUS = DESCRIPTOR.message_types_by_name['TaskStatus']
_TASK = DESCRIPTOR.message_types_by_name['Task']
_TASKINPUT = DESCRIPTOR.message_types_by_name['TaskInput']
_TASKOUTPUT = DESCRIPTOR.message_types_by_name['TaskOutput']
_CREATEREQUEST = DESCRIPTOR.message_types_by_name['CreateRequest']
_CREATERESPONSE = DESCRIPTOR.message_types_by_name['CreateResponse']
_GETREQUEST = DESCRIPTOR.message_types_by_name['GetRequest']
_GETRESPONSE = DESCRIPTOR.message_types_by_name['GetResponse']
_LISTREQUEST = DESCRIPTOR.message_types_by_name['ListRequest']
_LISTRESPONSE = DESCRIPTOR.message_types_by_name['ListResponse']
_SELECTOR = DESCRIPTOR.message_types_by_name['Selector']
_SELECTOR_MATCHLABELSENTRY = _SELECTOR.nested_types_by_name['MatchLabelsEntry']
_UPDATEREQUEST = DESCRIPTOR.message_types_by_name['UpdateRequest']
_UPDATERESPONSE = DESCRIPTOR.message_types_by_name['UpdateResponse']
_DELETEREQUEST = DESCRIPTOR.message_types_by_name['DeleteRequest']
_DELETERESPONSE = DESCRIPTOR.message_types_by_name['DeleteResponse']
_STATUSREQUEST = DESCRIPTOR.message_types_by_name['StatusRequest']
_STATUSRESPONSE = DESCRIPTOR.message_types_by_name['StatusResponse']
_STATUSRESPONSE_GROUPASSIGNMENTSENTRY = _STATUSRESPONSE.nested_types_by_name['GroupAssignmentsEntry']
_STATUSRESPONSE_TASKMETRICSENTRY = _STATUSRESPONSE.nested_types_by_name['TaskMetricsEntry']
_STOREDDATA = DESCRIPTOR.message_types_by_name['StoredData']
_STOREDDATA_GROUPASSIGNMENTSENTRY = _STOREDDATA.nested_types_by_name['GroupAssignmentsEntry']
Metadata = _reflection.GeneratedProtocolMessageType('Metadata', (_message.Message,), {

  'LabelsEntry' : _reflection.GeneratedProtocolMessageType('LabelsEntry', (_message.Message,), {
    'DESCRIPTOR' : _METADATA_LABELSENTRY,
    '__module__' : 'flaap.taskstore_pb2'
    # @@protoc_insertion_point(class_scope:flaap.v1alpha1.Metadata.LabelsEntry)
    })
  ,
  'DESCRIPTOR' : _METADATA,
  '__module__' : 'flaap.taskstore_pb2'
  # @@protoc_insertion_point(class_scope:flaap.v1alpha1.Metadata)
  })
_sym_db.RegisterMessage(Metadata)
_sym_db.RegisterMessage(Metadata.LabelsEntry)

Condition = _reflection.GeneratedProtocolMessageType('Condition', (_message.Message,), {
  'DESCRIPTOR' : _CONDITION,
  '__module__' : 'flaap.taskstore_pb2'
  # @@protoc_insertion_point(class_scope:flaap.v1alpha1.Condition)
  })
_sym_db.RegisterMessage(Condition)

TaskStatus = _reflection.GeneratedProtocolMessageType('TaskStatus', (_message.Message,), {
  'DESCRIPTOR' : _TASKSTATUS,
  '__module__' : 'flaap.taskstore_pb2'
  # @@protoc_insertion_point(class_scope:flaap.v1alpha1.TaskStatus)
  })
_sym_db.RegisterMessage(TaskStatus)

Task = _reflection.GeneratedProtocolMessageType('Task', (_message.Message,), {
  'DESCRIPTOR' : _TASK,
  '__module__' : 'flaap.taskstore_pb2'
  # @@protoc_insertion_point(class_scope:flaap.v1alpha1.Task)
  })
_sym_db.RegisterMessage(Task)

TaskInput = _reflection.GeneratedProtocolMessageType('TaskInput', (_message.Message,), {
  'DESCRIPTOR' : _TASKINPUT,
  '__module__' : 'flaap.taskstore_pb2'
  # @@protoc_insertion_point(class_scope:flaap.v1alpha1.TaskInput)
  })
_sym_db.RegisterMessage(TaskInput)

TaskOutput = _reflection.GeneratedProtocolMessageType('TaskOutput', (_message.Message,), {
  'DESCRIPTOR' : _TASKOUTPUT,
  '__module__' : 'flaap.taskstore_pb2'
  # @@protoc_insertion_point(class_scope:flaap.v1alpha1.TaskOutput)
  })
_sym_db.RegisterMessage(TaskOutput)

CreateRequest = _reflection.GeneratedProtocolMessageType('CreateRequest', (_message.Message,), {
  'DESCRIPTOR' : _CREATEREQUEST,
  '__module__' : 'flaap.taskstore_pb2'
  # @@protoc_insertion_point(class_scope:flaap.v1alpha1.CreateRequest)
  })
_sym_db.RegisterMessage(CreateRequest)

CreateResponse = _reflection.GeneratedProtocolMessageType('CreateResponse', (_message.Message,), {
  'DESCRIPTOR' : _CREATERESPONSE,
  '__module__' : 'flaap.taskstore_pb2'
  # @@protoc_insertion_point(class_scope:flaap.v1alpha1.CreateResponse)
  })
_sym_db.RegisterMessage(CreateResponse)

GetRequest = _reflection.GeneratedProtocolMessageType('GetRequest', (_message.Message,), {
  'DESCRIPTOR' : _GETREQUEST,
  '__module__' : 'flaap.taskstore_pb2'
  # @@protoc_insertion_point(class_scope:flaap.v1alpha1.GetRequest)
  })
_sym_db.RegisterMessage(GetRequest)

GetResponse = _reflection.GeneratedProtocolMessageType('GetResponse', (_message.Message,), {
  'DESCRIPTOR' : _GETRESPONSE,
  '__module__' : 'flaap.taskstore_pb2'
  # @@protoc_insertion_point(class_scope:flaap.v1alpha1.GetResponse)
  })
_sym_db.RegisterMessage(GetResponse)

ListRequest = _reflection.GeneratedProtocolMessageType('ListRequest', (_message.Message,), {
  'DESCRIPTOR' : _LISTREQUEST,
  '__module__' : 'flaap.taskstore_pb2'
  # @@protoc_insertion_point(class_scope:flaap.v1alpha1.ListRequest)
  })
_sym_db.RegisterMessage(ListRequest)

ListResponse = _reflection.GeneratedProtocolMessageType('ListResponse', (_message.Message,), {
  'DESCRIPTOR' : _LISTRESPONSE,
  '__module__' : 'flaap.taskstore_pb2'
  # @@protoc_insertion_point(class_scope:flaap.v1alpha1.ListResponse)
  })
_sym_db.RegisterMessage(ListResponse)

Selector = _reflection.GeneratedProtocolMessageType('Selector', (_message.Message,), {

  'MatchLabelsEntry' : _reflection.GeneratedProtocolMessageType('MatchLabelsEntry', (_message.Message,), {
    'DESCRIPTOR' : _SELECTOR_MATCHLABELSENTRY,
    '__module__' : 'flaap.taskstore_pb2'
    # @@protoc_insertion_point(class_scope:flaap.v1alpha1.Selector.MatchLabelsEntry)
    })
  ,
  'DESCRIPTOR' : _SELECTOR,
  '__module__' : 'flaap.taskstore_pb2'
  # @@protoc_insertion_point(class_scope:flaap.v1alpha1.Selector)
  })
_sym_db.RegisterMessage(Selector)
_sym_db.RegisterMessage(Selector.MatchLabelsEntry)

UpdateRequest = _reflection.GeneratedProtocolMessageType('UpdateRequest', (_message.Message,), {
  'DESCRIPTOR' : _UPDATEREQUEST,
  '__module__' : 'flaap.taskstore_pb2'
  # @@protoc_insertion_point(class_scope:flaap.v1alpha1.UpdateRequest)
  })
_sym_db.RegisterMessage(UpdateRequest)

UpdateResponse = _reflection.GeneratedProtocolMessageType('UpdateResponse', (_message.Message,), {
  'DESCRIPTOR' : _UPDATERESPONSE,
  '__module__' : 'flaap.taskstore_pb2'
  # @@protoc_insertion_point(class_scope:flaap.v1alpha1.UpdateResponse)
  })
_sym_db.RegisterMessage(UpdateResponse)

DeleteRequest = _reflection.GeneratedProtocolMessageType('DeleteRequest', (_message.Message,), {
  'DESCRIPTOR' : _DELETEREQUEST,
  '__module__' : 'flaap.taskstore_pb2'
  # @@protoc_insertion_point(class_scope:flaap.v1alpha1.DeleteRequest)
  })
_sym_db.RegisterMessage(DeleteRequest)

DeleteResponse = _reflection.GeneratedProtocolMessageType('DeleteResponse', (_message.Message,), {
  'DESCRIPTOR' : _DELETERESPONSE,
  '__module__' : 'flaap.taskstore_pb2'
  # @@protoc_insertion_point(class_scope:flaap.v1alpha1.DeleteResponse)
  })
_sym_db.RegisterMessage(DeleteResponse)

StatusRequest = _reflection.GeneratedProtocolMessageType('StatusRequest', (_message.Message,), {
  'DESCRIPTOR' : _STATUSREQUEST,
  '__module__' : 'flaap.taskstore_pb2'
  # @@protoc_insertion_point(class_scope:flaap.v1alpha1.StatusRequest)
  })
_sym_db.RegisterMessage(StatusRequest)

StatusResponse = _reflection.GeneratedProtocolMessageType('StatusResponse', (_message.Message,), {

  'GroupAssignmentsEntry' : _reflection.GeneratedProtocolMessageType('GroupAssignmentsEntry', (_message.Message,), {
    'DESCRIPTOR' : _STATUSRESPONSE_GROUPASSIGNMENTSENTRY,
    '__module__' : 'flaap.taskstore_pb2'
    # @@protoc_insertion_point(class_scope:flaap.v1alpha1.StatusResponse.GroupAssignmentsEntry)
    })
  ,

  'TaskMetricsEntry' : _reflection.GeneratedProtocolMessageType('TaskMetricsEntry', (_message.Message,), {
    'DESCRIPTOR' : _STATUSRESPONSE_TASKMETRICSENTRY,
    '__module__' : 'flaap.taskstore_pb2'
    # @@protoc_insertion_point(class_scope:flaap.v1alpha1.StatusResponse.TaskMetricsEntry)
    })
  ,
  'DESCRIPTOR' : _STATUSRESPONSE,
  '__module__' : 'flaap.taskstore_pb2'
  # @@protoc_insertion_point(class_scope:flaap.v1alpha1.StatusResponse)
  })
_sym_db.RegisterMessage(StatusResponse)
_sym_db.RegisterMessage(StatusResponse.GroupAssignmentsEntry)
_sym_db.RegisterMessage(StatusResponse.TaskMetricsEntry)

StoredData = _reflection.GeneratedProtocolMessageType('StoredData', (_message.Message,), {

  'GroupAssignmentsEntry' : _reflection.GeneratedProtocolMessageType('GroupAssignmentsEntry', (_message.Message,), {
    'DESCRIPTOR' : _STOREDDATA_GROUPASSIGNMENTSENTRY,
    '__module__' : 'flaap.taskstore_pb2'
    # @@protoc_insertion_point(class_scope:flaap.v1alpha1.StoredData.GroupAssignmentsEntry)
    })
  ,
  'DESCRIPTOR' : _STOREDDATA,
  '__module__' : 'flaap.taskstore_pb2'
  # @@protoc_insertion_point(class_scope:flaap.v1alpha1.StoredData)
  })
_sym_db.RegisterMessage(StoredData)
_sym_db.RegisterMessage(StoredData.GroupAssignmentsEntry)

_TASKSSERVICE = DESCRIPTOR.services_by_name['TasksService']
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'Z)github.com/jlewi/flaap/go/protos/v1alpha1'
  _METADATA_LABELSENTRY._options = None
  _METADATA_LABELSENTRY._serialized_options = b'8\001'
  _SELECTOR_MATCHLABELSENTRY._options = None
  _SELECTOR_MATCHLABELSENTRY._serialized_options = b'8\001'
  _STATUSRESPONSE_GROUPASSIGNMENTSENTRY._options = None
  _STATUSRESPONSE_GROUPASSIGNMENTSENTRY._serialized_options = b'8\001'
  _STATUSRESPONSE_TASKMETRICSENTRY._options = None
  _STATUSRESPONSE_TASKMETRICSENTRY._serialized_options = b'8\001'
  _STOREDDATA_GROUPASSIGNMENTSENTRY._options = None
  _STOREDDATA_GROUPASSIGNMENTSENTRY._serialized_options = b'8\001'
  _STATUSCONDITION._serialized_start=1923
  _STATUSCONDITION._serialized_end=1974
  _METADATA._serialized_start=42
  _METADATA._serialized_end=192
  _METADATA_LABELSENTRY._serialized_start=147
  _METADATA_LABELSENTRY._serialized_end=192
  _CONDITION._serialized_start=194
  _CONDITION._serialized_end=268
  _TASKSTATUS._serialized_start=270
  _TASKSTATUS._serialized_end=329
  _TASK._serialized_start=332
  _TASK._serialized_end=584
  _TASKINPUT._serialized_start=587
  _TASKINPUT._serialized_end=783
  _TASKOUTPUT._serialized_start=785
  _TASKOUTPUT._serialized_end=847
  _CREATEREQUEST._serialized_start=849
  _CREATEREQUEST._serialized_end=900
  _CREATERESPONSE._serialized_start=902
  _CREATERESPONSE._serialized_end=954
  _GETREQUEST._serialized_start=956
  _GETREQUEST._serialized_end=982
  _GETRESPONSE._serialized_start=984
  _GETRESPONSE._serialized_end=1033
  _LISTREQUEST._serialized_start=1035
  _LISTREQUEST._serialized_end=1081
  _LISTRESPONSE._serialized_start=1083
  _LISTRESPONSE._serialized_end=1134
  _SELECTOR._serialized_start=1136
  _SELECTOR._serialized_end=1263
  _SELECTOR_MATCHLABELSENTRY._serialized_start=1213
  _SELECTOR_MATCHLABELSENTRY._serialized_end=1263
  _UPDATEREQUEST._serialized_start=1265
  _UPDATEREQUEST._serialized_end=1335
  _UPDATERESPONSE._serialized_start=1337
  _UPDATERESPONSE._serialized_end=1389
  _DELETEREQUEST._serialized_start=1391
  _DELETEREQUEST._serialized_end=1420
  _DELETERESPONSE._serialized_start=1422
  _DELETERESPONSE._serialized_end=1438
  _STATUSREQUEST._serialized_start=1440
  _STATUSREQUEST._serialized_end=1455
  _STATUSRESPONSE._serialized_start=1458
  _STATUSRESPONSE._serialized_end=1735
  _STATUSRESPONSE_GROUPASSIGNMENTSENTRY._serialized_start=1628
  _STATUSRESPONSE_GROUPASSIGNMENTSENTRY._serialized_end=1683
  _STATUSRESPONSE_TASKMETRICSENTRY._serialized_start=1685
  _STATUSRESPONSE_TASKMETRICSENTRY._serialized_end=1735
  _STOREDDATA._serialized_start=1738
  _STOREDDATA._serialized_end=1921
  _STOREDDATA_GROUPASSIGNMENTSENTRY._serialized_start=1628
  _STOREDDATA_GROUPASSIGNMENTSENTRY._serialized_end=1683
  _TASKSSERVICE._serialized_start=1977
  _TASKSSERVICE._serialized_end=2426
# @@protoc_insertion_point(module_scope)
