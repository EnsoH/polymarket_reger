package utils

import (
	"time"

	"github.com/goombaio/namegenerator"
)

func GenerateRandomName() string {
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)

	name := nameGenerator.Generate()

	return name
}
