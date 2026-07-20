<div align="center">

# Callecto API

> Automated AI voicebot campaign calling and debt collection orchestration engine.

![img](https://cdn.discordapp.com/attachments/372372440334073859/1216723412244893738/go_1.gif?ex=66016cfb&is=65eef7fb&hm=76c9538616341d981031c54ade56cbeee7eadc96c71d64bbaa6f85cc60436b40&)

</div>

---

Callecto is a high-performance campaign dialer service designed to schedule, queue, trigger, and log automated collection calls using AI-powered voicebots.

The application is structured as a Go Fiber web service integrated with MongoDB and external AI call routing APIs (e.g. Botnoi Voicebot).

---

## 🛠️ Tech Stack

- **Backend Framework**: [Go Fiber v2](https://github.com/gofiber/fiber) (Go 1.22+)
- **Database**: [MongoDB](https://www.mongodb.com/) (using MongoDB Go Driver)
- **Containerization**: Docker / Podman
- **Hot-Reload Tooling**: Air
- **API Spec**: OpenAPI 3.1.0

---

## 🚀 Getting Started

### 1. Prerequisites

- **Go Compiler**: Go 1.21 or higher installed on your machine.
- **Database**: Access to a MongoDB database (local or Atlas cluster).
- **Node.js** (Optional, to regenerate documentation).

---

### 2. Installation

Clone the repository and run `go mod tidy` to download dependencies:

```bash
git clone https://github.com/natchanonsarasang/callecto-api.git
cd callecto-api
go mod tidy
```

---

### 3. Environment Configuration

Create a `.env` file in the root directory (based on the sample [.env](file:///home/cellul4r/Documents/botnoi/callecto-api/.env)):

```ini
PORT=8080

MONGODB_URI=mongodb+srv://<username>:<password>@cluster.mongodb.net
MONGODB_NAME=callecto_db

JWT_SECRET_KEY=YourJWTSecret
JWT_REFESH_SECRET_KEY=YourJWTRefreshSecret

JWK_SET_URL=https://<your-supabase-project>.supabase.co/auth/v1/.well-known/jwks.json

OUTBOUND_URL=http://<outbound-dialer-ip>:5667/outbound
OUTBOUND_ACCESS_TOKEN=<your-token>
```

| Env Variable | Type | Required | Description | Default |
|---|---|---|---|---|
| `PORT` | int | No | The port the Go application binds to. | `8080` |
| `MONGODB_URI` | string | Yes | Connection string for the MongoDB instance. | - |
| `JWT_SECRET_KEY` | string | Yes | Secret key used to decode authentication payloads. | - |
| `JWK_SET_URL` | string | Yes | Remote keys server endpoint to verify JWT signatures. | - |
| `OUTBOUND_URL` | string | Yes | API endpoint of the outbound dialer voicebot engine. | - |

---

### 4. Running the Application

#### A. Standard Run
```bash
go run main.go
```

#### B. Hot-Reload Mode (Development)
Install [Air](https://github.com/cosmtrek/air) for live hot-reloading:
```bash
go install github.com/cosmtrek/air@latest
air
```

#### C. Container Mode (Podman / Docker)
```bash
# Build the container image
podman build -t callecto-api .

# Run the container
podman run --rm -it -p 8080:8080 callecto-api
```

---

## 📂 Architecture & Directory Structure

```
.
├── cmd/                # Seeding scripts and commands
├── configuration/      # Fiber engine and client initializations
├── domain/             
│   ├── datasources/    # MongoDB client setup
│   ├── entities/       # Core model struct schemas
│   └── repositories/   # DB layer implementing queries
├── src/                
│   ├── gateways/       # HTTP route endpoints handlers
│   ├── middlewares/    # Logger & JWT validation routines
│   └── services/       # Dialer queues and campaign control business logic
└── docs/               # System manual and API reference guides
```

- System manuals and architecture decisions are detailed in [architecture_manual.md](file:///home/cellul4r/Documents/botnoi/callecto-api/docs/architecture_manual.md).
- Integration onboarding guide can be found at [developer_guide.md](file:///home/cellul4r/Documents/botnoi/callecto-api/docs/developer_guide.md).

---

## 📃 API Documentation Reference

Callecto API endpoints are fully documented and interactive:
- **Interactive Swagger Docs**: [interactive_docs.html](file:///home/cellul4r/Documents/botnoi/callecto-api/docs/interactive_docs.html) (Double click/open in web browser).
- **OpenAPI 3.1 Spec**: [openapi.yaml](file:///home/cellul4r/Documents/botnoi/callecto-api/openapi.yaml).
- **Integration Code Snippets**: Refer to the JS, Python, Go, and cURL snippets in the [docs/examples/](file:///home/cellul4r/Documents/botnoi/callecto-api/docs/examples/) directory.

---

## 📄 License
MIT
