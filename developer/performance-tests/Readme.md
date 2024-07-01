# Performance Tests using k6

This Note will walk you through the process of setting up the k6 
environment and running load tests using the provided Makefile.

## Available tests

- [Event HandlerTest](`event_handler_test.js)

## Setup

1. Navigate to the `performance-tests` directory.

2. Install `xk6` and build `k6` with Kubernetes support using the 
3. `setup-env` target in the Makefile.

```bash
make setup-env
```

This command installs `xk6` and builds `k6` with Kubernetes support.

## Running the Tests

1. Run the tests using the `run` target in the Makefile.

```bash
make run
```

This command runs the k6 test script. You can specify the script, 
namespace, YHS host, and YHS port by setting the `SCRIPT`, `NAMESPACE`, 
`YHS_HOST`, and `YHS_PORT` variables respectively. For example:

```bash
make run SCRIPT=script.js NAMESPACE=default YHS_HOST=localhost YHS_PORT=8989
```

## Cleanup

After running the tests, you can clean up the created pods 
using the `cleanup` target in the Makefile.

```bash
make cleanup
```

This command deletes the pods created during the test run.

## Note

The `Makefile` and the test script (`*test.js`) are configured to 
use environment variables for certain settings. You can override 
these settings by providing your own values when running the `make` commands.

For example, to run the tests with a different 
namespace and YHS URL, you can use the following command:

```bash
NAMESPACE=my-namespace YHS_URL=http://my-yhs-url make run
```

This command runs the tests with the namespace set to `my-namespace` and 
the YHS URL set to `http://my-yhs-url`.

## Important links
- [K6 documentation](https://grafana.com/docs/k6/latest/)
- [xk6 Extension](https://github.com/grafana/xk6-kubernetes?tab=readme-ov-file)
