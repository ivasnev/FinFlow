# Проверки всех самостоятельных Go-модулей репозитория.
MODULES := ff-auth ff-files ff-id ff-split ff-tvm ff-common
GO ?= go
GO_TEST_FLAGS ?= -count=1 -timeout=180s

.PHONY: test test-cov test-runner help

help:
	@echo "Available targets:"
	@echo "  test         - Run tests in all six modules"
	@echo "  test-cov     - Run all module tests with coverage"
	@echo "  test-runner  - Check failure propagation and module coverage"

# Подоболочка сохраняет рабочий каталог; ошибка Go или cd немедленно завершает команду.
test:
	@for dir in $(MODULES); do \
		echo "Testing $$dir..."; \
		(cd "$$dir" && $(GO) test $(GO_TEST_FLAGS) ./...) || exit $$?; \
	done
	@echo "All module tests passed."

test-cov:
	@for dir in $(MODULES); do \
		echo "Testing $$dir with coverage..."; \
		(cd "$$dir" && $(GO) test $(GO_TEST_FLAGS) -coverprofile=coverage.out ./...) || exit $$?; \
	done
	@echo "All coverage tests passed; profiles are saved in each module."

test-runner:
	python3 scripts/tests/test_makefile.py
