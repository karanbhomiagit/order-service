package repository

import (
	"errors"

	"github.com/karanbhomiagit/order-service/models"
	"github.com/karanbhomiagit/order-service/order"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type mongoOrderRepository struct {
	Conn *mgo.Database
}

const (
	COLLECTION = "orders"
)

func NewMongoOrderRepository(Conn *mgo.Database) order.Repository {
	return &mongoOrderRepository{Conn}
}

//FetchByID validates the provided ID and finds the corresponding document in the database
func (or *mongoOrderRepository) FetchByID(id string) (*models.Order, error) {
	var order models.Order
	//If the ID passed is not a valid Object ID, return error
	isValidID := bson.IsObjectIdHex(id)
	if isValidID != true {
		return nil, errors.New("Invalid Id")
	}
	//Find document in DB by ID
	err := or.Conn.C(COLLECTION).FindId(bson.ObjectIdHex(id)).One(&order)
	return &order, err
}

//UpdateByID finds the corresponding document in the database and updates it
func (or *mongoOrderRepository) UpdateByID(order *models.Order) error {
	//Update document in DB by ID
	return or.Conn.C(COLLECTION).UpdateId((*order).ID, order)
}

//FetchByRange finds the corresponding documents in the database for a particular range
func (or *mongoOrderRepository) FetchByRange(skip int, limit int) ([]models.Order, error) {
	var orders []models.Order
	//Find documents
	err := or.Conn.C(COLLECTION).Find(bson.M{}).Skip(skip).Limit(limit).All(&orders)
	return orders, err
}

//Store generates a new object id and inserts the document into the database
func (or *mongoOrderRepository) Store(order *models.Order) (*models.Order, error) {
	(*order).ID = bson.NewObjectId()
	err := or.Conn.C(COLLECTION).Insert(order)
	return order, err
}
