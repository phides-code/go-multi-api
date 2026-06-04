// Standard API response envelope and API Gateway proxy response builders.
package platform

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

const contentTypeJSON = "application/json"

type APIResponse struct {
	Data  any     `json:"data"`
	Error *string `json:"error"`
}

func JSONResponse(statusCode int, data any, errMsg *string) (events.APIGatewayProxyResponse, error) {
	body, marshalErr := json.Marshal(APIResponse{
		Data:  data,
		Error: errMsg,
	})
	if marshalErr != nil {
		return JSONResponse(http.StatusInternalServerError, nil, strPtr("internal server error"))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": contentTypeJSON,
		},
		Body: string(body),
	}, nil
}

func SuccessResponse(statusCode int, data any) (events.APIGatewayProxyResponse, error) {
	return JSONResponse(statusCode, data, nil)
}

func ErrorResponse(statusCode int, message string) (events.APIGatewayProxyResponse, error) {
	return JSONResponse(statusCode, nil, &message)
}

func strPtr(s string) *string {
	return &s
}
