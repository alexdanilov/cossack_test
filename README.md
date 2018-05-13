##Generator
Run generator `cd generator; go run main.go`

Available flags:
* speed - A count of generated content per second. Default: `1`
* addr - Logger address and port. Default: `127.0.0.1:8080`

Run logger with encrypted messages


##Logger

Run logger `cd logger; go run main.go`

Available flags:
* key - 16 or 32 key to encrypt messages. Default: not used
* buffer_size - A size of buffer for each client. Default: `1`
* flow_speed - A speed limit of received messages. Default: `2`
* file_path - Path to a log file. Default: `logger.log`
* addr - listen address and port. Default: `*:8080`

Run logger with encrypted messages

`go run main.go -key=AES256Key-32Characters1234567890`