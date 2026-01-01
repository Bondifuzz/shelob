package bodyParams

import (
	"bytes"
	"encoding/json"
	"encoding/xml"

	"shelob/generateInput"

	"github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"
)

func CreateBodyData(operation *openapi3.Operation) (string, *bytes.Buffer) {
	// As default mimetype

	mimeType := "application/json"
	bodyParams := bytes.NewBuffer([]byte{})
	if operation.RequestBody != nil {
		requestBody := operation.RequestBody.Value
		if requestBody.Required {
			for mimeType, schema := range requestBody.Content {
				bodyData := generateInput.GenerateRandomDataModels(schema.Schema.Value)
				switch mimeType {
				case "application/json":
					bodyParamsJson, err := json.Marshal(bodyData)
					if err != nil {
						log.Error("bodyParams.go	json.Marshal: ", err)
					}

					_, err = bodyParams.Write(bodyParamsJson)

					if err != nil {
						log.Error("bodyParams.go	bodyParams.Write: ", err)
					}

					goto Exit

				case "application/xml":
					// XML marshaling doesn't work well with map[string]interface{}
					// We need to handle this case specially
					bodyParamsXml, err := xml.Marshal(bodyData)
					if err != nil {
						log.Warn("bodyParams.go	xml.Marshal: ", err, " - using empty body for XML")
						// If XML marshaling fails, we'll use an empty body but still continue
						goto Exit
					}

					_, err = bodyParams.Write(bodyParamsXml)
					if err != nil {
						log.Error("bodyParams.go	bodyParams.Write: ", err)
					}

					goto Exit

				case "application/octet-stream":
					_, err := bodyParams.Write(bodyData.([]byte))
					if err != nil {
						log.Error("bodyParams.go	bodyParams.Write: ", err)
					}

					goto Exit

				case "text/plain":
					_, err := bodyParams.Write(bodyData.([]byte))
					if err != nil {
						log.Error("bodyParams.go	bodyParams.Write: ", err)
					}

					goto Exit
				default:
					log.Warn("bodyParams.go	Unresolved mime type: ", mimeType)
				}
			}
		}
	}
Exit:
	return mimeType, bodyParams
}
