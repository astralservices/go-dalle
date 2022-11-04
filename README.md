# go-dalle

DALL-E API Wrapper for Go.

## Installation

```bash
go get github.com/astralservices/go-dalle
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/astralservices/go-dalle"
)

func main() {
    apiKey := os.Getenv("DALLE_API_KEY")
    client := dalle.NewClient(apiKey)

    data, err := client.Generate("A horse in an elevator", nil, nil, nil, nil)

    if err != nil {
        fmt.Println(err)
    }

    fmt.Println(data[0].URL)
}
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.
