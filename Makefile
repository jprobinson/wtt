# Builds the API server for Linux
build:
	@cd ./cmd/server; \
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .

build-image: build
	@say -v Samantha "compiled\!"
	@docker build  --tag gcr.io/wheresthetrain-nyc/wtt:local .; say -v Samantha "image built\!";

# Cross compiles the server for Linux, builds a "local" docker image, pushes it and deploys it to dev
deploy-local: build-image
	@docker push gcr.io/wheresthetrain-nyc/wtt:local; say -v Samantha "image pushed\!";
	@cd cmd/server; \
	gcloud app deploy ./prd.yaml --version "local-`date +'%s'`"  --image-url gcr.io/wheresthetrain-nyc/wtt:local --project wheresthetrain-nyc --quiet; say -v Samantha "deployed\!";
	@rm cmd/server/server;

# Runs the go vet command, will be a dependency for any test.
vet:
	@go vet .
	@go vet ./cmd/...

# Builds the local server to run on OS X.
build-local:
	@cd ./cmd/server; \
	go build .

# Sets up the environment and emulator, builds and runs the local server on port 8080.
run: build-local
	@export GOOGLE_CLOUD_PROJECT=wheresthetrain-nyc; \
	export MTA_KEY=$(MTA_KEY); \
	export BASE_PATH="`pwd`"; \
	./cmd/server/server

watch:
	@export GO111MODULE=off; \
	go get github.com/radovskyb/watcher/...; \
	watcher -startcmd -ignore="./cmd/server/server" -cmd="make run" -keepalive=false ./

# A convenience command for devs to occasionally add their GCP credentials to the local environment.
gcloud-login:
	@gcloud auth application-default login
