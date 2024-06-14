<p align="center">
  <img src="assets/images/macgover.png" />
</p>

<center>
<h1>
    <b>Macgover</b>
</h1>
</center>

## Informations: 
- Location: https://gallery.ecr.aws/n3v7f6o4/macgover
- Version: public.ecr.aws/n3v7f6o4/macgover:22.09

## Usage

- Command:
    - `./macgover --mode server [--port 3000]`
    - `./macgover --mode batch --job metrics [--argument='{"job": "macgover_batch_job", "label": "macgover_batch_label", "value": 1}']`

- Parameters:
    - `--mode server` : to start a webserver (by default)
        - `[-- port]` : to specify a port number (by default 3000)
    - `--mode batch` : to start a job 
        - `--job metrics` : to launch the job "metrics"
        - `--argument <args>` : arguments of the job


## Webserver
- `GET    /`

### Paths

- `/` : redirect to the homepage
- `/whoami` : display a web page
    - `[?wait=5s]` : display a web page after 5 seconds
- `/ping` : display a lite web page
    - `[?format=json]` : display the result in JSON format
- `/echo` : display the request 
- `/healthcheck`: heath check
    - `[?code=404]` : returns a response with the status code defined (ex 404)
- `/ldap` try the connection and bind to a ldap 
    - environment variable : 
        - `LDAP_URL="ldap://xxxxxxx"`
        - `LDAP_BIND_DN="ou=programs,o=xxx"`
- `/db/:engine` connect to a mysql/postgresql database
    - environment variables :
        - `DB_USER` : database user
        - `DB_PASSWORD` : database password
        - `DB_HOST` : url format (ex localhost)
        - `DB_PORT` : port format (ex 3306)
        - `DB_NAME` : database name
        - `DB_TIMEOUT` : timeout in second of the connection (format integer, default=5)
    - `[/count/:table]` : display the number of row of one table
- `/metrics` 
    - get metrics in prometheus format
    - post metrics (format pushmetrics)
- `/url` : to check the connection with a website
    - `[?test=https://my.url.com]` : for testing a custom website
- `/network` : to check the connection on @ip port


## Build
`docker build --build-arg "MACGOVER_COMMIT=$(git show -s --format=%H)" -t macgover:beta .`

## Examples with docker

```console
#--- mode batch
$ docker run macgover:beta --mode batch --job metrics --argument='{"job": "macgover_batch_job", "label": "macgover_batch_label", "value": 1}'

#--- mode server
$ docker run -d -p 3000:3000 macgover:beta --mode server

$ curl "http://localhost:3000/v1/healthcheck"
Healthcheck return code: 200

$ curl "http://localhost:3000/v1/ping?format=json"
{"hostname":"9e51408e7842","date":1636991111,"message":"pong"}
```
