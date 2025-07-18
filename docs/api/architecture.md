# API Architecture

CCProxy provides a unified interface for multiple AI providers through a carefully designed architecture.

## Request Flow

The following diagram shows how requests flow through CCProxy:

```mermaid
flowchart LR
    A[Client] -->|API Request| B[CCProxy API Gateway]
    B --> C{Authentication}
    C -->|Valid| D[Request Parser]
    C -->|Invalid| E[401 Error]
    D --> F{Route by Provider}
    F -->|OpenAI| G[OpenAI Handler]
    F -->|Groq| H[Groq Handler]
    F -->|Gemini| I[Gemini Handler]
    F -->|Other| J[Provider Handler]
    G --> K[Transform Request]
    H --> K
    I --> K
    J --> K
    K --> L[Call Provider API]
    L --> M[Transform Response]
    M --> N[Stream/Return to Client]
```

## Provider Architecture

Each provider follows a common interface pattern:

```mermaid
classDiagram
    class Provider {
        <<interface>>
        +Name() string
        +HandleChatCompletion(req) Response
        +HandleStreamingResponse(stream) error
        +TransformRequest(req) ProviderRequest
        +TransformResponse(resp) StandardResponse
    }
    
    class BaseProvider {
        #apiKey string
        #httpClient HTTPClient
        #logger Logger
        +ValidateConfig() error
        +MakeRequest(req) Response
    }
    
    class OpenAIProvider {
        +HandleChatCompletion(req) Response
        +TransformRequest(req) OpenAIRequest
    }
    
    class GroqProvider {
        +HandleChatCompletion(req) Response
        +TransformRequest(req) GroqRequest
        +SupportsFunctions() bool
    }
    
    Provider <|-- BaseProvider
    BaseProvider <|-- OpenAIProvider
    BaseProvider <|-- GroqProvider
```

## Error Handling Flow

```mermaid
sequenceDiagram
    participant Client
    participant CCProxy
    participant Provider
    
    Client->>CCProxy: Request
    CCProxy->>CCProxy: Validate
    alt Validation Fails
        CCProxy-->>Client: 400 Bad Request
    else Validation Passes
        CCProxy->>Provider: Forward Request
        alt Provider Error
            Provider-->>CCProxy: Error Response
            CCProxy->>CCProxy: Log Error
            CCProxy-->>Client: 502 Bad Gateway
        else Success
            Provider-->>CCProxy: Success Response
            CCProxy-->>Client: 200 OK + Data
        end
    end
```

## Streaming Response Architecture

```mermaid
flowchart TD
    A[Provider Stream] --> B[Response Reader]
    B --> C{Parse Chunk}
    C -->|Valid JSON| D[Transform Data]
    C -->|Invalid| E[Skip Chunk]
    D --> F[Write to Client Stream]
    F --> G{More Data?}
    G -->|Yes| B
    G -->|No| H[Close Stream]
    E --> G
```

## Load Balancing Strategy

```mermaid
graph TB
    subgraph "Load Balancer"
        LB[Request Router]
    end
    
    subgraph "Provider Pool"
        P1[Provider 1<br/>Health: ✓]
        P2[Provider 2<br/>Health: ✓]
        P3[Provider 3<br/>Health: ✗]
    end
    
    subgraph "Health Monitor"
        HM[Health Checker]
    end
    
    Client --> LB
    LB --> P1
    LB --> P2
    LB -.->|Skip| P3
    HM --> P1
    HM --> P2
    HM --> P3
    
    style P3 fill:#f99,stroke:#333,stroke-width:2px
    style P1 fill:#9f9,stroke:#333,stroke-width:2px
    style P2 fill:#9f9,stroke:#333,stroke-width:2px
```