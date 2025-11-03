run:
	sudo docker compose up --build

stop:
	sudo docker compose down

gen:
	protoc --go_out=. --go-grpc_out=. proto/*.proto