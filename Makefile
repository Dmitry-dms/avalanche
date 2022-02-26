.SILENT:

WS_PORT=8000
MONITORING_PORT=8090
NUM_CONNECT=10000

r:
	go run cmd/main.go
cl:
	go run client/client.go ${WS_PORT} ${NUM_CONNECT}
c:
	go run test3-client-gorilla/main.go
run:
	# docker run -d --name=pyro -p 4040:4040 pyroscope/pyroscope server
	docker run  -d --name=ava -p ${WS_PORT}:8000 -p ${MONITORING_PORT}:8090 avalanche
graph:
	docker run -d -p 9090:9090 prom/prometheus
	docker run -d --name=grafana -p 3000:3000 grafana/grafana	
pr: 
	docker build -t my-prometheus prometheus/
	docker run -d --name=prom -p 9090:9090 my-prometheus
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
	#make cl
build1:
	docker build -t avalanche:latest .
	
goroutine:
	go tool pprof  http://127.0.0.1:${MONITORING_PORT}/goroutine
heap:
	go tool pprof  http://127.0.0.1:${MONITORING_PORT}/heap	
profile:
	go tool pprof  http://127.0.0.1:${MONITORING_PORT}/profile
allocs:
	go tool pprof  http://127.0.0.1:${MONITORING_PORT}/allocs
# docker run -d --name=pyro -p 4040:4040 pyroscope/pyroscope server
# docker rm -f pyro
# --ulimit nofile=100000:100009