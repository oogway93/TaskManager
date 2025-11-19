run:
	sudo docker compose up --build

stop:
	sudo docker compose down

gen:
	protoc --go_out=. --go-grpc_out=. proto/*.proto

resetdb:
	migrate -path migrations/ -database "postgresql://postgres:postgres@localhost:5432/taskmanager?sslmode=disable" -verbose drop && migrate -path migrations/ -database "postgresql://postgres:postgres@localhost:5432/taskmanager?sslmode=disable" -verbose up 
