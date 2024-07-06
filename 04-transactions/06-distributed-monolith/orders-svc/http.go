package main

import (
	"encoding/json"
	"net/http"
)

func NewHTTPHandler(
	addDiscountHandler AddDiscountHandler,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /add-discount", func(w http.ResponseWriter, r *http.Request) {
		type payload struct {
			UserID   int `json:"user_id"`
			Discount int `json:"discount"`
		}

		var p payload
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cmd := AddDiscount{
			UserID:   p.UserID,
			Discount: p.Discount,
		}

		err = addDiscountHandler.Handle(r.Context(), cmd)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	})

	return mux
}
