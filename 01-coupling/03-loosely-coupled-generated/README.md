# Loosely Coupled with Code Generation

This example improves further over the [loosely coupled service](../02-loosely-coupled).

To address the issue of writing boilerplate manually, we use [oapi-codegen](https://github.com/deepmap/oapi-codegen) to generate HTTP models and routes, and
[sqlboiler](https://github.com/volatiletech/sqlboiler) to generate MySQL models.

## Generating code

Generating both MySQL and OpenAPI code happens automatically when starting [docker-compose](../docker-compose.yml).

To regenerate the models, you can restart the containers:

```
docker-compose restart sqlboiler
docker-compose restart oapi-codegen
```

See details in [docker/sqlboiler](../docker/sqlboiler) and [docker/oapi-codegen](../docker/oapi-codegen).