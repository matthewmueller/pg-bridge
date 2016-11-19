release:
	@echo "[+] releasing $(VERSION)"
	@echo "[+] re-generating"
	@go generate ./...
	@echo "[+] building"
	@$(MAKE) build
	@echo "[+] committing"
	@git release $(VERSION)
	@echo "[+] docker build"
	@docker build -t mattmueller/pg-sns-bridge:$(VERSION) .
	@echo "[+] docker push"
	@docker push mattmueller/pg-sns-bridge:$(VERSION)
	@echo "[+] complete"
.PHONY: release

test:
	@go test -cover ./...
.PHONY: test

build:
	mkdir -p releases/$(VERSION)
	@(cd releases/$(VERSION) && gox -os="linux darwin windows openbsd" ./...)
.PHONY: build

clean:
	@git clean -f
.PHONY: clean
