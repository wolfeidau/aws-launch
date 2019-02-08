mocks:
	mockery -all -dir ./pkg/launcher
	mockery -all -dir ./pkg/cwlogs
.PHONY: mocks