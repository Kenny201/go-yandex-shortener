package storage

var urlStorage = map[string]string{}

func init() {
	urlStorage = make(map[string]string)
}

func GetStorage() *map[string]string {
	return &urlStorage
}

func CheckExistsValueIntoURLStorage(shortValue string) (string, bool) {
	for key, value := range urlStorage {
		if value == shortValue {
			return key, true
		}
	}

	return "", false
}
