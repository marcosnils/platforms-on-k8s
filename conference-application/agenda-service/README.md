# Agenda Service

## Endpoints

- `/` POST 
- `/` GET
- `/` DELETE
- `/{id}` GET
- `/{id}` DELETE
- `/highlights` GET
- `/day/{day}` GET
- `/health/readiness` GET
- `/health/liveness` GET

### Swagger

This application uses Swagger 2.0 using [swaggo/swag](https://github.com/swaggo/swag) tool.

Installing `swaggo/swag` tooling:

```shell
go install github.com/swaggo/swag/cmd/swag@latest
```

Update the documentation:

```shell
swag init -g agenda-service.go
```

Execute the application:

```shell
./agenda-service
```

Open your browser and access the following URL:
```http request
http://localhost:8080/swagger/index.html
```


## Build from source

## Create Container