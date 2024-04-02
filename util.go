package nebula_go

func indexOf(collection []string, element string) int {
	for i, item := range collection {
		if item == element {
			return i
		}
	}

	return -1
}
