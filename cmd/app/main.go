package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reviewer-assignment-service/config"
	"reviewer-assignment-service/internal/dto"
)

func main() {
	cfg := config.MustLoad()

	http.HandleFunc("/", func(writer http.ResponseWriter, r *http.Request) {
		er := dto.ErrorResponse{
			Error: dto.ErrorResponseDetail{
				Code:    "NOT_FOUND",
				Message: "not found resource",
			},
		}
		s, err := json.Marshal(er)
		if err != nil {
			return
		}
		writer.Write(s)
	})

	err := http.ListenAndServe(cfg.HTTPServer.Address, nil)
	if err != nil {
		fmt.Println("Ошибка при запуске сервера:", err)
	}
}
