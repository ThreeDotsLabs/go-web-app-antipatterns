.PHONY: test
test:
	@cd tests && go test -count=1

.PHONY: mysql
mysql:
	@docker-compose exec mysql mysql

.PHONY: cloc
cloc:
	@for dir in 0*; do \
		echo "### $$dir"; \
		cloc --exclude-lang=Markdown \
		--quiet \
		--hide-rate \
		--md \
		$$dir; \
		echo; \
	done
