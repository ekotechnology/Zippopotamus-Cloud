datafiles = all_countries.txt gb.txt nl.txt

.PHONY: alldata
alldata: all_countries.txt gb.txt nl.txt

all_countries.txt:
	curl http://download.geonames.org/export/zip/allCountries.zip > allCountries.zip
	unzip allCountries.zip allCountries.txt
	mv allCountries.txt all_countries.txt
	rm allCountries.zip

gb.txt:
	curl http://download.geonames.org/export/zip/GB_full.csv.zip > gb.zip
	unzip gb.zip GB_full.txt
	mv GB_full.txt gb.txt
	rm gb.zip

nl.txt:
	curl http://download.geonames.org/export/zip/NL_full.csv.zip > nl.zip
	unzip nl.zip NL_full.txt
	mv NL_full.txt nl.txt
	rm nl.zip


.PHONY: build-data
build-data: alldata build-core build-gb build-nl
	docker-compose exec redisearch redis-cli -h redisearch save
	docker-compose build data

.PHONY:run-redisearch
run-redisearch:
	docker-compose up -d redisearch

.PHONY: build-core
build-core: alldata run-redisearch
	grep -v -e "^GB" -e "^NL" all_countries.txt | ./cmd/zp-parser/zp-parser -redis-addr 127.0.0.1:6380

.PHONY:build-gb
build-gb: alldata run-redisearch
	cat gb.txt | ./cmd/zp-parser/zp-parser -redis-addr 127.0.0.1:6380

.PHONY: build-nl
build-nl: alldata run-redisearch
	cat nl.txt | ./cmd/zp-parser/zp-parser -redis-addr 127.0.0.1:6380

.PHONY: clean-data
clean-data:
	-rm $(datafiles)