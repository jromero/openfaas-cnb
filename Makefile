build/:
	mkdir -p build/

clean:
	rm -rf build/

test:
	@echo "> Runing tests..."
	./test/test.sh

package-deps:
	@echo "> Ensuring packaging dependencies are present..."
	@which yj
	@which jq

package: package-deps build/
	@echo "> Packaging..."
	@tar cvzf build/openfaas-cnb-`yj -tj < buildpack.toml | jq -r '.buildpack.version'`.tgz buildpack.toml bin/
	@ls build/

.PHONY: clean package package-deps test