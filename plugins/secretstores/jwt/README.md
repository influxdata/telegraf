# JWT Token Generator

The JWT Token Generator is a plugin for [Telegraf](https://www.influxdata.com/time-series-platform/telegraf/), which provides a way to generate and manage JSON Web Tokens (JWTs) for multiple users.

## Configuration

The JWT Token Generator can be configured using the following properties:

- `id`: A unique identifier for the generator.
- `dynamic`: A boolean that defines whether the JWT tokens should be dynamically updated.
- `users`: A list of user names for which the JWT tokens are to be generated.
- `urls`: A list of URLs corresponding to each user, used to generate the JWT tokens.
- `passwords`: A list of passwords corresponding to each user.

Please refer to the `sample.conf` file for an example of how to configure the JWT Token Generator.

## Usage

After setting up the configuration, you can use the plugin's methods to manage JWT tokens:

- `Get(key string)`: Returns the JWT token for the specified user.
- `Set(key string, value string)`: Sets a new JWT token for a specified user.
- `List()`: Returns a list of all user names for which JWT tokens are generated.

Refer to the source code for more details on these methods and their usage.
