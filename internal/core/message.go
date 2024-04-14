package core

import (
	"time"
)

type Message struct {
	CreatedAt time.Time `bson:"createdAt"`
	Content   string    `bson:"content"`
	CreatedBy string    `bson:"createdBy"`
}
