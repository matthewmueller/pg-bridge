# Postgres → SNS Bridge

Simple service connecting [PostgreSQL notifications](https://www.postgresql.org/docs/9.0/static/sql-notify.html) to [Amazon SNS](https://aws.amazon.com/sns/).

## Features

  - Flexible Routing: Supports one-to-one and one-to-many routing
  - Heroku-friendly: Configuration is done entirely through environment variables
  - Health checks: Optional HTTP endpoing to ensure that this service is operating normally

## Environment Variables

- `AWS_ACCESS_KEY_ID`: (optional) AWS account id.
- `AWS_SECRET_ACCESS_KEY`: (optional) AWS access key.
- `POSTGRES_URL`: (required) URL string to connect to Postgres.
- `HEALTH_PORT`: (optional) 5000
- `BRIDGE_ROUTES`: (required) comma-delimited list of routes. see below

> I recommend using [direnv](http://direnv.net) to manage your environment variables

## Routes Format

Here's the format:

    BRIDGE_ROUTES="CHANNEL_A,SNS_TOPIC_ARN_1;CHANNEL_B,SNS_TOPIC_ARN_1"

Here's an example:

    BRIDGE_ROUTES="task.create,arn:aws:sns:us-west-2:123:taskcreate;task.update,arn:aws:sns:us-west-2:456:taskupdate"


## License

MIT
