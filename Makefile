.SILENT:
include .env

#WS_PORT=8002
#MONITORING_PORT=8090
NUM_CONNECT=10000
WS_ADDR=172.17.0.1
MIN=0
DELTA=20000

r:
	go run cmd/main.go
cl:
	go run client/client.go ${WS_PORT} ${NUM_CONNECT}
c:
	docker run --rm --net testnet --ip 172.18.0.10 --ulimit nofile=100000:150000 client ./goapp ${WS_ADDR}:${WS_PORT} ${MIN} ${DELTA}
cw:
	go run ./client/client.go ${WS_ADDR}:${WS_PORT} ${MIN} ${DELTA}
cw2:
	go run ./client2/client.go ${WS_ADDR}:${WS_PORT} ${MIN} ${DELTA}
setup:
	bash setup.sh ${WS_ADDR}:${WS_PORT} ${MIN} ${DELTA} 3
run:
	# docker run -d --name=pyro -p 4040:4040 pyroscope/pyroscope server
	docker run  --name=ava --net testnet --ulimit nofile=220000:230000 -p ${WS_PORT}:8000 -p ${MONITORING_PORT}:8090 avalanche
graph:
	#docker run -d -p 9090:9090 my-prometheus
	docker run -d --name=grafana -p 3000:3000 grafana/grafana	
pr: 
	docker build -t my-prometheus prometheus/
	docker run -d --name=prom -p 9090:9090 my-prometheus
run1:
	docker run -d --name=redis -p 6050:6379 redis
	docker run -d --rm --name=ava --network=chat-system -p 8446:8000 -p 8667:8090  avalanche
	docker run --name=monitoring-chat --network=chat-system -p 8780:8780 -d monitoring:latest
start-t:
	docker run -d --name=t1 --rm --net testnet --ip 172.18.0.5 --ulimit nofile=40000:45000 client ./goapp 172.17.0.1:8000 ${MIN} ${DELTA}
	docker run -d --name=t2 --rm --net testnet --ip 172.18.0.6 --ulimit nofile=40000:45000 client ./goapp 172.17.0.1:8000  $(shell expr $(DELTA) + 1 ) ${DELTA}
	docker run -d --name=t3 --rm --net testnet --ip 172.18.0.7 --ulimit nofile=40000:45000 client ./goapp 172.17.0.1:8000  $(shell expr $(DELTA) + $(DELTA) + 1 ) ${DELTA}
	docker run -d --name=t4 --rm --net testnet --ip 172.18.0.8 --ulimit nofile=40000:45000 client ./goapp 172.17.0.1:8000  $(shell expr $(DELTA) + $(DELTA) + $(DELTA) + 1 ) ${DELTA}
	docker run -d --name=t5 --rm --net testnet --ip 172.18.0.9 --ulimit nofile=40000:45000 client ./goapp 172.17.0.1:8000  $(shell expr $(DELTA) + $(DELTA) + $(DELTA) + $(DELTA) + 1 ) ${DELTA}
destr:
	docker rm -vf $(docker ps -q --filter label=go-websockets)

stop-t:
	docker stop t1
	docker stop t2
	docker stop t3
	docker stop t4
	docker stop t5
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
start:
	docker start redis
	docker start prom
	docker start grafana
stop:
	docker stop redis
	docker stop prom
	docker stop grafana
# docker run -d --name=pyro -p 4040:4040 pyroscope/pyroscope server
# docker rm -f pyro
# --ulimit nofile=100000:100009