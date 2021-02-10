.SILENT:


run:
	# docker run -d --name=pyro -p 4040:4040 pyroscope/pyroscope server
	docker run -d --rm --name=ava -p 8080:8080 -p 8090:8090 --ulimit nofile=100000:100009 avalanche
	# docker run -d --rm --name=test1 --ulimit nofile=100000:100009  test
	#docker run -d --name=test2 --ulimit nofile=100000:100009  test
	# docker run -d --name=test3 --ulimit nofile=100000:100009  test
	
start-t:
	#docker run --ulimit nofile=25000:30000 --name=test -d test
	#docker run --ulimit nofile=25000:30000 --name=test2 -d test2
rm:
	docker rm -f ava
	docker rm -f test
	docker rm -f pyro
	

# docker run -d --name=pyro -p 4040:4040 pyroscope/pyroscope server
# docker rm -f pyro