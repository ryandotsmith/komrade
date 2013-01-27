# Komrade

HTTP based queueing service.

* PUT jobs
* GET jobs
* DELETE jobs

## PUT Job

```
> PUT /jobs/:id
> Authorization: b64
> {"payload": {}}
< {"id": id, "request-id": uuid}
```

### Increment Failed Count

```
> PUT /jobs/:id/failures/:failure_id
> Authorization: b64
< {"job-id": id, "request-id": uuid}
```


## GET Jobs

```
> GET /jobs?limit=1
> Authorization: b64
< [{"job-id": id, "request-id": uuid, "payload": {}}]
```

## DELETE Job

```
> DELETE /jobs/:id
> Authorization: b64
< [{"job-id": id, "request-id": uuid, "payload": {}}]
```
