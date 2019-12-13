.PHONY: admincodes
admincodes: admin_1.txt admin_2.txt countries.txt
	go run internal/gen.go
	make clean-admincodes

admin_1.txt:
	curl http://download.geonames.org/export/dump/admin1CodesASCII.txt > admin_1.txt

admin_2.txt:
	curl http://download.geonames.org/export/dump/admin2Codes.txt > admin_2.txt

countries.txt:
	curl http://download.geonames.org/export/dump/countryInfo.txt > countries.txt

.PHONY: clean-admincodes
clean-admincodes:
	-rm admin_1.txt
	-rm admin_2.txt
	-rm countries.txt

.PHONY: clean-generated-admincodes
clean-generated-admincodes:
	-rm internal/static_*