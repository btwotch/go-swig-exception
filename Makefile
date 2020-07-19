all:
	docker build -t swig-go .
	docker run -it swig-go
