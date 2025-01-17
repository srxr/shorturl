package main

import (
	"fmt"
	nurl "net/url"
	"strings"
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
	u, err := parse(target)
	if err != nil {
		return nil, err
	}

	url = &URL{ID: GenerateID(), URL: u.String(), CreatedAt: time.Now()}
	err = db.Save(url)
	return
}

func parse(target string) (u *nurl.URL, err error) {
	u, err = nurl.Parse(strings.TrimSpace(target))
	if err != nil {
		return nil, fmt.Errorf("URL (%s) no satisfied", target)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("URL (%s) without scheme or host", target)
	}
	return u, nil
}

// SetName ...
func (u *URL) SetName(name string) error {
	u.Name = name
	u.UpdatedAt = time.Now()
	return db.Save(&u)
}

func (u *URL) update(id, target string) error {
	url, err := parse(target)
	if err != nil {
		return err
	}

	if err := del(u.ID); err != nil {
		return err
	}

	u.ID = id
	u.URL = url.String()
	u.UpdatedAt = time.Now()
	return db.Save(u)
}

func del(id string) error {
	var u URL

	err := db.One("ID", id, &u)
	if err != nil {
		return err
	}

	err = db.DeleteStruct(&u)
	if err != nil {
		return err
	}

	return nil
}
