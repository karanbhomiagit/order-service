package usecase

import (
	"context"
	"errors"
	"os"
	"strconv"

	"github.com/karanbhomiagit/order-service/models"
	"github.com/karanbhomiagit/order-service/order"
	"googlemaps.github.io/maps"
)

type OrderUsecase struct {
	orderRepository order.Repository
}

func NewOrderUsecase(or order.Repository) order.Usecase {
	return &OrderUsecase{
		orderRepository: or,
	}
}

const (
	StatusUnassigned = "UNASSIGNED"
	StatusTaken      = "TAKEN"
	StatusSuccess    = "SUCCESS"
)

//AssignByID updates the status of an already existing order
func (ou *OrderUsecase) AssignByID(id string, status string) (*map[string]string, error) {
	//Check request body is correct
	if status == "" || status != StatusTaken {
		return nil, errors.New("This API route only supports assigning of orders. Please provide requested status as TAKEN")
	}
	//Call repository function to fetch order by ID
	order, err := ou.orderRepository.FetchByID(id)
	if err != nil {
		return nil, err
	}
	if (*order).Status != StatusUnassigned {
		return nil, errors.New("Order is already assigned")
	}
	//Update status of the order
	(*order).Status = StatusTaken
	//Call repository function to update the order
	err = ou.orderRepository.UpdateByID(order)
	if err != nil {
		return nil, err
	}
	return &map[string]string{"status": StatusSuccess}, nil
}

//FetchByRange returns a list of orders based on paging parameters
func (ou *OrderUsecase) FetchByRange(page int, limit int) ([]models.Order, error) {
	pageSizeEnv := pageSize()
	pageSize, _ := strconv.Atoi(pageSizeEnv)
	//If limit is zero, return
	if limit == 0 {
		return []models.Order{}, nil
	}
	//If limit is more than page size, change it to page size
	if limit > pageSize {
		limit = pageSize
	}
	//Call repository layer to fetch orders in the range
	return ou.orderRepository.FetchByRange((page-1)*pageSize, limit)
}

//Store calculates distance and stores the order record
func (ou *OrderUsecase) Store(orderReq *models.OrderRequest) (*models.Order, error) {
	distance, err := getDistanceFromExternalService(orderReq.Origin, orderReq.Destination)
	if err != nil {
		return nil, err
	}
	//Create Order record
	order := models.Order{
		Distance: distance,
		Status:   StatusUnassigned,
	}
	//Call repository layer to store the order
	return ou.orderRepository.Store(&order)
}

//getDistanceFromExternalService calls google maps library functions to calculate distance between coordinates
func getDistanceFromExternalService(origin []string, destination []string) (distance int, err error) {
	defer func() {
		// recover from panic if one occured.
		if recover() != nil {
			err = errors.New("Unable to fetch distance from Google APIs. Please ensure data is in correct format")
		}
	}()
	apiKey := os.Getenv("GOOGLE_API_KEY")
	serverURL := os.Getenv("GOOGLE_SERVER_URL")

	c, err := maps.NewClient(maps.WithAPIKey(apiKey), maps.WithBaseURL(serverURL))
	if err != nil {
		return
	}
	r := &maps.DistanceMatrixRequest{
		Origins:      []string{origin[0] + "," + origin[1]},
		Destinations: []string{destination[0] + "," + destination[1]},
	}

	resp, err := c.DistanceMatrix(context.Background(), r)
	if err != nil {
		return
	}
	//Return error if status is other than OK, like ZERO_RESULTS
	if resp.Rows[0].Elements[0].Status != "OK" {
		err = errors.New("Unable to fetch distance from Google APIs, Status : " + resp.Rows[0].Elements[0].Status)
		return
	}
	distance = resp.Rows[0].Elements[0].Distance.Meters
	return
}

func pageSize() string {
	pageSize := os.Getenv("PAGE_SIZE")
	if len(pageSize) == 0 {
		pageSize = "10"
	}
	return pageSize
}
