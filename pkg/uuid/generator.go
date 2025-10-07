package uuid

import "github.com/gofrs/uuid/v5"

func GenerateUUID7() (string, error) {
	u, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
