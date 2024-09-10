all: fe be 

be: 
	cd be && go run main.go

fe: 
	cd fe && npm run start