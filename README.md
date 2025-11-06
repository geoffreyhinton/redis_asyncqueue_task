# Redis Async Queue Task

A simple async task queue implementation using Redis and Go.

## Features

- ✅ Immediate task processing
- ⏰ Scheduled task execution
- 🔄 Worker pool with configurable concurrency
- 📊 Support for different task types
- 🚀 Built on Redis for reliability and scalability

## Prerequisites

1. **Redis Server**: Make sure Redis is running locally
   ```bash
   # Install Redis (macOS)
   brew install redis
   
   # Start Redis
   brew services start redis
   
   # Or run Redis manually
   redis-server
   ```

2. **Go Modules**: This project uses Go modules
   ```bash
   go mod tidy
   ```

## Usage

### Basic Example

```bash
cd examples
go run main.go
```

This will:
1. Create a Redis client for enqueuing tasks
2. Start a worker pool with 2 concurrent workers
3. Enqueue several immediate tasks (email sending, image processing)
4. Schedule a task to run 5 seconds later
5. Show real-time task processing

### Task Types

The example demonstrates different task types:

- **send_email**: Simulates sending an email
- **process_image**: Simulates image processing
- **generate_report**: Simulates report generation

### Expected Output

```
=== Simple Redis Async Queue Test ===
✓ Client created
✓ Worker launcher created
🚀 Starting workers...

=== Enqueuing Immediate Tasks ===
✓ Enqueued task 1: send_email
✓ Enqueued task 2: process_image
✓ Enqueued task 3: send_email
🔄 Processing task: send_email
  📧 Sending email to: user1@example.com, Subject: Welcome to our service!
🔄 Processing task: process_image
  🖼️  Processing image: profile_photo.jpg
✅ Completed task: send_email
🔄 Processing task: send_email
  📧 Sending email to: user2@example.com, Subject: Your order is confirmed
...
```

## Code Structure

- `async.go` - Main library with Client and Workers
- `examples/main.go` - Test and demonstration code
- `go.mod` - Go module dependencies

## Key Components

### Client
```go
client := async.NewClient(&async.RedisOpt{
    Addr: "localhost:6379",
    Password: "",
})
```

### Workers
```go
launcher := async.NewWorkers(poolSize, redisOpt)
launcher.Start(taskHandler)
```

### Task Definition
```go
task := &async.Task{
    Type: "send_email",
    Payload: map[string]interface{}{
        "email": "user@example.com",
        "subject": "Hello World",
    },
}
```

### Processing Tasks
```go
// Immediate processing
client.Process(task, time.Now())

// Scheduled processing (5 seconds later)
client.Process(task, time.Now().Add(5*time.Second))
```

## Troubleshooting

1. **Redis Connection Error**: Make sure Redis is running on `localhost:6379`
2. **Import Errors**: Run `go mod tidy` to download dependencies
3. **No Tasks Processing**: Check that workers are started with `launcher.Start(handler)`

## Stopping the Program

Press `Ctrl+C` to stop the workers and exit the program.