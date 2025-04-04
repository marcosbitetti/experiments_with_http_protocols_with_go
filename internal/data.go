package internal

// JSendResponse follows the JSEND spec with a "status" and a "data" field.
type JSendResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}
