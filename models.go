package main

import (
	"time"
)

// URL ...
type URL struct {
	ID        string    `storm:"id"`
	URL       string    `storm:"index"`
	Name      string    `storm:"index"`
	CreatedAt time.Time `storm:"index"`
	UpdatedAt time.Time `storm:"index"`
}

func GenerateID() string {
	for {
		// TODO: Make length (5) configurable
		id := RandomString(5)
		err := db.One("ID", id, nil)
		if err != nil {
			return id
		}
	}
}

func NewURL(target string) (url *URL, err error) {
	url = &URL{ID: GenerateID(), URL: target, CreatedAt: time.Now()}
	err = db.Save(url)
	return
}

// SetName ...
func (u *URL) SetName(name string) error {
	u.Name = name
	u.UpdatedAt = time.Now()
	return db.Save(&u)
}
