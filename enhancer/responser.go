package enhancer

import (
	"encoding/json"
	"net/http"
)

/*
Response - ответ клиенту с code и resp
*/
func Response(request *http.Request, writer http.ResponseWriter, resp map[string]interface{}, code int) {
	if request.Header.Get("Content-Type") != "application/json" {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusConflict)
		response, _ := json.Marshal(map[string]string{"error": "you packet in non json format!"})
		writer.Write(response)
	} else {
		response, _ := json.Marshal(resp)
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(code)
		writer.Write(response)
	}
}
