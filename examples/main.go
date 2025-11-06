package main

import (
	"fmt"
	"log"
	"time"

	async "github.com/geoffreyhinton/redis_asyncqueue_task"
)

func main() {
	fmt.Println("=== Simple Redis Async Queue Test ===")

	// Step 1: Configure Redis connection
	redisOpt := &async.RedisOpt{
		Addr:     "localhost:6379", // Make sure Redis is running on this address
		Password: "",               // No password for local development
	}

	// Step 2: Create a client to enqueue tasks
	client := async.NewClient(redisOpt)
	fmt.Println("✓ Client created")

	// Step 3: Create workers to process tasks
	launcher := async.NewLauncher(2, redisOpt) // 2 concurrent workers
	fmt.Println("✓ Worker launcher created")

	// Step 4: Define how to handle different types of tasks
	taskHandler := func(task *async.Task) error {
		fmt.Printf("🔄 Processing task: %s\n", task.Type)

		switch task.Type {
		case "send_email":
			email := task.Payload["email"].(string)
			subject := task.Payload["subject"].(string)
			fmt.Printf("  📧 Sending email to: %s, Subject: %s\n", email, subject)
			time.Sleep(1 * time.Second) // Simulate sending email

		case "process_image":
			filename := task.Payload["filename"].(string)
			fmt.Printf("  🖼️  Processing image: %s\n", filename)
			time.Sleep(2 * time.Second) // Simulate image processing

		case "generate_report":
			reportType := task.Payload["type"].(string)
			fmt.Printf("  📊 Generating %s report\n", reportType)
			time.Sleep(3 * time.Second) // Simulate report generation

		default:
			fmt.Printf("  ❓ Unknown task type: %s\n", task.Type)
		}

		fmt.Printf("✅ Completed task: %s\n", task.Type)
		return nil
	}

	// Step 5: Start the workers
	go func() {
		fmt.Println("🚀 Starting workers...")
		launcher.Start(taskHandler)
	}()

	// Give workers a moment to start
	time.Sleep(1 * time.Second)

	// Step 6: Enqueue some immediate tasks
	fmt.Println("\n=== Enqueuing Immediate Tasks ===")

	tasks := []*async.Task{
		{
			Type: "send_email",
			Payload: map[string]interface{}{
				"email":   "user1@example.com",
				"subject": "Welcome to our service!",
			},
		},
		{
			Type: "process_image",
			Payload: map[string]interface{}{
				"filename": "profile_photo.jpg",
			},
		},
		{
			Type: "send_email",
			Payload: map[string]interface{}{
				"email":   "user2@example.com",
				"subject": "Your order is confirmed",
			},
		},
	}

	// Enqueue immediate tasks
	for i, task := range tasks {
		err := client.Process(task, time.Now())
		if err != nil {
			log.Printf("❌ Failed to enqueue task %d: %v", i+1, err)
		} else {
			fmt.Printf("✓ Enqueued task %d: %s\n", i+1, task.Type)
		}
	}

	// Step 7: Enqueue a scheduled task
	fmt.Println("\n=== Enqueuing Scheduled Task ===")

	scheduledTask := &async.Task{
		Type: "generate_report",
		Payload: map[string]interface{}{
			"type": "daily_sales",
		},
	}

	// Schedule this task to run 5 seconds from now
	executeAt := time.Now().Add(5 * time.Second)
	err := client.Process(scheduledTask, executeAt)
	if err != nil {
		log.Printf("❌ Failed to schedule task: %v", err)
	} else {
		fmt.Printf("✓ Scheduled report task for %s\n", executeAt.Format("15:04:05"))
	}

	// Step 8: Wait and observe
	fmt.Println("\n=== Watching Tasks Execute ===")
	fmt.Println("⏳ Waiting for tasks to complete...")

	// Wait for all tasks to complete
	time.Sleep(10 * time.Second)

	// Add one more task to show the system is still working
	fmt.Println("\n=== Adding Final Task ===")
	finalTask := &async.Task{
		Type: "send_email",
		Payload: map[string]interface{}{
			"email":   "admin@example.com",
			"subject": "Test completed successfully!",
		},
	}

	err = client.Process(finalTask, time.Now())
	if err != nil {
		log.Printf("❌ Failed to enqueue final task: %v", err)
	} else {
		fmt.Println("✓ Enqueued final task")
	}

	// Wait a bit more to see the final task execute
	time.Sleep(3 * time.Second)

	fmt.Println("\n=== Test Complete! ===")
	fmt.Println("🎉 All tasks have been processed successfully!")
	fmt.Println("Note: Workers are still running. Press Ctrl+C to stop.")

	// Keep the program running
	select {}
}
