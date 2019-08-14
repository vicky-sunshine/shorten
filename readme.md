# Redis Shorten URL

[Work in Progress]
Use go and redis to implement basic shorten url service.

## Run

1. You need a redis service and run it up.

2. Run this service

    ```shell
    go run main.go
    ```

## API

### GET `/:shortid`

Given a short id, it will redirect to target url if found.
If not found, it will return 400.

### POST `/shorten`

#### request body example

```json
{
    "URL": "http://google.com"
}
```

#### response body example

```json
{
  "URL": "https://github.com/vicky-sunshine",
  "ID": "855215"
}
```
