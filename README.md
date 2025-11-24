# Task Manager (task manager) - is a digital analog of the usual notepad with a list of tasks, but with "super abilities".

## Business value for different users:
1. 👤 For individual user:
2. ✅ Do not forget about the important - deadlines, meetings, personal goals
3. 📊 See progress - what tasks have been completed, what work
4. 🎯 Prioritize - what to do first, what can wait
5. ⏰ Time planning - estimation of deadlines and actual execution time
6. 👥 For team/company:
7. 🔄 Process transparency - who does what, at which stage of the task
8. 📈 Analytics productivity - how many tasks are performed, where "bottle necks"
9. 🗓️ Coordination of work - see the dependence of tasks on each other
10. 📋 Documentation of work - execution history, attached files

### Real business case:
```text
Customer support
Task: "Process client request #12345"
Priority: high
Status: in_progress completed
Deadline: 2024-01-15 18:00
```

### Implemented functionalities
1. Graceful shutdown
2. JWT validation
3. RabbitMQ(Email sending)
4. Docker containers(several services with different business logic)
5. Golang Migrate schemas
6. Logger `Zap`
7. GRPC microservices
8. Prometheus + Grafana 
   
### How to start?
- Create .env file with params:
```text
SERVER_HOST="api-gateway" (dont change)
SERVER_PORT=8080

APP_ENV=development

JWT_SECRET=your-super-secure-jwt-secret-key-here
JWT_ACCESS_TTL=60
JWT_REFRESH_TTL=720

DB_HOST="postgres" (dont change)
DB_PORT=5432
DB_USER="postgres"
DB_PASSWORD="postgres"
DB_NAME="postgres"

Task_GRPC_HOST="taskservice"(dont change)
Task_GRPC_PORT=50052

Auth_GRPC_HOST="authservice"(dont change)
Auth_GRPC_PORT=50051

SMTP_FROM_EMAIL="email@gmail.com"
SMTP_FROM_PASS="..."
```

- Write in terminal a command:
```text
make run
```  
