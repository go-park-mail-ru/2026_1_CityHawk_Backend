package main

import "net/http"

type homePlaceCard struct {
	ImageURL    string `json:"imageUrl"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type moodCard struct {
	ImageURL string `json:"imageUrl"`
	Title    string `json:"title"`
	Modifier string `json:"modifier,omitempty"`
}

type homeMoodTall struct {
	ImageURL string `json:"imageUrl"`
	Title    string `json:"title"`
}

type homePayload struct {
	Places   []homePlaceCard `json:"places"`
	MoodLeft []moodCard      `json:"moodLeft"`
	MoodTall homeMoodTall    `json:"moodTall"`
}

func homeHandler() http.HandlerFunc {
	payload := homePayload{
		Places: []homePlaceCard{
			{ImageURL: "/public/static/img/futurione.jpeg", Title: "Futurione", Description: "ВДНХ, Москва"},
			{ImageURL: "/public/static/img/navka.jpeg", Title: "Ледовое шоу Татьяны Навки", Description: "10 февраля - 21 марта, Navka arena, Москва"},
			{ImageURL: "/public/static/img/standup.png", Title: "Женский стендап", Description: "27 марта, Live Арена, Москва"},
			{ImageURL: "/public/static/img/balet.jpg", Title: "Балет Щелкунчик", Description: "14 марта, Большой театр, Москва"},
			{ImageURL: "/public/static/img/art.png", Title: "Искусство XX века", Description: "10–24 марта, Третьяковская галерея, Москва"},
			{ImageURL: "/public/static/img/rok.jpg", Title: "Рок-концерт", Description: "20 марта, Атмосфера, Москва"},
			{ImageURL: "/public/static/img/concert.jpeg", Title: "Руки Вверх", Description: "17 марта, Лужники, Москва"},
			{ImageURL: "/public/static/img/horror.jpg", Title: "Квест Искупление", Description: "улица Пруд-Ключики, 5, Москва"},
		},
		MoodLeft: []moodCard{
			{ImageURL: "/public/static/img/club.jpeg", Title: "Ночная жизнь", Modifier: "mood-card--wide"},
			{ImageURL: "/public/static/img/eat.jpg", Title: "Гастро-места"},
			{ImageURL: "/public/static/img/love.jpg", Title: "Романтический вечер"},
		},
		MoodTall: homeMoodTall{
			ImageURL: "/public/static/img/photo.jpeg",
			Title:    "Места для фото",
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		writeJSON(w, http.StatusOK, payload)
	}
}
