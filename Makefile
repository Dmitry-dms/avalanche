.SILENT:


r:
	go run cmd/main.go
cl:
	go run client/client.go
c:
	go run test3-client-gorilla/main.go
run:
	
	# docker run -d --name=pyro -p 4040:4040 pyroscope/pyroscope server
	
	docker run -d --name=ava -p 8446:8000 -p 8667:8090 avalanche
	# docker run -d --rm --name=test1 --ulimit nofile=100000:100009  test
	#docker run -d --name=test2 --ulimit nofile=100000:100009  test
	# docker run -d --name=test3 --ulimit nofile=100000:100009  test
	
run1:
	docker run -d --name=redis -p 6050:6379 redis
	docker run -d --rm --name=ava --network=chat-system -p 8446:8000 -p 8667:8090  avalanche
	docker run --name=monitoring-chat --network=chat-system -p 8780:8780 -d monitoring:latest
start-t:
	#docker run --ulimit nofile=25000:30000 --name=test -d test
	#docker run --ulimit nofile=25000:30000 --name=test2 -d test2
rm:
	docker rm -f ava
	docker rm -f test
	docker rm -f pyro
binsize:
	go tool nm -size client.exe | go-binsize-treemap > binsize.svg
	
build:
	docker rmi avalanche
	docker build -t avalanche:latest .
	make run
# docker run -d --name=pyro -p 4040:4040 pyroscope/pyroscope server
# docker rm -f pyro
# --ulimit nofile=100000:100009