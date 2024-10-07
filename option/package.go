package option

import (
	"os"
)

func Getenv(key string) String {
	result := os.Getenv(key)

	return IfProvided(result)
}
