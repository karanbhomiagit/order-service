package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	mgo "gopkg.in/mgo.v2"

	httpDeliver "github.com/karanbhomiagit/order-service/order/delivery/http"
	orderRepo "github.com/karanbhomiagit/order-service/order/repository"
	orderUsecase "github.com/karanbhomiagit/order-service/order/usecase"
)

func main() {
	//Connect to the database
	var db *mgo.Database
	mongodbURL := os.Getenv("MONGODB_URL")
	databaseName := os.Getenv("DATABASE_NAME")
	session, err := mgo.Dial(mongodbURL)
	if err != nil {
		fmt.Println("Unable to connect to the database.")
		log.Fatal(err)
	}
	db = session.DB(databaseName)

	//Initializing the repository
	or := orderRepo.NewMongoOrderRepository(db)

	//Initializing the usecase
	ou := orderUsecase.NewOrderUsecase(or)

	//Initializing the delivery
	httpDeliver.NewOrderHttpHandler(ou)

	//Start the server
	log.Fatal(http.ListenAndServe(port(), nil))
}

func port() string {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	return ":" + port
}
