package usecase

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/karanbhomiagit/order-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/mgo.v2/bson"
)

type MockedOrderRepository struct {
	mock.Mock
}

func (or *MockedOrderRepository) FetchByID(id string) (*models.Order, error) {
	args := or.Called(id)
	return args.Get(0).(*models.Order), args.Error(1)
}

func (or *MockedOrderRepository) UpdateByID(order *models.Order) error {
	args := or.Called(order)
	return args.Error(0)
}

func (or *MockedOrderRepository) FetchByRange(skip int, limit int) ([]models.Order, error) {
	args := or.Called(skip, limit)
	return args.Get(0).([]models.Order), args.Error(1)
}

func (or *MockedOrderRepository) Store(order *models.Order) (*models.Order, error) {
	args := or.Called(order)
	return args.Get(0).(*models.Order), args.Error(1)
}

/*
	Actual test functions
*/

func TestAssignByID(t *testing.T) {

	t.Run("Successfully assign an order", func(t *testing.T) {
		testObj := new(MockedOrderRepository)
		testOrder := models.Order{
			ID:       "5c2b2aaf4530558539f91859",
			Distance: 12345,
			Status:   "UNASSIGNED",
		}
		testObj.On("FetchByID", "5c2b2aaf4530558539f91859").Return(&testOrder, nil)
		changedTestOrder := models.Order{
			ID:       "5c2b2aaf4530558539f91859",
			Distance: 12345,
			Status:   "TAKEN",
		}
		testObj.On("UpdateByID", &changedTestOrder).Return(nil)

		orderUsecase := NewOrderUsecase(testObj)
		response, err := orderUsecase.AssignByID("5c2b2aaf4530558539f91859", "TAKEN")
		assert := assert.New(t)
		assert.Nil(err)
		assert.Equal(response, &map[string]string{"status": "SUCCESS"})
		testObj.AssertExpectations(t)
	})

	t.Run("Return error for wrong status request", func(t *testing.T) {
		testObj := new(MockedOrderRepository)
		orderUsecase := NewOrderUsecase(testObj)
		_, err := orderUsecase.AssignByID("5c2b2aaf4530558539f91859", "RELEIVE")
		assert := assert.New(t)
		if assert.NotNil(err) {
			assert.Equal("This API route only supports assigning of orders. Please provide requested status as TAKEN", err.Error())
		}
		testObj.AssertExpectations(t)
	})

	t.Run("Return error if order already assigned", func(t *testing.T) {
		testObj := new(MockedOrderRepository)
		testOrder := models.Order{
			ID:       "5c2b2aaf4530558539f91859",
			Distance: 12345,
			Status:   "TAKEN",
		}
		testObj.On("FetchByID", "5c2b2aaf4530558539f91859").Return(&testOrder, nil)

		orderUsecase := NewOrderUsecase(testObj)
		_, err := orderUsecase.AssignByID("5c2b2aaf4530558539f91859", "TAKEN")
		assert := assert.New(t)
		if assert.NotNil(err) {
			assert.Equal("Order is already assigned", err.Error())
		}
		testObj.AssertExpectations(t)
	})

	t.Run("Return error if order does not exist", func(t *testing.T) {
		testObj := new(MockedOrderRepository)
		testObj.On("FetchByID", "5c2b2aaf4530558539f91859").Return(&models.Order{}, errors.New("not found"))

		orderUsecase := NewOrderUsecase(testObj)
		_, err := orderUsecase.AssignByID("5c2b2aaf4530558539f91859", "TAKEN")
		assert := assert.New(t)
		if assert.NotNil(err) {
			assert.Equal("not found", err.Error())
		}
		testObj.AssertExpectations(t)
	})

	t.Run("Return error if order update fails", func(t *testing.T) {
		testObj := new(MockedOrderRepository)
		testOrder := models.Order{
			ID:       "5c2b2aaf4530558539f91859",
			Distance: 12345,
			Status:   "UNASSIGNED",
		}
		testObj.On("FetchByID", "5c2b2aaf4530558539f91859").Return(&testOrder, nil)
		changedTestOrder := models.Order{
			ID:       "5c2b2aaf4530558539f91859",
			Distance: 12345,
			Status:   "TAKEN",
		}
		testObj.On("UpdateByID", &changedTestOrder).Return(errors.New("connection lost"))

		orderUsecase := NewOrderUsecase(testObj)
		_, err := orderUsecase.AssignByID("5c2b2aaf4530558539f91859", "TAKEN")
		assert := assert.New(t)
		if assert.NotNil(err) {
			assert.Equal("connection lost", err.Error())
		}
		testObj.AssertExpectations(t)
	})

}

