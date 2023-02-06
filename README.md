### Golang OpenTelemetry example with Zipkins and Jaeger

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