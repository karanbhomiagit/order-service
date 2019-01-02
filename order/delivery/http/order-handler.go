package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/karanbhomiagit/order-service/models"
	"github.com/karanbhomiagit/order-service/order"
)

type OrderHttpHandler struct {
	orderUsecase order.Usecase
}

func NewOrderHttpHandler(ou order.Usecase) {
	handler := &OrderHttpHandler{
		orderUsecase: ou,
	}
	http.HandleFunc("/orders/", handler.OrderHandler)
	http.HandleFunc("/orders", handler.OrdersHandler)
}

//OrderHandler is the entrypoint for any requests received for the path "/orders/"
func (h *OrderHttpHandler) OrderHandler(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	//Only PATCH method is supported on /orders/:id
	switch method {
	case http.MethodPatch:
		h.patchOrderByID(w, r)
	default:
		//Return 405 http response code
		respondWithError(w, http.StatusMethodNotAllowed, "Unsupported Request Method")
	}
}

func (h *OrderHttpHandler) patchOrderByID(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	//Extract id from the URL
	id := r.URL.Path[len("/orders/"):]
	fmt.Println("Request PATCH orders/" + id)

	var m map[string]string
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		fmt.Println("Error : ", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	//Make call to usecase layer to assign the order by id
	res, err := h.orderUsecase.AssignByID(id, m["status"])
	if err != nil {
		if err.Error() == "not found" || err.Error() == "Invalid Id" {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	//Marshal the json
	b, err := json.Marshal(res)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	fmt.Println("Error : ", statusCode, message)
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	response := map[string]string{
		"error": message,
	}
	b, _ := json.Marshal(response)
	w.Write(b)
}

//OrdersHandler is the entrypoint for any requests received for the path "/orders"
func (h *OrderHttpHandler) OrdersHandler(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	//Only GET and POST methods are supported on /orders
	switch method {
	case http.MethodGet:
		h.getOrders(w, r)
	case http.MethodPost:
		h.postOrder(w, r)
	default:
		//Return 405 http response code
		respondWithError(w, http.StatusMethodNotAllowed, "Unsupported Request Method")
	}
}

func (h *OrderHttpHandler) getOrders(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request GET /orders")
	//Check if page and limit params were passed
	pageParam := r.URL.Query()["page"]
	limitParam := r.URL.Query()["limit"]
	//Use default values if not passed
	pageParamVal := "1"
	limitParamVal := "10"
	if pageParam != nil {
		pageParamVal = pageParam[0]
	}
	if limitParam != nil {
		limitParamVal = limitParam[0]
	}
	//Convert values of skip and top to integer
	page, err := strconv.Atoi(pageParamVal)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "page parameter should be a number")
		return
	}
	limit, err := strconv.Atoi(limitParamVal)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "limit parameter should be a number")
		return
	}
	//Call helper method to get the orders in specified range
	h.getOrdersInRange(page, limit, w)
}

func (h *OrderHttpHandler) getOrdersInRange(page int, limit int, w http.ResponseWriter) {
	//Make call to usecase layer to fetch the orders
	res, err := h.orderUsecase.FetchByRange(page, limit)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	if res == nil {
		res = make([]models.Order, 0)
	}
	//Marshal the json
	b, err := json.Marshal(res)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (h *OrderHttpHandler) postOrder(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var orderReq models.OrderRequest
	fmt.Println("Request POST /orders")
	//Encode the object received in request body to OrderRequest type json
	if err := json.NewDecoder(r.Body).Decode(&orderReq); err != nil {
		fmt.Println("Error : ", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	//Make call to usecase layer to store the order
	res, err := h.orderUsecase.Store(&orderReq)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	//Marshal the json
	b, err := json.Marshal(res)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
