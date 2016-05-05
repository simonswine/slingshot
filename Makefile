PACKAGE_NAME=github.com/simonswine/slingshot
GO_VERSION=1.6

APP_NAME=slingshot

BUILD_DIR=_build
TEST_DIR=_test

SYMLINK_DEST=Godeps/_workspace/src/$(PACKAGE_NAME)

CONTAINER_DIR=/go/src/${PACKAGE_NAME}

depend:
	which godep || go get github.com/tools/godep
	mkdir -p ${TEST_DIR}

test: test_slingshot test_utils

test_%: depend
	godep go test -coverprofile=${TEST_DIR}/cover.$*.out ./pkg/$*
	go tool cover -html=${TEST_DIR}/cover.$*.out -o ${TEST_DIR}/coverage.$*.html

build: depend
	mkdir -p ${BUILD_DIR}
	godep go build -o ${BUILD_DIR}/${APP_NAME}

build_all: depend
	mkdir -p ${BUILD_DIR}
	GOOS=linux GOARCH=amd64  godep go build -o ${BUILD_DIR}/${APP_NAME}-linux-amd64
	GOOS=darwin GOARCH=amd64 godep go build -o ${BUILD_DIR}/${APP_NAME}-darwin-amd64

all: test build_all

docker:
	# create a container
	$(eval CONTAINER_ID := $(shell docker create \
		-i \
		-w $(CONTAINER_DIR) \
		golang:${GO_VERSION} \
		/bin/bash -c "tar xf - && make all" \
	))
	
	# run build inside container
	tar cf - . | docker start -a -i $(CONTAINER_ID)

	# copy artifacts over
	rm -rf $(BUILD_DIR)/ $(TEST_DIR)/
	docker cp $(CONTAINER_ID):$(CONTAINER_DIR)/$(BUILD_DIR)/ .
	docker cp $(CONTAINER_ID):$(CONTAINER_DIR)/$(TEST_DIR)/ .

	# remove container
	docker rm $(CONTAINER_ID)
