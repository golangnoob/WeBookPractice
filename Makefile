.PHONY: docker
docker:
	@rm webooktrial || true
	@GOOS=linux GOARCH=arm go build -o webooktrial .
	@docker rmi -f golangnoob/webooktrial:v0.0.1
	@docker build -t golangnoob/webooktrial:v0.0.1 .