# Comrade

HTTP based queueing service

* take jobs
* deliver jobs


## Take Jobs

```
Request Line: POST /queue
Request Body: {"payload": {}}
Response Body: {"job-id": uuid, "request-id": uuid}
```

Optional queue name

```
Request Line: POST /queue/:name
Request Body: {"payload": {}}
Response Body: {"job-id": uuid, "request-id": uuid}
```

## Get Jobs

```
Request Line: GET /queue/jobs
Request Body: {}
Response Body: {"job-id": uuid, "request-id": uuid, "payload": {}}
```

Optional queue name


```
Request Line: GET /queue/:name/jobs
Request Body: {}
Response Body: {"job-id": uuid, "request-id": uuid, "payload": {}}
```
