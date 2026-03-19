package domain

import "time"

type Client struct {
	ID        string    `json:"id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	BirthDate time.Time `json:"birth_date"`
	Country   string    `json:"country"`
	CreatedAt time.Time `json:"created_at"`
}

func (c Client) Age(at time.Time) int {
	years := at.Year() - c.BirthDate.Year()
	if at.YearDay() < c.BirthDate.YearDay() {
		years--
	}

	return years
}
