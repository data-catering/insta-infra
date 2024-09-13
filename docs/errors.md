# Errors

## network insta-infra_default was found but has incorrect label com.docker.compose.network set to ""

This is due to insta-infra previously using `docker-compose` command which is part of V1 which is now deprecated and now
V2 uses `docker compose`. To resolve, run:
```shell
docker network rm insta-infra_default
```
Then try run insta-infra again.
