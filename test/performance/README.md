# Performance Tests using k6
K6 is an open-source load testing tool for testing the performance of your backend infrastructure. `xk6` is a tool that makes easy to 
interact with kubernetes while writing load tests using `k6.` In this setup we use `k6` with `xk6` extension to test the performance of 
different components of the Yunikorn History Server.

## Prerequisites
- [K6](https://grafana.com/docs/k6/latest/)
- [xk6 Extension](https://github.com/grafana/xk6-kubernetes?tab=readme-ov-file)

## Available tests

- [Event HandlerTest](`event_handler_test.js`) : When an event is generated in the cluster `event_handler` listen to the event and store 
it in the database. This test is used to test the
performance of the event handler when multiple events are generated in the cluster.

## Running the Tests

1. Run the tests using the `performance-tests` target in the Makefile.

```bash
make performance-tests
```
