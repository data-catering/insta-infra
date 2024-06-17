# Start Command

## Usage

```shell
./run.sh <services>
./run.sh postgres mysql
```

## Example Output

| Service  | Container To Container | Host To Container | Container To Host           |
|----------|-------------------------|-------------------|-----------------------------|
| postgres | postgres:5432           | localhost:5432    | host.docker.internal:5432   |
| mysql    | mysql:3306              | localhost:3306    | host.docker.internal:3306   |
