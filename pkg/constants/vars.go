package constants

import (
	"os"

	"github.com/go-playground/validator/v10"
)

var (
	MONGO_URI          = os.Getenv("MONGO_URI")
	DATABASE           = os.Getenv("DATABASE")
	PERSONS_COLLECTION = os.Getenv("PERSONS_COLLECTION")
)

var Validate = validator.New()
