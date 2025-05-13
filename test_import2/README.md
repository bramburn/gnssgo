# GNSSGO Library Import Example

This is a simple example of how to import and use the GNSSGO library in your Go application.

## Setup

1. Create a new Go module:
   ```bash
   mkdir myapp
   cd myapp
   go mod init myapp
   ```

2. Create a go.work file to use the GNSSGO library:
   ```bash
   # Create a go.work file
   cat > go.work << EOF
   go 1.21

   use (
       .
       /path/to/gnssgo/src
   )
   EOF
   ```

3. Create a main.go file:
   ```go
   package main

   import (
       "fmt"

       "github.com/bramburn/gnssgo"
   )

   func main() {
       fmt.Printf("GNSSGO Version: %s\n", gnssgo.VER_GNSSGO)
       
       // Example: Open a serial port stream
       // var stream gnssgo.Stream
       // stream.OpenStream(gnssgo.STR_SERIAL, gnssgo.STR_MODE_RW, "COM1:115200:8:N:1")
       
       // Read data
       // buff := make([]byte, 1024)
       // n := stream.StreamRead(buff, 1024)
       
       // Close when done
       // stream.StreamClose()
   }
   ```

4. Run the application:
   ```bash
   go run main.go
   ```

## Using GNSSGO in Your Project

To use GNSSGO in your project, you can either:

1. Use a go.work file as shown above (good for development)
2. Clone the GNSSGO repository and use it as a local module
3. Import directly from GitHub once it's published

### Example: Serial Port Usage

```go
import (
    "github.com/bramburn/gnssgo"
)

func main() {
    // Open a serial port stream
    // Format: port[:brate[:bsize[:parity[:stopb[:fctr[#port]]]]]]
    // Example: COM1:115200:8:N:1:off
    var stream gnssgo.Stream
    stream.OpenStream(gnssgo.STR_SERIAL, gnssgo.STR_MODE_RW, "COM1:115200:8:N:1")
    
    // Read data
    buff := make([]byte, 1024)
    n := stream.StreamRead(buff, 1024)
    
    // Close when done
    stream.StreamClose()
}
```
