v2:
	rm -rf build
	mkdir -p build
	api-console build -t "RAML 1.0" -a api_v2.raml; sed -i 's/World Music API/Hub3 API Documentation/g' build/index.html   

serve:
	api-console serve -o
