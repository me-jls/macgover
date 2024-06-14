swagger: 
	go get -u github.com/swaggo/swag/cmd/swag

run:
	swag init
	go run main.go jwt.go ldap.go database.go

init: swagger run

build:
	docker build -t macgover:beta .

