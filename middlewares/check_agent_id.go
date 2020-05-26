package middlewares

import (
	"log"
	"net/http"

	"github.com/kubitre/diplom/enhancer"
)

/*CheckAgentID - проверка подставленного agentID в запрос*/
func CheckAgentID(agentID string, next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		log.Println("start checking agent id")
		agentFromRequest := request.URL.Query().Get("runner_id")
		if agentFromRequest != agentID {
			enhancer.Response(request, writer, map[string]interface{}{
				"context": map[string]string{
					"runner": "master",
				},
				"status": "runner can not execute your request. Invalid agent_id",
			}, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
