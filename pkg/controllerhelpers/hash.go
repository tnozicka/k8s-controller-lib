package controllerhelpers

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
)

func HashObjects(objs ...any) (string, error) {
	hasher := sha512.New()
	encoder := json.NewEncoder(hasher)
	for _, obj := range objs {
		if err := encoder.Encode(obj); err != nil {
			return "", err
		}
	}

	return base64.StdEncoding.EncodeToString(hasher.Sum(nil)), nil
}
