# Getting Started

## Installation

To install `insta-infra`, follow these steps:

1. Clone the repository:

```shell
git clone https://github.com/data-catering/insta-infra.git
```

1. Navigate to the directory:

```shell
cd insta-infra
```

## Starting Services

To start the services, use the following command:

```shell
./insta <services>
./insta postgres mysql
```

## Example Output

When you start the services, you'll see an output like this, which shows how to connect to each service:

| Service  | Container To Container | Host To Container | Container To Host         |
| -------- | ---------------------- | ----------------- | ------------------------- |
| postgres | postgres:5432          | localhost:5432    | host.docker.internal:5432 |
| mysql    | mysql:3306             | localhost:3306    | host.docker.internal:3306 |
