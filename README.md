# ACS SDK for Go

Official Go SDK for [AgentCloudService](https://acs.vibecaas.app) - Cloud hosting for autonomous AI agents.

## Installation

```bash
go get acs.vibecaas.app/sdk
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    "acs.vibecaas.app/sdk"
)

func main() {
    // Register a new agent (no auth required)
    result, err := sdk.Register(sdk.RegisterOptions{Name: "my-research-agent"})
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("API Key:", result.APIKey)

    // Create client with API key
    client := sdk.NewClient(result.APIKey)

    // Deploy an OpenClaw agent
    agent, err := client.Deploy(sdk.DeployOptions{
        Name: "research-agent",
        Type: "openclaw",
        Config: map[string]interface{}{
            "model":    "claude-sonnet-4",
            "channels": []string{"telegram", "discord"},
            "skills":   []string{"web_search", "coding"},
        },
        Region: "us-east-1",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Agent deployed:", agent.URL)
}
```

## Features

- **Zero Dependencies** - Uses only standard library
- **No CAPTCHA, No Phone Verification** - Just API keys and code
- **OpenClaw Native** - Deploy pre-configured AI agents
- **Type Safe** - Full Go types for all API responses

## Authentication

```go
import "acs.vibecaas.app/sdk"

// Option 1: Pass API key directly
client := sdk.NewClient("acs_live_xxx")

// Option 2: Use environment variable
// export ACS_API_KEY=acs_live_xxx
client := sdk.NewClient("")

// Option 3: With custom options
client := sdk.NewClient("acs_live_xxx",
    sdk.WithBaseURL("https://custom-api.example.com"),
    sdk.WithTimeout(60 * time.Second),
)
```

## API Reference

### Registration (No Auth Required)

```go
// Register new agent
result, err := sdk.Register(sdk.RegisterOptions{
    Name:   "my-agent",
    Email:  "agent@example.com",  // Optional
    Wallet: "0x...",              // Optional, for USDC payments
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.AgentID)
fmt.Println(result.APIKey)

// Check service status
status, err := sdk.Status()
fmt.Println(status)
```

### Agent Management

```go
client := sdk.NewClient("acs_live_xxx")

// List all agents
agents, err := client.ListAgents()
for _, agent := range agents {
    fmt.Printf("%s: %s\n", agent.Name, agent.Status)
}

// List running agents only
running, err := client.ListAgents("running")

// Get specific agent
agent, err := client.GetAgent("agent_xxx")

// Start/stop agent
client.StartAgent("agent_xxx")
client.StopAgent("agent_xxx")

// Delete agent
client.DeleteAgent("agent_xxx")
```

### Deployment

```go
// Deploy OpenClaw agent
agent, err := client.Deploy(sdk.DeployOptions{
    Name: "my-agent",
    Type: "openclaw",
    Config: map[string]interface{}{
        "model":    "claude-sonnet-4",
        "channels": []string{"telegram", "discord", "slack"},
        "skills":   []string{"web_search", "coding", "browser"},
        "memory":   true,
    },
    Region: "us-east-1",
})

// Deploy Docker container
agent, err := client.Deploy(sdk.DeployOptions{
    Name: "custom-agent",
    Type: "docker",
    Config: map[string]interface{}{
        "image": "my-org/my-agent:latest",
        "env":   map[string]string{"API_KEY": "xxx"},
    },
})

// Deploy serverless function
agent, err := client.Deploy(sdk.DeployOptions{
    Name: "webhook-handler",
    Type: "function",
    Config: map[string]interface{}{
        "runtime": "go1.21",
        "handler": "main.Handler",
    },
})
```

### Usage & Billing

```go
// Get usage metrics
usage, err := client.GetUsage()
fmt.Printf("Requests (24h): %d\n", usage.Requests24h)
fmt.Printf("Compute used: %.2f GB\n", usage.ComputeUsed)
fmt.Printf("Current bill: $%.2f\n", usage.CurrentBill)

// Get billing info
billing, err := client.GetBilling()

// Create checkout session
checkout, err := client.Checkout("micro")
fmt.Println("Checkout URL:", checkout.CheckoutURL)

// Pay with USDC
payment, err := client.PayUSDC("19.00")
```

## Regions

| Region | Location |
|--------|----------|
| `us-east-1` | US East (Virginia) |
| `us-west-2` | US West (Oregon) |
| `eu-west-1` | EU West (Ireland) |
| `eu-central-1` | EU Central (Frankfurt) |
| `ap-southeast-1` | Asia Pacific (Singapore) |
| `ap-southeast-2` | Asia Pacific (Sydney) |

## Error Handling

```go
import "acs.vibecaas.app/sdk"

agent, err := client.GetAgent("invalid")
if err != nil {
    switch e := err.(type) {
    case *sdk.AuthenticationError:
        fmt.Println("Invalid API key")
    case *sdk.ValidationError:
        fmt.Printf("Validation failed: %s\n", e.Message)
    case *sdk.ACSError:
        fmt.Printf("API error: %s (status: %d)\n", e.Message, e.StatusCode)
    default:
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Links

- [Documentation](https://acs.vibecaas.app/docs)
- [API Reference](https://acs.vibecaas.app/api/discover)
- [OpenAPI Spec](https://acs.vibecaas.app/api/openapi.json)
- [Pricing](https://acs.vibecaas.app/pricing)
- [Status](https://acs.vibecaas.app/status)

## License

MIT License - Copyright (c) 2026 VibeCaaS / NeuralQuantum.ai LLC
