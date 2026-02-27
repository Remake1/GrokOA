package room

import "errors"

var ErrCodeCollision = errors.New("failed to generate unique room code")
