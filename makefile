help:
	@echo call make run-NAME-OF-TARGET

run-%:
	go run runners/$*/main.go

test:
	curl -X POST localhost/ -d '{"msg": "something"}'
