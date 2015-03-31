package functions

func addHandler(input interface{}) (interface{}, error) {
	if err := validatePayload("$add", input); err != nil {
		return nil, err
	}

	data := input.(map[string]interface{})

	return data["$left"].(float64) + data["$right"].(float64), nil
}
