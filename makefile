help:
	@echo call make run-NAME-OF-TARGET

run-%:
	go run runners/$*/main.go

test:
	curl -X POST -H "header:123" localhost/some-path -d '{"msg": "something"}'
