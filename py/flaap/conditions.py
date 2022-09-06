"""Utility routines for working with conditions."""

import typing

from flaap import taskstore_pb2

# Constants for various conditions
SUCCEEDED = "succeeded"


def is_done(resource) -> bool:
    """Return true if the resource has finished (whether successfully or unsuccessfully)

    Tasks should follow the Knative convention of having a succeeded condition.
    https://github.com/knative/specs/blob/main/specs/common/error-signalling.md

    Returns:
      True if the task is done and false otherwise.
    """
    status = get(resource, SUCCEEDED)
    if status is None:
        # There is no succeeded condition
        return False

    return status != taskstore_pb2.UNKNOWN


def get(resource, name: str) -> typing.Union[None, taskstore_pb2.Condition]:
    """Get the specified condition or None if the condition isn't defined"""
    # TODO(jeremy): Should we really be ignoring case here?
    name = name.lower()
    for c in resource.status.conditions:
        if c.type.lower() == name:
            return c.status
    return None


def set(resource, condition: str, status: taskstore_pb2.StatusCondition):
    """Set the specified condition"""
    condition = condition.lower()
    for index, c in enumerate(resource.status.conditions):
        if c.type.lower() == condition:
            resource.status.conditions[index].status = status
            return
    new_condition = taskstore_pb2.Condition()
    new_condition.type = condition
    new_condition.status = status
    resource.status.conditions.append(new_condition)
