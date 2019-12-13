.PHONY: $(targets)
$(targets): format
	@$(eval target = $@)
	@$(eval zptarget = $(addprefix zp-,$(target)))
	@echo "Building $(target)"
	go build -ldflags "$(FLAGS)" -o cmd/$(zptarget)/$(zptarget) $(package)/cmd/$(zptarget)

.PHONY: clean-bins
clean-bins:
	-rm cmd/zp-api/zp-api
	-rm cmd/zp-parser/zp-parser