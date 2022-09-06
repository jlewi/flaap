import pytest
from flaap import conditions, taskstore_pb2


@pytest.mark.parametrize(
    "test_conditions,expected",
    [
        ({"succeeded": taskstore_pb2.TRUE}, True),
        ({"succeeded": taskstore_pb2.FALSE}, True),
        ({"succeeded": taskstore_pb2.UNKNOWN}, False),
        ({"other": taskstore_pb2.UNKNOWN}, False),
        ({"other": taskstore_pb2.TRUE}, False),
    ],
)
def test_is_done(test_conditions, expected):
    task = taskstore_pb2.Task()
    for k, v in test_conditions.items():
        condition = taskstore_pb2.Condition()
        condition.type = k
        condition.status = v
        task.status.conditions.append(condition)

    assert conditions.is_done(task) == expected


@pytest.mark.parametrize(
    "test_conditions,name,expected",
    [
        ({"succeeded": taskstore_pb2.TRUE}, "succeeded", taskstore_pb2.TRUE),
        ({"succeeded": taskstore_pb2.FALSE}, "succeeded", taskstore_pb2.FALSE),
        ({"succeeded": taskstore_pb2.UNKNOWN}, "succeeded", taskstore_pb2.UNKNOWN),
        ({"other": taskstore_pb2.UNKNOWN}, "succeeded", None),
    ],
)
def test_get(test_conditions, name, expected):
    task = taskstore_pb2.Task()
    for k, v in test_conditions.items():
        condition = taskstore_pb2.Condition()
        condition.type = k
        condition.status = v
        task.status.conditions.append(condition)

    assert conditions.get(task, name) == expected


@pytest.mark.parametrize(
    "test_conditions,name,expected",
    [
        ({"succeeded": taskstore_pb2.TRUE}, "succeeded", taskstore_pb2.TRUE),
        ({"succeeded": taskstore_pb2.TRUE}, "succeeded", taskstore_pb2.FALSE),
        ({"succeeded": taskstore_pb2.UNKNOWN}, "succeeded", taskstore_pb2.FALSE),
        ({"other": taskstore_pb2.TRUE}, "succeeded", taskstore_pb2.TRUE),
    ],
)
def test_set(test_conditions, name, expected):
    task = taskstore_pb2.Task()
    for k, v in test_conditions.items():
        condition = taskstore_pb2.Condition()
        condition.type = k
        condition.status = v
        task.status.conditions.append(condition)

    conditions.set(task, name, expected)
    assert conditions.get(task, name) == expected
