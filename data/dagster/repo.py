import os
from collections import Counter

from dagster import In, config_from_files, file_relative_path, graph, op, repository
from dagster_aws.s3 import s3_pickle_io_manager, s3_resource
from dagster_k8s import k8s_job_executor


@op(ins={"word": In(str)}, config_schema={"factor": int})
def multiply_the_word(context, word):
    return word * context.op_config["factor"]


@op(ins={"word": In(str)})
def count_letters(word):
    return dict(Counter(word))


@graph
def example_graph():
    count_letters(multiply_the_word())


example_job = example_graph.to_job(
    name="example_job",
    description="Example job. Use this to test your deployment.",
    config=config_from_files(
        [
            file_relative_path(__file__, os.path.join("run_config", "pipeline.yaml")),
        ]
    ),
)


pod_per_op_job = example_graph.to_job(
    name="pod_per_op_job",
    description="""
    Example job that uses the `k8s_job_executor` to run each op in a separate pod.
        
    **NOTE:** this job uses the s3_pickle_io_manager, which requires
    [AWS credentials](https://docs.dagster.io/deployment/guides/aws#using-s3-for-io-management).
    """,
    resource_defs={"s3": s3_resource, "io_manager": s3_pickle_io_manager},
    executor_def=k8s_job_executor,
    config=config_from_files(
        [
            file_relative_path(__file__, os.path.join("run_config", "k8s.yaml")),
            file_relative_path(__file__, os.path.join("run_config", "pipeline.yaml")),
        ]
    ),
)

@repository
def example_repo():
    return [example_job, pod_per_op_job]