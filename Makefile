# FinFlow Backend Root Makefile

.PHONY: test test-cov help

# Default target
help:
	@echo "Available targets:"
	@echo "  test        - Run tests in all services"
	@echo "  test-cov    - Run tests with coverage in all services"

# Run tests in all services
test:
	@echo "=== Running tests in all services ==="
	@for dir in ff-auth ff-files ff-id ff-split ff-tvm; do \
		echo ""; \
		echo "📦 Testing $$dir..."; \
		if [ -f "$$dir/Makefile" ]; then \
			cd $$dir && make test && cd ..; \
		else \
			echo "  ⚠️  No Makefile found in $$dir"; \
		fi; \
	done
	@echo ""
	@echo "✅ All tests completed!"

# Run tests with coverage in all services
test-cov:
	@echo "=== Running tests with coverage in all services ==="
	@for dir in ff-auth ff-files ff-id ff-split ff-tvm; do \
		echo ""; \
		echo "📦 Testing $$dir with coverage..."; \
		if [ -f "$$dir/Makefile" ]; then \
			cd $$dir && make test-cov && cd ..; \
		else \
			echo "  ⚠️  No Makefile found in $$dir"; \
		fi; \
	done
	@echo ""
	@echo "✅ All coverage tests completed!"

