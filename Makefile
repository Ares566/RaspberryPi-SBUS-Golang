dev: 
	go run main.go
deploy:
	env GOOS=linux GOARCH=arm GOARM=5 go build -o firmware ./main.go
	scp firmware pi@raspberrypi.local:~/