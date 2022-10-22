package constants

import "os"

var (
	MONGO_URI          = os.Getenv("MONGO_URI")
	DATABASE           = os.Getenv("DATABASE")
	PERSONS_COLLECTION = os.Getenv("PERSONS_COLLECTION")
)
