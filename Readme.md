# PostgreSQL → SNS Bridge

Simple service connecting [PostgreSQL notifications](https://www.postgresql.org/docs/9.0/static/sql-notify.html) to [Amazon SNS](https://aws.amazon.com/sns/).

## Features

  - Flexible Routing: Supports one-to-one and one-to-many routing
  - Heroku-friendly: Configuration is done entirely through environment variables
  - Health checks: Optional HTTP endpoint to ensure that this service is operating normally

## Environment Variables

- `POSTGRES_URL`: **(required)** URL string to connect to Postgres.
- `BRIDGE_ROUTES`: **(required)** comma-delimited list of routes. See [Routes Format](#routes-format) for how to configure this variable.
- `AWS_ACCESS_KEY_ID`: **(optional)** AWS account ID. This can come from `~/.aws` too.
- `AWS_SECRET_ACCESS_KEY`: **(optional)** AWS access key. This can come from `~/.aws` too.
- `AWS_REGION`: **(optional)** AWS region. This can come from `~/.aws` too.
- `PORT`: **(optional)** Port to serve health information on.
- `HEALTH_PATH`: **(optional)** Path your health information is on. Defaults to `/health`.

> I recommend using [direnv](http://direnv.net) to manage your environment variables

## Routes Format

Here's the format:

    BRIDGE_ROUTES="CHANNEL_A,SNS_TOPIC_ARN_A;CHANNEL_B,SNS_TOPIC_ARN_B;CHANNEL_C,SNS_TOPIC_ARN_B"

Here's an example:

    BRIDGE_ROUTES="task.create,arn:aws:sns:us-west-2:123:taskcreate;task.update,arn:aws:sns:us-west-2:456:taskupdate"


## License

MIT
