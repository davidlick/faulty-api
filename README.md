# faulty-api
An API which simulates loaded responses to demonstrate load testing.

## Getting Started

Some notes:
- A rate limit is initialized to a default of 10 concurrent requests.
- You may change this limit by making a `POST` to `http://localhost:8080/limit` with the following payload:
```
{"limit":10}
```
- When the rate limit is reached requests will be made to wait until others finish.
- Additionally, when the rate limit is exceeded faults will be injected into the system. The proportion of failed responses you can expect is based on the following formula:
```
PercFailedResponses = NumberRequestsOverLimit / Limit
```

## Spec

- `GET http://localhost:8080/data`
    - A basic response will be returned. The response time will be random between 0ms-500ms.
- `POST http://localhost:8080/limit`
    - Change the limit for allowed concurrent requests. The default is 10.
    
## Usage

Run from the `/cmd/api` directory:

```
go run *.go
```
