package enhancer

import (
	"encoding/json"
	"net/http"
)

/*
Response - ответ клиенту с code и resp
*/
func Response(request *http.Request, writer http.ResponseWriter, resp map[string]interface{}, code int) {
	response, _ := json.Marshal(resp)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(code)
	writer.Write(response)
}
