# Micro-batching Mock Service

## Overview

This service demonstrates a micro-batching library, which processes tasks in batches to improve throughput. The library allows for submitting jobs, processing them in batches, and retrieving results. It is configurable for batch size and interval, and provides a shutdown method to ensure all jobs are processed before the service exits.

## Running the Service within a Docker Container

1. **Run Docker Container:**

    ```sh
    docker run --rm -it -v ./:/app -p 8080:8080 golang:1.18-alpine /bin/sh
    ```

2. **Navigate to the Application Directory:**

    ```sh
    cd /app
    ```

3. **Install GCC and Other Necessary Build Tools:**

    ```sh
    apk update && apk add --no-cache gcc musl-dev
    ```

4. **Build the Service:**

    ```sh
    go build -o microbatch-service main.go
    ```

5. **Run the Service:**

    ```sh
    ./microbatch-service
    ```

6. **Send Requests:**

    - Submit a job:

        ```sh
        curl -X POST -H "Content-Type: application/json" -d '{"data":"sample data","value":123}' http://localhost:8080/submit
        ```

    - Shutdown the service:

        ```sh
        curl -X POST http://localhost:8080/shutdown
        ```

    - Get the result of a job:

        ```sh
        curl -X POST http://localhost:8080/result?id=3
        ```

7. **Run Unit Tests:**

    ```sh
    go test ./microbatch
    ```

## Running the Docker Image

1. **Build the Docker Image:**

    ```sh
    docker build -t microbatch-service .
    ```

2. **Run the Docker Container:**

    ```sh
    docker run -p 8080:8080 -e BATCH_SIZE=5 -e BATCH_INTERVAL=10 microbatch-service
    ```

## Configuration

- **BATCH_SIZE:** The number of jobs to accumulate before processing a batch.
- **BATCH_INTERVAL:** The maximum time interval (in seconds) to wait before processing a batch.

## Example Usage

Submit a job to the service:

```sh
curl -X POST -H "Content-Type: application/json" -d '{"data":"example data","value":456}' http://localhost:8080/submit
```

Shutdown the service:

```sh
curl -X POST http://localhost:8080/shutdown
```

Retrieve the result of a job:

```sh
curl -X GET "http://localhost:8080/result?id=1"
```

## Development and Testing

Ensure all dependencies are installed and the development environment is set up. Run the unit tests to verify the functionality of the micro-batching library.

```sh
go test ./microbatch
```
