# pprof Analysis Script
# Run while server is under load for best results

# 1. Capture heap profile (while server is running)
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/heap

# 2. Capture CPU profile (30 seconds under load)
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/profile?seconds=30

# 3. Capture goroutine (current state)
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/goroutine?debug=1

# Alternative: Save to file then analyze
curl http://localhost:6060/debug/pprof/heap -o heap.prof
go tool pprof -text heap.prof | head -50

# For more detailed analysis:
go tool pprof -alloc_space heap.prof  # Show allocation locations
go tool pprof -inuse_space heap.prof  # Show current memory usage