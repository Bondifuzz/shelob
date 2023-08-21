# Shelob

## Shelob fuzzer for OpenAPI specification

Shelob is a powerful tool designed to test and validate APIs that are defined using the OpenAPI specification. This tool helps identify potential vulnerabilities, security flaws, and unexpected behavior within your API endpoints by generating a variety of requests based on the OpenAPI schema.

## Features

* Automated Fuzzing: Shelob Fuzzer automates the process of generating and sending requests to your API endpoints based on the OpenAPI specification. This helps discover edge cases, security vulnerabilities, and potential data inconsistencies.

* Customizable Inputs: Configure the fuzzer to generate requests with a wide range of inputs, including valid and invalid data types, various data formats, and different parameter values. This allows you to thoroughly test your API's handling of different inputs.

* Authorization Methods: Shelob Fuzzer supports various authorization methods defined in your API's OpenAPI specification. Whether your API uses API keys, OAuth tokens, or other authentication mechanisms, the Shelob will include the required headers or tokens in its requests.

* Detailed Reports: After each fuzzing session, Shelob Fuzzer generates detailed reports containing information about the requests sent, the responses received, and any anomalies or errors encountered. This aids in debugging and improving the quality of your API.

## Getting Started

### Basic usage
 
```bash
go run main.go -spec=**OPENAPI FILE** -url=**TARGET URL**
```

### Advanced options

all options:
```bash
-apikey string
    api key for auth
-detailed
    include successful test cases
-duration duration
    time duration of fuzzing (default 1h0m0s)
-output string
    output directory (default "fuzzer_output")
-password string
    password (auth)
-spec string
    openapi file specification (Required)
-token string
    token for bearer auth
-url string
    target URL (Required)
-user string
    username (auth)
```

## Demo

You can try Shelob using local demo (`demo` directory).
Just simply run `demo/createDocker.sh`.
It uses [Petstore example](https://hub.docker.com/r/swaggerapi/petstore3).
The following OpenAPI specification is located in `demo/openapi.json`

## Contributing

We welcome contributions from the community to enhance the functionality, performance, and security of the OpenAPI Fuzzer. If you're interested in contributing, please follow these steps:

* Fork the repository and create a new branch for your contribution.
* Make your changes and test thoroughly.
* Submit a pull request, including a detailed description of your changes and the problem they solve.

## License

Shelob Fuzzer is open-source software licensed under the [Apache License 2.0](https://github.com/Bondifuzz/shelob/blob/main/LICENSE).

---
Feel free to explore, use, and customize the Shelob Fuzzer to suit your API testing needs. Happy fuzzing! 🚀