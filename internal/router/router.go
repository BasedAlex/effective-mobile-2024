package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/basedalex/effective-mobile-test/docs"
	"github.com/basedalex/effective-mobile-test/internal/api"
	"github.com/basedalex/effective-mobile-test/internal/config"
	"github.com/basedalex/effective-mobile-test/internal/db"
	"github.com/basedalex/effective-mobile-test/internal/types"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	log "github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
)

type HTTPResponse struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type carService interface {
	CreateCar(ctx context.Context, c db.Car) error
	GetCar(ctx context.Context, c *types.GetCarQuery) ([]*db.Car, error)
	UpdateCar(ctx context.Context, c db.Car) (db.Car, error)
	DeleteCar(ctx context.Context, id int) error
}

type Handler struct {
	service carService
	apiClient api.Client
}

func NewServer(ctx context.Context, cfg *config.Config, service carService, apiClient api.Client) error {
	srv := &http.Server{
		Addr:              ":" + cfg.Env.Port,
		Handler:           newRouter(service, apiClient),
		ReadHeaderTimeout: 3 * time.Second,
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*15)

	go func() {
		<-ctx.Done()

		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Warn(err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("error with the server: %w", err)
	}

	return nil
}

func newRouter(service carService, apiClient api.Client) *chi.Mux {
	handler := &Handler{
		service: service,
		apiClient: apiClient,
	}

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders: []string{"Link"},
		AllowCredentials: true,
		MaxAge: 300,
	}))

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8181/swagger/doc.json"),
	))

	r.Post("/api/v1/car", handler.createCar)
	r.Get("/api/v1/car", handler.getCar)
	r.Patch("/api/v1/car/{id}", handler.updateCar)
	r.Delete("/api/v1/car/{id}", handler.deleteCar)

	return r
}

type payload struct {
	RegNums []string `json:"regNums"`
}

type updatePayload struct {
	RegNum string `json:"regNum"`
	Mark string `json:"mark"`
	Model string `json:"model"`
	Year int `json:"year"`
	Name string `json:"name"`
	Surname string `json:"surname"`
	Patronymic string `json:"patronymic"`
}

// @Summary CreateCar
// @Tags car
// @Description create car
// @ID create-car
// @Accept json
// @Produce json
// @Param request body payload true "regnum array to create cars in car catalogue API"
// @Success 201 {integer} integer 1
// @Failure 400 {object} HTTPResponse 
// @Failure 500 {object} HTTPResponse
// @Failure default {object} HTTPResponse
// @Router /api/v1/car [post]
func (h *Handler) createCar(w http.ResponseWriter, r *http.Request) {
	var cars payload

	err := json.NewDecoder(r.Body).Decode(&cars)
	if err != nil {
		log.Warn(err)
		writeErrResponse(w, http.StatusBadRequest, err)
		return
	}

	counter := 0

	for _, v := range cars.RegNums {
		car, err := h.apiClient.GetInfo(r.Context(), v)
		if err != nil {
			log.Warnf("%s: %s", v, err)

			continue
		}
		dbCar := db.Car{
			RegNum: car.RegNum,
			Mark: strings.ToLower(car.Mark),
			Model: strings.ToLower(car.Model),
			Year: car.Year,
			Owner: db.People{
				Name: strings.ToLower(car.Owner.Name),
				Surname: strings.ToLower(car.Owner.Surname),
				Patronymic: strings.ToLower(car.Owner.Patronymic),
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err = h.service.CreateCar(r.Context(), dbCar)
		if err != nil {
			log.Warnf("%s: %s", v, err)

			continue
		}
		counter++
	}

	if counter > 0 {
		writeOkResponse(w, http.StatusCreated, nil)
	} else {
		writeErrResponse(w, http.StatusInternalServerError, err)
	}
}

func getQuery(r *http.Request) types.GetCarQuery {
	yearStr := r.URL.Query().Get("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		log.Warn(err)
		year = 0
	}
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		log.Warn(err)
		limit = 0
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		log.Warn(err)
		offset = 0
	}

	return types.GetCarQuery{
		RegNum: r.URL.Query().Get("regNum"),
		Mark: strings.ToLower(r.URL.Query().Get("mark")),
		Model: strings.ToLower(r.URL.Query().Get("model")),
		Year: year,
		Name: strings.ToLower(r.URL.Query().Get("name")),
		Surname: strings.ToLower(r.URL.Query().Get("surname")),
		Patronymic: strings.ToLower(r.URL.Query().Get("patronymic")),
		Limit: limit,
		Offset: offset,
	}
}

// @Summary GetCar
// @Tags car
// @Description get car
// @ID get-car
// @Accept json
// @Produce json
// @Param q query string false "search options"
// @Success 200 {intgeger} HTTPResponse
// @Failure 500 {object} HTTPResponse
// @Failure default {object} HTTPResponse
// @Router /api/v1/car [get]
func (h *Handler) getCar(w http.ResponseWriter, r *http.Request) {
	payload := getQuery(r)

	data, err := h.service.GetCar(r.Context(), &payload)
	if err != nil {
		writeErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	writeOkResponse(w, http.StatusOK, data)
}

// @Summary UpdateCar
// @Tags car
// @Description update car
// @ID update-car
// @Accept json
// @Produce json
// @Param id path int true "Car ID"
// @Param request body updatePayload false "update options"
// @Success 200 {integer} HTTPResponse
// @Failure 400 {object} HTTPResponse
// @Failure 500 {object} HTTPResponse
// @Failure default {object} HTTPResponse
// @Router /api/v1/car/{id} [patch]
func (h *Handler) updateCar(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeErrResponse(w, http.StatusBadRequest, err)
		return
	}

	var payload updatePayload

	err = json.NewDecoder(r.Body).Decode(&payload)
	
	if err != nil {
		writeErrResponse(w, http.StatusBadRequest, err)
		return
	}

	car := db.Car{
		ID: id,
		RegNum: strings.ToLower(payload.RegNum),
		Model: strings.ToLower(payload.Model),
		Mark: strings.ToLower(payload.Mark),
		Year: payload.Year,
		Owner: db.People{
			Name: strings.ToLower(payload.Name),
			Surname: strings.ToLower(payload.Surname),
			Patronymic: strings.ToLower(payload.Patronymic),
		},
	}

	updatedCar, err := h.service.UpdateCar(r.Context(), car)
	if err != nil {
		writeErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	writeOkResponse(w, http.StatusOK, updatedCar)
}

// @Summary DeleteCar
// @Tags car
// @Description delete car
// @ID delete-car
// @Accept json
// @Produce json
// @Param id path int true "Car ID"
// @Success 204 {integer} HTTPResponse
// @Failure 500 {object} HTTPResponse
// @Failure default {object} HTTPResponse
// @Router /api/v1/car/{id} [delete]
func (h *Handler) deleteCar(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeErrResponse(w, http.StatusBadRequest, err)
		return
	}

	err = h.service.DeleteCar(r.Context(), id)
	if err != nil {
		writeErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	writeOkResponse(w, http.StatusNoContent, nil)
}

func writeOkResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		err := json.NewEncoder(w).Encode(HTTPResponse{Data: data})
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func writeErrResponse(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	log.Error(err)

	jsonErr := json.NewEncoder(w).Encode(HTTPResponse{Error: err.Error()})
	if jsonErr != nil {
		log.Error(jsonErr)
	}
}