func TestFetchByRange(t *testing.T) {

	t.Run("Successfully fetch orders in range", func(t *testing.T) {
		testObj := new(MockedOrderRepository)
		testOrder1 := models.Order{
			ID:       "5c2b2aaf4530558539f91859",
			Distance: 12345,
			Status:   "UNASSIGNED",
		}
		testOrder2 := models.Order{
			ID:       "5c2b2aaf4530558539f91858",
			Distance: 52345,
			Status:   "TAKEN",
		}
		testObj.On("FetchByRange", 0, 10).Return([]models.Order{testOrder1, testOrder2}, nil)

		orderUsecase := NewOrderUsecase(testObj)
		os.Setenv("PAGE_SIZE", "10")
		res, err := orderUsecase.FetchByRange(1, 10)
		assert := assert.New(t)
		assert.Nil(err)
		if assert.NotNil(res) {
			assert.Equal(2, len(res))
		}
		testObj.AssertExpectations(t)
	})

	t.Run("Successfully adjust limit to pagesize and fetch orders in range", func(t *testing.T) {
		testObj := new(MockedOrderRepository)
		testOrder1 := models.Order{
			ID:       "5c2b2aaf4530558539f91859",
			Distance: 12345,
			Status:   "UNASSIGNED",
		}
		testOrder2 := models.Order{
			ID:       "5c2b2aaf4530558539f91858",
			Distance: 52345,
			Status:   "TAKEN",
		}
		testObj.On("FetchByRange", 10, 10).Return([]models.Order{testOrder1, testOrder2}, nil)

		orderUsecase := NewOrderUsecase(testObj)
		os.Setenv("PAGE_SIZE", "10")
		res, err := orderUsecase.FetchByRange(2, 11)
		assert := assert.New(t)
		assert.Nil(err)
		if assert.NotNil(res) {
			assert.Equal(2, len(res))
		}
		testObj.AssertExpectations(t)
	})

	t.Run("Successfully return empty list if limit is 0", func(t *testing.T) {
		testObj := new(MockedOrderRepository)
		orderUsecase := NewOrderUsecase(testObj)
		os.Setenv("PAGE_SIZE", "10")
		res, err := orderUsecase.FetchByRange(2, 0)
		assert := assert.New(t)
		assert.Nil(err)
		if assert.NotNil(res) {
			assert.Equal(0, len(res))
		}
		testObj.AssertExpectations(t)
	})

}

