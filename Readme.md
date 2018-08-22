# PostgreSQL Bridge

Simple bridge for [PostgreSQL notifications](https://www.postgresql.org/docs/9.0/static/sql-notify.html)

## Bridge Support

  - PostgreSQL → [Amazon SNS](https://aws.amazon.com/sns/)
  - PostgreSQL → [HTTP Webhooks](https://requestb.in/)

## Additional Features

  - Optional health checks to ensure that this service is operating normally
  - A Dockerfile to easily deploy this service to any docker-friendly cloud provider

## Running this Service with config file

```sh
pg-bridge --conf bridge.json
```

## Example Configuration

```json
{
  "postgres": {
    "url": "postgres://user:pass@localhost:5432/database"
  },
  "routes": [
    "task.create http://requestb.in/1bpu3kl1",
    "task.update arn:aws:sns:us-west-2:1234:somewhere"
  ],
  "health": {
    "port": 5000,
    "path": "/health"
  }
}
```

## Running this Service with environment variables

```sh

export PGBRIDGE='{"postgres": {"url": "postgres://user:pass@localhost:5432/database"},"routes": ["task.create http://requestb.in/1bpu3kl1", "task.update arn:aws:sns:us-west-2:1234:somewhere"], "health": {"port": 5000, "path": "/health"}}'

pg-bridge
```

##
Possible issues:

You may need to set the following AWS environment variables:
AWS_REGION=eu-west-2
AWS_ACCESS_KEY_ID=...
AWS_SECRET_ACCESS_KEY=...

## TODO

- SQS support

## License

MIT
