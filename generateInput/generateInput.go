package generateInput

import (
	"encoding/base64"
	"math"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"
)

func GenerateRandomDataModels(schema *openapi3.Schema) interface{} {
	switch schema.Type {
	case "string":
		if schema.Pattern != "" {
			return CheckStringPattern(schema.Pattern)
		}
		if schema.Format != "" {
			return CheckStringFormat(schema.Format)
		}
		if schema.Enum != nil {
			if schema.Default == nil {
				return schema.Enum[0]
			}
			return schema.Default
		}
		if schema.Example != nil {
			return schema.Example
		}
		return gofakeit.Generate("????????????????????????????????????????????????????????????????????????????????????????????????????")
	case "number":
		return CheckNumberFormat(schema.Format)
	case "integer":
		return CheckIntegerFormat(schema.Format)
	case "boolean":
		return gofakeit.Bool()
	case "array":
		var array []interface{}
		array = append(array, GenerateRandomDataModels(schema.Items.Value))
		return array
	case "object":
		objects := make(map[string]interface{})
		for property, schemaInternal := range schema.Properties {
			objects[property] = GenerateRandomDataModels(schemaInternal.Value)
		}
		return objects
	default:
		log.Warn("Unresolved schema type:", schema.Type)
	}
	return ""
}

func CheckNumberFormat(format string) interface{} {
	switch format {
	case "float":
		return gofakeit.Float32()
		//        return gofakeit.Float32Range(float32(*min), float32(*max))
	case "double":
		return gofakeit.Float64()
		//        return gofakeit.Float64Range(*min, *max)
	default:
		return gofakeit.Number(int(math.Inf(-1)), int(math.Inf(1)))
	}
}

func CheckIntegerFormat(format string) interface{} {
	switch format {
	case "int32":
		return gofakeit.Int32()
		//        return strconv.FormatInt(int64(gofakeit.Int32()),10)
	case "int64":
		return gofakeit.Int64()
		//        return strconv.FormatInt(gofakeit.Int64(), 10)
	default:
		return gofakeit.IntRange(int(math.Inf(-1)), int(math.Inf(1)))
		//        return strconv.FormatInt(int64(gofakeit.IntRange(int(math.Inf(-1)), int(math.Inf(1)))), 10)
	}
}

func CheckStringFormat(format string) interface{} {
	switch format {
	case "date":
		return gofakeit.Generate("####-##-##")
	case "date-time":
		return gofakeit.Date()
	case "password":

		// Add support to change password length

		randLen := gofakeit.IntRange(0, 255)
		return gofakeit.Password(true, true, true, true, true, randLen)
	case "byte":
		randLen := gofakeit.IntRange(int(math.Inf(-1)), int(math.Inf(1)))
		randStr := gofakeit.LetterN(uint(randLen))
		return base64.StdEncoding.EncodeToString([]byte(randStr))
	case "binary":
		randLen := gofakeit.IntRange(0, 1024)
		return []byte(gofakeit.LetterN(uint(randLen)))
	case "email":
		return gofakeit.Email()
	case "uuid":
		return gofakeit.UUID()
	case "uri":
		return gofakeit.URL()
	case "hostname":
		return gofakeit.DomainName() + gofakeit.DomainSuffix()
	case "ipv4":
		return gofakeit.IPv4Address()
	case "ipv6":
		return gofakeit.IPv6Address()
	default:
		randLen := gofakeit.IntRange(int(math.Inf(-1)), int(math.Inf(1)))
		return gofakeit.LetterN(uint(randLen))
	}
}

func CheckStringPattern(pattern string) interface{} {
	return gofakeit.Regex(pattern)
}
