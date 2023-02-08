### Golang OpenTelemetry example with Datadog, Zipkins and Jaeger

Start Jaeger, Zipkin and OTEL Collector
```shell
$ docker-compose up
```

Start `serviceone`
```shell
$ cd serviceone/
$ go run main.go
```

Start `servicetwo`
```shell
$ cd servicetwo/
$ go run main.go
```

Run `client`
```shell
$ cd client/
$ go run main.go Samsung
```

### TODO
- Add Metrics

#### Datadog span

[<img src="./docs/datadog_span.PNG" width="600">](https://github.com/musobarlab)
<br/><br/>

#### Jaeger span

[<img src="./docs/jaeger_span.PNG" width="600">](https://github.com/musobarlab)
<br/><br/>

#### Zipkin span

[<img src="./docs/zipkin_span.PNG" width="600">](https://github.com/musobarlab)
<br/><br/>

#### Datadog metric

[<img src="./docs/datadog_metric.PNG" width="600">](https://github.com/musobarlab)
<br/><br/>

#### Grafana metric

[<img src="./docs/grafana_metric.PNG" width="600">](https://github.com/musobarlab)
<br/><br/>

### Errors
- Docker Compose error: `prometheus docker /prometheus/queries.active: permission denied`

Set permission to `Prometheus volume and Grafana volume`
```shell
$ id -u
$ 1000
$ sudo chown -R 1000:1000 volumes/prometheus/data/
$ sudo chown -R 1000:1000 volumes/grafana/data/
```