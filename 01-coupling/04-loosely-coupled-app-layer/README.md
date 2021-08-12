# Loosely Coupled with Application Layer

The final example is based on the [loosely coupled service with code generation](../03-loosely-coupled-generated).

The last iteration moves application logic out of the HTTP handlers to a separate layer.

## Generating code

Generating both MySQL and OpenAPI code happens automatically when starting [docker-compose](../docker-compose.yml).

To regenerate the models, you can restart the containers:

```
docker-compose restart sqlboiler
docker-compose restart oapi-codegen
```

See details in [docker/sqlboiler](../docker/sqlboiler) and [docker/oapi-codegen](../docker/oapi-codegen).

![](https://threedots.tech/post/common-anti-patterns-in-go-web-applications/application-layer.png)
