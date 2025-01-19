# extended-jobs

A simple command-line tool written in Go.

This tool displays information about running processes on Linux-based systems (similar to `ps`).

## Features

- Lists only processes started from the terminal
- Displays detailed process information, including:
    - **PID**: The Process ID.
    - **STAT**: The current process state (e.g., RUNNING, STOPPED, or DEFUNCT).
    - **COMMAND**: The original command used to start the process, including arguments.
    - **DIRECTORY**: The current working directory of the process.
- Automatically adjusts column widths for better readability
- Lightweight and fast, with minimal system resource usage
- No external dependencies; a single standalone binary

## Requirements

- Linux-based operating system

    (Tested on Ubuntu 22.04, but should work on other distributions such as Fedora or Debian)

## Installation

### Option 1: Download a pre-built binary

1. Download the appropriate binary for your system from the [releases page](https://github.com/poponta1218/extended-jobs/releases).
2. Place the binary file in a directory of your choice.
3. (Optional) Add the binary's directory to your `PATH` environment variable for easy access.

### Option 2: Install from source

1. Ensure that [Go](https://go.dev/doc/install) is installed on your system.
2. Clone the repository and build the binary:

    ```bash
    git clone https://github.com/poponta1218/extended-jobs.git
    cd extended-jobs
    go build -o ejobs
    ```

## Usage

Run the compiled binary from your terminal:

```bash
./ejobs
```

To simplify execution, consider adding the binary's directory to your PATH environment variable.

### Example output

```bash
 PID    STAT      COMMAND DIRECTORY
1234 RUNNING less foo.dat /home/user
5678 RUNNING  vim main.go /home/user/projects
9999 STOPPED    sh bar.sh /home/user
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
