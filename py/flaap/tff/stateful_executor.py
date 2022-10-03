"""A wrapper for another executor that maintains state."""


from tensorflow_federated.python.common_libs import py_typecheck, tracing
from tensorflow_federated.python.core.impl.executors import executor_base

# def _set_invalid_arg_err(context: grpc.ServicerContext, err):
#   logging.error(traceback.format_exc())
#   context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
#   context.set_details(str(err))


# def _set_unknown_err(context: grpc.ServicerContext, err):
#   logging.error(traceback.format_exc())
#   context.set_code(grpc.StatusCode.UNKNOWN)
#   context.set_details(str(err))


# def _set_unavailable_error(context: grpc.ServicerContext, err):
#   logging.error(traceback.format_exc())
#   context.set_code(grpc.StatusCode.UNAVAILABLE)
#   context.set_details(str(err))


# def _propagate_grpc_code_err(context: grpc.ServicerContext, err: grpc.RpcError):
#   logging.error(traceback.format_exc())
#   context.set_code(err.code())
#   context.set_details(str(err))


# def _get_hashable_key(cardinalities: executor_factory.CardinalitiesType) -> str:
#   return str(tuple(sorted((str(k), v) for k, v in cardinalities.items())))


class StatefulWrapper:
    """A wrapper around a target executor to keep track of state.

    When running in a remote production value; clients need to keep track of
    their values because TFFs protocol is stateful. So values resulting
    from one computation shouldn't necessarily be materialized back
    to the control plane. Instead the should be cached and potentially
    used in subsequent computations.

    Implementation is inspired by the ExecutorService:
    https://github.com/tensorflow/federated/blob/a6506385def304c260a424b29e5b34c6d905760e/tensorflow_federated/python/core/impl/executors/executor_service.py#L73

    One differe is the wrapper doesn't maintain a mapping to multiple
    executors corresponding to different cardinalities. If that's needed it should
    potentially be handled at a higher level.

    This doesn't implement the executor_base.Executor interface because
    unlike Executor some methods need to take in the name under which to cache the
    result and not return any value.
    """

    def __init__(self, target_executor: executor_base.Executor, *args, **kwargs):
        """
        Args:
          target_executor: The executor to delegate to.
        """
        py_typecheck.check_type(target_executor, executor_base.Executor)
        # The keys in this dictionary are value ids and the values are the values
        # returned by the target executor.
        self._values = {}
        self._target_executor = target_executor

    @tracing.trace(span=True)
    async def create_value(self, name, value, type_spec=None):
        """Creates a value embedded in the executor.

        Arg:
          name: Name to use to store the resulting value
          value: Value to create
        """
        # Pass actual value along to the executor
        result = await self._target_executor.create_value(value, type_spec)

        # Cache the result using the supplied name
        self._values[name] = result

    @tracing.trace(span=True)
    async def create_call(self, name, comp_name, arg_name=None):
        """Creates a call embedded in the executor.

        Args:
          name: Name to store the created call as
          comp_name: Name of the value for the computation
          arg_name: Optional the value to be passed to the computation
        """
        comp = self._values[comp_name]
        arg = None
        if arg_name is not None:
            arg = self._values[arg_name]
        # Create_call creates a callable. We invoke that callable to
        # produce the result which we cache.
        # EagerExecutor.create_call returns a EagerValue; the actual value
        # can be obtained by calling compute. We store EagerValue rather than
        # the result of compute because EagerValue includes type information.
        self._values[name] = await self._target_executor.create_call(comp, arg)

    @tracing.trace(span=True)
    async def create_struct(self, name, element_names):
        elements = []
        for n in element_names:
            elements = elements.append(self._values[n])
        self._values[name] = self._target_executor.create_struct(elements)

    @tracing.trace(span=True)
    async def create_selection(self, name, source_name, index):
        """Creates a call embedded in the executor.

        Args:
          name: Name to store the created selection as
          source_name: Name of the value for the selection
          index: the index
        """
        source = self._values[source_name]
        self._values[name] = self._target_executor.create_selection(source, index)

    def get_value(self, name):
        """Return the specified value"""
        return self._values[name]


# class StatefulValue(executor_value_base.ExecutorValue):
#   """Represent a stateful value."""
#   def __init__(self, name: str, value):
#       """Creates the value.
#       Args:
#         name: Name used to identify the value
#         value: actual value
#       """
#       # py_typecheck.check_type(type_spec, computation_types.Type)
#       # py_typecheck.check_type(executor, TaskStoreExecutor)
#       self.name = name
#       self._value = value
#       # self._type_signature = type_spec

#       # # Clean up the value and the memory associated with it on the remote
#       # # worker when no references to it remain.
#       # def finalizer(task_name, executor):
#       #     executor._dispose(task_name)  # pylint: disable=protected-access

#       # weakref.finalize(self, finalizer, name, executor)

#   @property
#   def type_signature(self):
#     # TODO(jeremy): Is this right?
#     return self._value.type_signature
