.PHONY: all docker run-event-producer run-event-consumer run-event-consumer-2 run-event-consumer-3 run-command-producer run-command-consumer run-command-consumer-2 run-command-consumer-3

all: docker
	@sleep 5
	@echo "Running event targets in parallel..."
	@make -j 7 \
	run-event-consumer \
	run-event-consumer-2 \
	run-event-consumer-3 \
	run-command-producer \
	run-command-consumer \
	run-command-consumer-2 \
	run-command-consumer-3

docker:
	@echo "Run docker-compose..."
	@docker-compose -f docker-compose.yml up -d --remove-orphans

run-event-producer:
	@echo "Running event producer..."
	@APP_NAME="event_producer_1" go run ./event_producer/main.go
	@echo "Event producer finished."

run-event-consumer:
	@echo "Running event consumer..."
	@APP_NAME="event_consumer_1" go run ./event_consumer/main.go
	@echo "Event consumer 1 finished."

run-event-consumer-2:
	@echo "Running event consumer 2..."
	@APP_NAME="event_consumer_2" go run ./event_consumer/main.go
	@echo "Event consumer 2 finished."

run-event-consumer-3:
	@echo "Running event consumer 3..."
	@APP_NAME="event_consumer_3" go run ./event_consumer/main.go
	@echo "Event consumer 3 finished."

run-command-producer:
	@echo "Running command producer..."
	@APP_NAME="command_producer_1" go run ./command_producer/main.go
	@echo "Command producer finished."

run-command-consumer:
	@echo "Running command consumer..."
	@APP_NAME="command_consumer_1" go run ./command_consumer/main.go
	@echo "Command consumer 1 finished."
	
run-command-consumer-2:
	@echo "Running command consumer 2..."
	@APP_NAME="command_consumer_2" go run ./command_consumer/main.go
	@echo "Command consumer 2 finished."
	
run-command-consumer-3:
	@echo "Running command consumer 3..."
	@APP_NAME="command_consumer_3" go run ./command_consumer/main.go
	@echo "Command consumer 3 finished."
	