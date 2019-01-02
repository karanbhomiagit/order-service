package http

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/karanbhomiagit/order-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/mgo.v2/bson"
)

type MockedOrderUsecase struct {
	mock.Mock
}

func (ou *MockedOrderUsecase) AssignByID(id string, status string) (*map[string]string, error) {
	args := ou.Called(id, status)
	return args.Get(0).(*map[string]string), args.Error(1)
}

func (ou *MockedOrderUsecase) FetchByRange(page int, limit int) ([]models.Order, error) {
	args := ou.Called(page, limit)
	return args.Get(0).([]models.Order), args.Error(1)
}

func (ou *MockedOrderUsecase) Store(orderReq *models.OrderRequest) (*models.Order, error) {
	args := ou.Called(orderReq)
	return args.Get(0).(*models.Order), args.Error(1)
}

/*
	Actual test functions
*/

func TestOrderHandler(t *testing.T) {

	//PATCH /orders/:id tests
	t.Run("Should respond with 405 for GET /orders/id", func(t *testing.T) {
		testObj := new(MockedOrderUsecase)
		handler := &OrderHttpHandler{
			orderUsecase: testObj,
		}

		req, err := http.NewRequest(http.MethodGet, "/orders/1234", strings.NewReader(""))
		assert.NoError(t, err)
		rec := httptest.NewRecorder()

		handler.OrderHandler(rec, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
		body, _ := ioutil.ReadAll(rec.Body)
		assert.Equal(t, `{"error":"Unsupported Request Method"}`, string(body))
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		testObj.AssertExpectations(t)
	})

	t.Run("Should respond with 200 for PATCH /orders/id when assigned successfully", func(t *testing.T) {
		testObj := new(MockedOrderUsecase)
		testObj.On("AssignByID", "1234", "TAKEN").Return(&map[string]string{"status": "SUCCESS"}, nil)
		handler := &OrderHttpHandler{
			orderUsecase: testObj,
		}

		var jsonStr = []byte(`{"status":"TAKEN"}`)
		req, err := http.NewRequest(http.MethodPatch, "/orders/1234", bytes.NewBuffer(jsonStr))
		assert.NoError(t, err)
		rec := httptest.NewRecorder()

		handler.OrderHandler(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body, _ := ioutil.ReadAll(rec.Body)
		assert.Equal(t, `{"status":"SUCCESS"}`, string(body))
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		testObj.AssertExpectations(t)
	})

	t.Run("Should respond with 400 error for PATCH /orders/id when usecase layer returns error", func(t *testing.T) {
		testObj := new(MockedOrderUsecase)
		testObj.On("AssignByID", "1234", "TAKEN").Return(&map[string]string{}, errors.New("Order is already assigned"))
		handler := &OrderHttpHandler{
			orderUsecase: testObj,
		}

		var jsonStr = []byte(`{"status":"TAKEN"}`)
		req, err := http.NewRequest(http.MethodPatch, "/orders/1234", bytes.NewBuffer(jsonStr))
		assert.NoError(t, err)
		rec := httptest.NewRecorder()

		handler.OrderHandler(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		body, _ := ioutil.ReadAll(rec.Body)
		assert.Equal(t, `{"error":"Order is already assigned"}`, string(body))
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		testObj.AssertExpectations(t)
	})

	t.Run("Should respond with 404 error for PATCH /orders/id when usecase layer returns not found", func(t *testing.T) {
		testObj := new(MockedOrderUsecase)
		testObj.On("AssignByID", "1234", "TAKEN").Return(&map[string]string{}, errors.New("not found"))
		handler := &OrderHttpHandler{
			orderUsecase: testObj,
		}

		var jsonStr = []byte(`{"status":"TAKEN"}`)
		req, err := http.NewRequest(http.MethodPatch, "/orders/1234", bytes.NewBuffer(jsonStr))
		assert.NoError(t, err)
		rec := httptest.NewRecorder()

		handler.OrderHandler(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		body, _ := ioutil.ReadAll(rec.Body)
		assert.Equal(t, `{"error":"not found"}`, string(body))
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		testObj.AssertExpectations(t)
	})

}

func TestOrdersHandler(t *testing.T) {

	//POST /orders tests
	t.Run("Should respond with 405 for DELETE /orders", func(t *testing.T) {
		testObj := new(MockedOrderUsecase)
		handler := &OrderHttpHandler{
			orderUsecase: testObj,
		}

		req, err := http.NewRequest(http.MethodDelete, "/orders", strings.NewReader(""))
		assert.NoError(t, err)
		rec := httptest.NewRecorder()

		handler.OrdersHandler(rec, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
		body, _ := ioutil.ReadAll(rec.Body)
		assert.Equal(t, `{"error":"Unsupported Request Method"}`, string(body))
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		testObj.AssertExpectations(t)
	})

	t.Run("Should respond with 400 for POST /orders when request body is not correct", func(t *testing.T) {
		testObj := new(MockedOrderUsecase)
		handler := &OrderHttpHandler{
			orderUsecase: testObj,
		}

		var jsonStr = []byte(`{"origin": "a"}`)
		req, err := http.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(jsonStr))
		assert.NoError(t, err)
		rec := httptest.NewRecorder()

		handler.OrdersHandler(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		body, _ := ioutil.ReadAll(rec.Body)
		assert.Equal(t, `{"error":"Invalid request payload"}`, string(body))
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		testObj.AssertExpectations(t)
	})

	t.Run("Should succesfully POST /orders if usecase layer returns proper response", func(t *testing.T) {
		testObj := new(MockedOrderUsecase)
		testOrderReq := models.OrderRequest{
			Origin:      []string{"1", "2"},
			Destination: []string{"3", "4"},
		}
		testOrderRes := models.Order{
			ID:       bson.ObjectId("12345"),
			Distance: 12345,
			Status:   "UNASSIGNED",
		}
		testObj.On("Store", &testOrderReq).Return(&testOrderRes, nil)
		handler := &OrderHttpHandler{
			orderUsecase: testObj,
		}

		var jsonStr = []byte(`{"origin":["1", "2"], "destination":["3","4"]}`)
		req, err := http.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(jsonStr))
		assert.NoError(t, err)
		rec := httptest.NewRecorder()

		handler.OrdersHandler(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body, _ := ioutil.ReadAll(rec.Body)
		assert.Equal(t, `{"id":"3132333435","distance":12345,"status":"UNASSIGNED"}`, string(body))
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		testObj.AssertExpectations(t)
	})

	t.Run("Should respond with 400 for POST /orders if usecase layer returns error", func(t *testing.T) {
		testObj := new(MockedOrderUsecase)
		testOrderReq := models.OrderRequest{
			Origin:      []string{"1", "2"},
			Destination: []string{"3", "4"},
		}

		testObj.On("Store", &testOrderReq).Return(&models.Order{}, errors.New("Unable to fetch distance from Google APIs"))
		handler := &OrderHttpHandler{
			orderUsecase: testObj,
		}

		var jsonStr = []byte(`{"origin":["1", "2"], "destination":["3","4"]}`)
		req, err := http.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(jsonStr))
		assert.NoError(t, err)
		rec := httptest.NewRecorder()

		handler.OrdersHandler(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		body, _ := ioutil.ReadAll(rec.Body)
		assert.Equal(t, `{"error":"Unable to fetch distance from Google APIs"}`, string(body))
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		testObj.AssertExpectations(t)
	})

	//GET /orders tests
	t.Run("Should return first page of orders if no page/limit specified for GET /orders", func(t *testing.T) {
		testObj := new(MockedOrderUsecase)
		testOrder1 := models.Order{
			ID:       bson.ObjectId("12345"),
			Distance: 12345,
			Status:   "UNASSIGNED",
		}
		testOrder2 := models.Order{
			ID:       bson.ObjectId("12346"),
			Distance: 52345,
			Status:   "TAKEN",
		}
		testObj.On("FetchByRange", 1, 10).Return([]models.Order{testOrder1, testOrder2}, nil)
		handler := &OrderHttpHandler{
			orderUsecase: testObj,
		}

		req, err := http.NewRequest(http.MethodGet, "/orders", strings.NewReader(""))
		assert.NoError(t, err)
		rec := httptest.NewRecorder()

		handler.OrdersHandler(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		body, _ := ioutil.ReadAll(rec.Body)
		assert.Equal(t, `[{"id":"3132333435","distance":12345,"status":"UNASSIGNED"},{"id":"3132333436","distance":52345,"status":"TAKEN"}]`, string(body))
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		testObj.AssertExpectations(t)
	})

	t.Run("Should return error if GET /orders is requested with incorrect page", func(t *testing.T) {
		testObj := new(MockedOrderUsecase)
		handler := &OrderHttpHandler{
			orderUsecase: testObj,
		}

		req, err := http.NewRequest(http.MethodGet, "/orders?page=x", strings.NewReader(""))
		assert.NoError(t, err)
		rec := httptest.NewRecorder()

		handler.OrdersHandler(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		body, _ := ioutil.ReadAll(rec.Body)
		assert.Equal(t, `{"error":"page parameter should be a number"}`, string(body))
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		testObj.AssertExpectations(t)
	})

	t.Run("Should return error if GET /orders is requested with incorrect limit", func(t *testing.T) {
		testObj := new(MockedOrderUsecase)
		handler := &OrderHttpHandler{
			orderUsecase: testObj,
		}

		req, err := http.NewRequest(http.MethodGet, "/orders?page=1&limit=x", strings.NewReader(""))
		assert.NoError(t, err)
		rec := httptest.NewRecorder()

		handler.OrdersHandler(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		body, _ := ioutil.ReadAll(rec.Body)
		assert.Equal(t, `{"error":"limit parameter should be a number"}`, string(body))
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		testObj.AssertExpectations(t)
	})

	t.Run("Should return error for GET /orders if usecase layer returns error", func(t *testing.T) {
		testObj := new(MockedOrderUsecase)

		testObj.On("FetchByRange", 2, 8).Return([]models.Order{}, errors.New("connection lost"))
		handler := &OrderHttpHandler{
			orderUsecase: testObj,
		}

		req, err := http.NewRequest(http.MethodGet, "/orders?page=2&limit=8", strings.NewReader(""))
		assert.NoError(t, err)
		rec := httptest.NewRecorder()

		handler.OrdersHandler(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		body, _ := ioutil.ReadAll(rec.Body)
		assert.Equal(t, `{"error":"connection lost"}`, string(body))
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		testObj.AssertExpectations(t)
	})
}
