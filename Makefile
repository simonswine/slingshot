PACKAGE_NAME=github.com/simonswine/slingshot
GO_VERSION=1.6

APP_NAME=slingshot

BUILD_DIR=_build
TEST_DIR=_test

SYMLINK_DEST=Godeps/_workspace/src/$(PACKAGE_NAME)

CONTAINER_DIR=/go/src/${PACKAGE_NAME}

depend:
	which godep || go get github.com/tools/godep
	#$(shell mkdir -p `dirname $(SYMLINK_DEST)`)
	#rm -f $(SYMLINK_DEST)
	#ln -s ../../../../../ $(SYMLINK_DEST)


test_1: depend
	mkdir -p ${TEST_DIR}
	godep go test -coverprofile=${TEST_DIR}/cover.slingshot.out ./pkg/slingshot
	godep go test -coverprofile=${TEST_DIR}/cover.utils.out ./pkg/utils
	go tool cover -html=${TEST_DIR}/cover.slingshot.out -o ${TEST_DIR}/coverage.slingshot.html
	go tool cover -html=${TEST_DIR}/cover.utils.out     -o ${TEST_DIR}/coverage.utils.html

test: test_slingshot test_utils

test_%: depend
	cd pkg/$* && godep go test

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
