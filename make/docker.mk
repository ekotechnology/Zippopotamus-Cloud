.PHONY: api-docker
api-docker:
	@echo "Building api Docker image"
	docker-compose build --build-arg version=$(version) --build-arg gitCommit=$(SHA) --build-arg buildDate=$(date) api

.PHONY: parser-docker
parser-docker:
	@echo "Building parser Docker image"
	docker-compose build --build-arg version=$(version) --build-arg gitCommit=$(SHA) --build-arg buildDate=$(date) parser
