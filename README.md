# Go Raydium API SDK

## Usage
```go
import "github.com/ipanardian/go-raydium-api"

var res raydium.RaydiumData
cl := raydium.NewRaydiumAPI(RaydiumURI)
headers := make(map[string]string)
err = cl.SwapQuote(&res, headers, raydium.RaydiumQuoteRequest{
    InputMint:      baseAddress,
    OutputMint:     quoteAddress,
    Amount:         1000000,
    SlippageBps:    50,
    TxVersion:      "v0",
})
```