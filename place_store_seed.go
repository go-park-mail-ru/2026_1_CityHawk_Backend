package main

func newPlaceStore() *placeStore {
	s := &placeStore{
		byID: make(map[string]place),
	}

	for _, p := range []place{
		{
			ID:               "futurione",
			Title:            "Futurione",
			Categories:       []string{"park", "exhibition", "family"},
			LikeCount:        150,
			ShortDescription: "ВДНХ, Москва",
			FullDescription:  "Иммерсивная выставка Futurione на территории ВДНХ с мультимедийными инсталляциями и интерактивными зонами. Подходит для посещения с друзьями и семьей.",
			Address:          "Москва, проспект Мира, 119",
			ImageURL:         "/public/static/img/futurione.jpeg",
			WorkingHours:     "ежедневно, 10:00-22:00",
			PriceLevel:       "medium",
		},
		{
			ID:               "navka-show",
			Title:            "Ледовое шоу Татьяны Навки",
			Categories:       []string{"show", "family"},
			LikeCount:        132,
			ShortDescription: "10 февраля - 21 марта, Navka arena, Москва",
			FullDescription:  "Большое ледовое шоу Татьяны Навки с постановочными номерами и театральной драматургией. Формат подходит для вечернего досуга и семейного похода.",
			Address:          "Москва, Navka arena",
			ImageURL:         "/public/static/img/navka.jpeg",
			WorkingHours:     "по расписанию площадки",
			PriceLevel:       "medium",
		},
		{
			ID:               "womens-standup",
			Title:            "Женский стендап",
			Categories:       []string{"show", "comedy"},
			LikeCount:        121,
			ShortDescription: "27 марта, Live Арена, Москва",
			FullDescription:  "Концертный стендап-формат с выступлениями резидентов и актуальными монологами. Подойдет для компании друзей и насыщенного вечернего досуга.",
			Address:          "Москва, Live Арена",
			ImageURL:         "/public/static/img/standup.png",
			WorkingHours:     "по расписанию площадки",
			PriceLevel:       "medium",
		},
		{
			ID:               "balet-shelkunchik",
			Title:            "Балет Щелкунчик",
			Categories:       []string{"theatre", "classic"},
			LikeCount:        118,
			ShortDescription: "14 марта, Большой театр, Москва",
			FullDescription:  "Классическая постановка балета «Щелкунчик» на исторической сцене Большого театра. Подходит для романтического вечера и культурного маршрута.",
			Address:          "Москва, Большой театр",
			ImageURL:         "/public/static/img/balet.jpg",
			WorkingHours:     "по расписанию театра",
			PriceLevel:       "medium",
		},
		{
			ID:               "art-xx",
			Title:            "Искусство XX века",
			Categories:       []string{"museum", "photo"},
			LikeCount:        110,
			ShortDescription: "10–24 марта, Третьяковская галерея, Москва",
			FullDescription:  "Выставочный проект об искусстве XX века с работами ключевых авторов и тематическими залами. Хороший выбор для вдумчивого культурного посещения.",
			Address:          "Москва, Третьяковская галерея",
			ImageURL:         "/public/static/img/art.png",
			WorkingHours:     "по расписанию музея",
			PriceLevel:       "medium",
		},
		{
			ID:               "rock-concert",
			Title:            "Рок-концерт",
			Categories:       []string{"concert", "music"},
			LikeCount:        104,
			ShortDescription: "20 марта, Атмосфера, Москва",
			FullDescription:  "Большой рок-концерт с живым звуком и вечерней программой. Подходит для любителей концертного формата и активного отдыха.",
			Address:          "Москва, Атмосфера",
			ImageURL:         "/public/static/img/rok.jpg",
			WorkingHours:     "по расписанию площадки",
			PriceLevel:       "medium",
		},
		{
			ID:               "ruki-vverh",
			Title:            "Руки Вверх",
			Categories:       []string{"concert", "music"},
			LikeCount:        127,
			ShortDescription: "17 марта, Лужники, Москва",
			FullDescription:  "Концерт группы «Руки Вверх» на большой площадке с ретро-хитами и масштабным шоу. Формат рассчитан на вечерний досуг и большую компанию.",
			Address:          "Москва, Лужники",
			ImageURL:         "/public/static/img/concert.jpeg",
			WorkingHours:     "по расписанию площадки",
			PriceLevel:       "medium",
		},
		{
			ID:               "quest-iskuplenie",
			Title:            "Квест Искупление",
			Categories:       []string{"quest", "adventure"},
			LikeCount:        96,
			ShortDescription: "улица Пруд-Ключики, 5, Москва",
			FullDescription:  "Атмосферный квест «Искупление» с сюжетными загадками и командным прохождением. Подходит для небольших групп и вечернего досуга.",
			Address:          "Москва, улица Пруд-Ключики, 5",
			ImageURL:         "/public/static/img/horror.jpg",
			WorkingHours:     "ежедневно, 10:00-23:00",
			PriceLevel:       "medium",
		},
	} {
		s.byID[p.ID] = p
		s.order = append(s.order, p.ID)
	}

	return s
}
