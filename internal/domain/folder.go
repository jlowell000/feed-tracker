package domain

import "time"

type Folder struct {
	ID        string
	Name      string
	CreatedAt time.Time
}