func TestStore(t *testing.T) {

	t.Run("Successfully save order", func(t *testing.T) {
		response := `{
			"destination_addresses" : [
				 "Av Instituto Politécnico Nacional 3600, San Pedro Zacatenco, 07360 Ciudad de México, CDMX, Mexico"
			],
			"origin_addresses" : [
				 "Cto. Fuentes del Pedregal 555, Los Framboyanes, 14150 Ciudad de México, CDMX, Mexico"
			],
			"rows" : [
				 {
						"elements" : [
							 {
									"distance" : {
										 "text" : "30.5 km",
										 "value" : 30539
									},
									"duration" : {
										 "text" : "50 mins",
										 "value" : 3001
									},
									"duration_in_traffic" : {
										 "text" : "51 mins",
										 "value" : 3040
									},
									"status" : "OK"
							 }
						]
				 }
			],
			"status" : "OK"
		}`
		server := mockServer(200, response)
		fmt.Println("server ", server.URL)
		os.Setenv("GOOGLE_SERVER_URL", server.URL)
		os.Setenv("GOOGLE_API_KEY", apiKey)
		defer server.Close()

		testObj := new(MockedOrderRepository)
		testOrder := models.Order{
			Distance: 30539,
			Status:   "UNASSIGNED",
		}
		testOrderResponse := models.Order{
			ID:       "5c2b2aaf4530558539f91858",
			Distance: 30539,
			Status:   "UNASSIGNED",
		}
		testObj.On("Store", &testOrder).Return(&testOrderResponse, nil)

		orderUsecase := NewOrderUsecase(testObj)
		orderReq := models.OrderRequest{
			Origin:      []string{"1", "2"},
			Destination: []string{"3", "4"},
		}
		resp, err := orderUsecase.Store(&orderReq)
		assert := assert.New(t)
		assert.Nil(err)
		if assert.NotNil(resp) {
			assert.Equal(30539, resp.Distance)
			assert.Equal("UNASSIGNED", resp.Status)
			assert.Equal(bson.ObjectId("5c2b2aaf4530558539f91858"), resp.ID)
		}
		testObj.AssertExpectations(t)
	})

	t.Run("Return error when origin coordinates in wrong format", func(t *testing.T) {
		response := `{
			"destination_addresses" : [
				 "Av Instituto Politécnico Nacional 3600, San Pedro Zacatenco, 07360 Ciudad de México, CDMX, Mexico"
			],
			"origin_addresses" : [
				 "Cto. Fuentes del Pedregal 555, Los Framboyanes, 14150 Ciudad de México, CDMX, Mexico"
			],
			"rows" : [
				 {
						"elements" : [
							 {
									"distance" : {
										 "text" : "30.5 km",
										 "value" : 30539
									},
									"duration" : {
										 "text" : "50 mins",
										 "value" : 3001
									},
									"duration_in_traffic" : {
										 "text" : "51 mins",
										 "value" : 3040
									},
									"status" : "OK"
							 }
						]
				 }
			],
			"status" : "OK"
		}`
		server := mockServer(200, response)
		fmt.Println("server ", server.URL)
		os.Setenv("GOOGLE_SERVER_URL", server.URL)
		os.Setenv("GOOGLE_API_KEY", apiKey)
		defer server.Close()

		testObj := new(MockedOrderRepository)

		orderUsecase := NewOrderUsecase(testObj)
		orderReq := models.OrderRequest{
			Origin:      []string{"1"},
			Destination: []string{"3", "4"},
		}
		_, err := orderUsecase.Store(&orderReq)
		assert := assert.New(t)
		if assert.NotNil(err) {
			assert.Equal("Unable to fetch distance from Google APIs. Please ensure data is in correct format", err.Error())
		}
		testObj.AssertExpectations(t)
	})

	t.Run("Return error when Google APIs return ZERO_RESULTS", func(t *testing.T) {
		response := `{
			"destination_addresses" : [
				 "Av Instituto Politécnico Nacional 3600, San Pedro Zacatenco, 07360 Ciudad de México, CDMX, Mexico"
			],
			"origin_addresses" : [
				 "Cto. Fuentes del Pedregal 555, Los Framboyanes, 14150 Ciudad de México, CDMX, Mexico"
			],
			"rows" : [
				 {
						"elements" : [
							 {
									"status" : "ZERO_RESULTS"
							 }
						]
				 }
			],
			"status" : "OK"
		}`
		server := mockServer(200, response)
		fmt.Println("server ", server.URL)
		os.Setenv("GOOGLE_SERVER_URL", server.URL)
		os.Setenv("GOOGLE_API_KEY", apiKey)
		defer server.Close()

		testObj := new(MockedOrderRepository)

		orderUsecase := NewOrderUsecase(testObj)
		orderReq := models.OrderRequest{
			Origin:      []string{"1", "2"},
			Destination: []string{"3", "4"},
		}
		_, err := orderUsecase.Store(&orderReq)
		assert := assert.New(t)
		if assert.NotNil(err) {
			assert.Equal("Unable to fetch distance from Google APIs, Status : ZERO_RESULTS", err.Error())
		}
		testObj.AssertExpectations(t)
	})

	t.Run("Return error if save operation fails", func(t *testing.T) {
		response := `{
			"destination_addresses" : [
				 "Av Instituto Politécnico Nacional 3600, San Pedro Zacatenco, 07360 Ciudad de México, CDMX, Mexico"
			],
			"origin_addresses" : [
				 "Cto. Fuentes del Pedregal 555, Los Framboyanes, 14150 Ciudad de México, CDMX, Mexico"
			],
			"rows" : [
				 {
						"elements" : [
							 {
									"distance" : {
										 "text" : "30.5 km",
										 "value" : 30539
									},
									"duration" : {
										 "text" : "50 mins",
										 "value" : 3001
									},
									"duration_in_traffic" : {
										 "text" : "51 mins",
										 "value" : 3040
									},
									"status" : "OK"
							 }
						]
				 }
			],
			"status" : "OK"
		}`
		server := mockServer(200, response)
		fmt.Println("server ", server.URL)
		os.Setenv("GOOGLE_SERVER_URL", server.URL)
		os.Setenv("GOOGLE_API_KEY", apiKey)
		defer server.Close()

		testObj := new(MockedOrderRepository)
		testOrder := models.Order{
			Distance: 30539,
			Status:   "UNASSIGNED",
		}
		testObj.On("Store", &testOrder).Return(&models.Order{}, errors.New("connection lost"))

		orderUsecase := NewOrderUsecase(testObj)
		orderReq := models.OrderRequest{
			Origin:      []string{"1", "2"},
			Destination: []string{"3", "4"},
		}
		_, err := orderUsecase.Store(&orderReq)
		assert := assert.New(t)
		if assert.NotNil(err) {
			assert.Equal("connection lost", err.Error())
		}
		testObj.AssertExpectations(t)
	})
}

const apiKey = "AIzaNotReallyAnAPIKey"

type countingServer struct {
	s          *httptest.Server
	successful int
}

func mockServer(code int, body string) *httptest.Server {
	serv := mockServerForQuery("", code, body)
	return serv.s
}

func mockServerForQuery(query string, code int, body string) *countingServer {
	server := &countingServer{}
	server.s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.successful++
		w.WriteHeader(code)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		fmt.Fprintln(w, body)
	}))
	return server
}
