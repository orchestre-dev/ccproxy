# API Architecture

CCProxy provides a unified interface for multiple AI providers through a carefully designed architecture with comprehensive security, performance monitoring, and testing capabilities.

## System Architecture

```mermaid
graph TB
    subgraph "External"
        C[Claude Code/Client]
        P1[Anthropic API]
        P2[OpenAI API]
        P3[Other Providers]
    end
    
    subgraph "CCProxy Core"
        G[API Gateway]
        A[Auth Middleware]
        R[Router]
        T[Transformer]
        E[Error Handler]
        M[Metrics]
        S[Security]
    end
    
    subgraph "Supporting Systems"
        CB[Circuit Breaker]
        RL[Rate Limiter]
        EV[Event Bus]
        ST[State Manager]
        PC[Perf Monitor]
    end
    
    C --> G
    G --> A
    A --> S
    S --> R
    R --> T
    T --> P1
    T --> P2
    T --> P3
    
    R --> CB
    A --> RL
    G --> M
    M --> PC
    R --> EV
    G --> ST
    
    style G fill:#f9f,stroke:#333,stroke-width:4px
    style S fill:#9ff,stroke:#333,stroke-width:2px
```

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

## Core Components

### Security Manager

Coordinates all security operations:

```mermaid
classDiagram
    class SecurityManager {
        +ValidateRequest(req) error
        +ValidateResponse(resp) error  
        +CheckRateLimit(key) bool
        +RecordAuditLog(event) 
    }
    
    class RequestValidator {
        +ValidateFormat(req) error
        +CheckPatterns(content) error
        +ValidateTokens(count) error
    }
    
    class ResponseSanitizer {
        +SanitizeContent(resp) error
        +MaskSensitiveData(data) string
        +RemovePatterns(content) string
    }
    
    class RateLimiter {
        +CheckLimit(key) (allowed, remaining)
        +RecordRequest(key)
        +GetStats(key) Stats
    }
    
    SecurityManager --> RequestValidator
    SecurityManager --> ResponseSanitizer
    SecurityManager --> RateLimiter
```

### Performance Monitor

Tracks system performance metrics:

```mermaid
flowchart LR
    subgraph "Metrics Collection"
        L[Latency Tracker]
        R[Resource Monitor]
        T[Throughput Counter]
    end
    
    subgraph "Analysis"
        P[Percentile Calc]
        A[Aggregator]
        H[Histogram]
    end
    
    subgraph "Output"
        M[Metrics Endpoint]
        G[Grafana Export]
        AL[Alerts]
    end
    
    L --> P
    R --> A
    T --> H
    
    P --> M
    A --> M
    H --> M
    
    M --> G
    M --> AL
```

### State Management

Manages service and component states:

```mermaid
stateDiagram-v2
    [*] --> Initializing
    Initializing --> Starting
    Starting --> Ready
    Ready --> Degraded
    Ready --> Stopping
    Degraded --> Ready
    Degraded --> Error
    Degraded --> Stopping
    Error --> Stopping
    Stopping --> Stopped
    Stopped --> [*]
    
    Ready --> Ready : Healthy
    Degraded --> Degraded : Partial Failure
    Error --> Error : Complete Failure
```

### Event System

Event-driven architecture for extensibility:

```mermaid
flowchart TB
    subgraph "Event Producers"
        API[API Gateway]
        AUTH[Auth System]  
        PROV[Providers]
    end
    
    subgraph "Event Bus"
        EB{Event Bus}
        Q1[High Priority]
        Q2[Normal Priority]
        Q3[Low Priority]
    end
    
    subgraph "Event Consumers"
        LOG[Logger]
        MET[Metrics]
        AUD[Audit]
        ALT[Alerts]
    end
    
    API --> EB
    AUTH --> EB
    PROV --> EB
    
    EB --> Q1
    EB --> Q2
    EB --> Q3
    
    Q1 --> ALT
    Q2 --> LOG
    Q2 --> MET
    Q3 --> AUD
```

## Testing Architecture

### Test Framework Components

```mermaid
classDiagram
    class TestFramework {
        +NewTestContext() Context
        +StartServer() Server
        +CreateMockProvider() Mock
    }
    
    class Fixtures {
        +GetRequest(name) Request
        +GetResponse(name) Response
        +GenerateData(size) Data
    }
    
    class Assertions {
        +AssertJSONEqual()
        +AssertEventually()
        +AssertMetrics()
    }
    
    class MockServer {
        +AddRoute(path, response)
        +SetDelay(duration)
        +GetRequests() []Request
    }
    
    TestFramework --> Fixtures
    TestFramework --> Assertions
    TestFramework --> MockServer
```

## Build and Deployment

### CI/CD Pipeline

```mermaid
flowchart LR
    subgraph "Development"
        D[Developer] --> G[Git Push]
    end
    
    subgraph "CI/CD"
        G --> T[Tests]
        T --> L[Lint]
        L --> B[Build]
        B --> S[Security Scan]
        S --> P[Package]
    end
    
    subgraph "Deployment"
        P --> D1[Docker Registry]
        P --> R[GitHub Release]
        P --> H[Helm Chart]
    end
    
    style T fill:#9f9
    style S fill:#9ff
```