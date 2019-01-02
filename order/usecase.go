package order

import "github.com/karanbhomiagit/order-service/models"

// Usecase represents the order's business logic as an interface
type Usecase interface {
	AssignByID(string, string) (*map[string]string, error)
	FetchByRange(int, int) ([]models.Order, error)
	Store(*models.OrderRequest) (*models.Order, error)
}
