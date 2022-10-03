"""Do a simple federated average using the taskstore.

It is based on: 
https://www.tensorflow.org/federated/tutorials/high_performance_simulation_with_kubernetes
https://www.tensorflow.org/federated/tutorials/custom_federated_algorithms_1

Per the tutorial. It relies on the lower level federate core (FC) interfaces.

The purpose of this simple example is to see whether we can successfully use
the TaskStoreExecutore to create tasks in the taskstore rather than 
sending them via RPC to remote workers.

"""
import logging

import fire
import grpc
import numpy as np
import tensorflow as tf
import tensorflow_federated as tff
from flaap import taskstore_pb2, taskstore_pb2_grpc
from flaap.tff import executors


# TODO(jeremy): This is duplicating code in federated_average; maybe we should just import it?
@tff.tf_computation(tff.SequenceType(tf.float32))
def get_local_temperature_average(local_temperatures):
    """This function uses TF to define the computations to be run on each worker.

    The argument of tf_computation defines the type of the input. Using tff.SequenceType
    means each worker will be instantiated with a sequence of float 32s. On
    the worker the instantiated type will be of tf.Dataset.

    """
    # The concrete type of local_temperatures on the workers will be a TF.Dataset.
    # We use the tf.Dataset reduce function to compute the local average for the
    # values of that worker.
    sum_and_count = local_temperatures.reduce(
        (0.0, 0), lambda x, y: (x[0] + y, x[1] + 1)
    )
    return sum_and_count[0] / tf.cast(sum_and_count[1], tf.float32)


@tff.federated_computation(tff.type_at_clients(tff.SequenceType(tf.float32)))
def get_global_temperature_average(sensor_readings):
    """ "This defines the federated computation."""
    # get_local_temperature_average operates on a single workers values; so to
    # apply it to all values in a federated dataset we use federated map
    return tff.federated_mean(
        tff.federated_map(get_local_temperature_average, sensor_readings)
    )


class Runner:
    @staticmethod
    def run(data=None, taskstore="localhost:8081", log_level="INFO"):
        """This runs the federated computation.
        Args:
          data is a list of lists. The outer list iterates over the list of devices
            represented by the group tff.CLIENTs. Inner list is the data for each device
            e.g.
            data[i] is a list containing the items for the i'th worker.
          taskstore: host:port of the gRPC service for the Taskstore API

        """

        logging.basicConfig(
            level=log_level.upper(),
            format=(
                "%(levelname)s|%(asctime)s" "|%(pathname)s|%(lineno)d| %(message)s"
            ),
            datefmt="%Y-%m-%dT%H:%M:%S",
        )
        logging.getLogger().setLevel(log_level.upper())
        logging.info("Running federated average; taskstore=%s", taskstore)

        # TODO(jeremy): Should we use with grpc.insecure_channel ... to wrap it in a context?
        # I removed the with to see if it was causing problems but I don't think that was the problem
        channel = grpc.insecure_channel(taskstore)

        # Make sure we can connect to the taskstore
        # TODO(jeremy): We should catch the gRPC error and report an approriate errore.
        logging.info("Checking connectivity to taskstore")
        stub = taskstore_pb2_grpc.TasksServiceStub(channel)
        try:
            stub.Status(taskstore_pb2.StatusRequest())
        except grpc.RpcError as e:
            logging.error("Could not connect to taskstore: %s", taskstore)
            raise

        # Change the executor to the TaskStoreExecutor so tasks are handled by creating tasks in
        # the task store. We need to create a factory function that will create the executor with the provided channel
        def ex_fn(device: tf.config.LogicalDevice) -> executors.TaskStoreExecutor:
            return executors.TaskStoreExecutor(channel=channel)

        factory = tff.framework.local_executor_factory(leaf_executor_fn=ex_fn)
        # If we don't override the executor function it works.
        # factory = tff.framework.local_executor_factory()
        ctx = tff.framework.ExecutionContext(executor_fn=factory)
        tff.framework.set_default_context(ctx)

        # Load default data if none is provided.
        if data is None:
            data = [[68.0, 70.0]]

        # Compute the expected value to confince ourselves
        # Note that the way the federated average is implemented each sample isn't properly weighted.
        # Instead the algorithm works by computing the average local to each worker and then averaging
        # the averages (so giving equal weight to the average from each worker)
        expected = np.mean([np.mean(x) for x in data])
        result = get_global_temperature_average(data)        
        # N.B.
        print(f"Actual={result} Properly Weighted={expected}")
        logging.info("Result=%s", result)
        logging.info("fed_average.py finished")

if __name__ == "__main__":

    fire.Fire(Runner)
