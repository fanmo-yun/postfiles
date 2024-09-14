# PostFiles

PostFiles is a file transfer tool written in Go that supports both client and server modes. The server can provide files for the client to download, the client can download files from the server.

## Usage
### Basic Usage
-- server-side mode
```bash
postfiles server -i 127.0.0.1 -p 9090 -f file1.txt -f file2.txt
```
-- client-side mode
```bash
postfiles client -i 127.0.0.1 -p 9090 -s /path/to/save
```
