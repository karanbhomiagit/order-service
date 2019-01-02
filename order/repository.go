package order

import "github.com/karanbhomiagit/order-service/models"

// Repository represents the order's storage/retrieval as an interface
type Repository interface {
	FetchByID(string) (*models.Order, error)
	FetchByRange(int, int) ([]models.Order, error)
	Store(*models.Order) (*models.Order, error)
	UpdateByID(*models.Order) error
}
