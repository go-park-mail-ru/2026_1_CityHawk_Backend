package repository

import (
	"time"

	placemodel "cityhawk/backend/internal/place/model"
)

func SeedPlaces() []placemodel.EventDetailsView {
	city := placemodel.EventSessionPlaceCityView{
		ID:          "city-moscow",
		Name:        "Москва",
		CountryName: "Россия",
		Timezone:    "Europe/Moscow",
	}

	author := placemodel.EventAuthorView{
		ID:       "user-1",
		Username: "seed_author",
	}

	return []placemodel.EventDetailsView{
		newSeedEvent("futurione", "Futurione", "ВДНХ, Москва", "Иммерсивная выставка Futurione на территории ВДНХ с мультимедийными инсталляциями и интерактивными зонами. Подходит для посещения с друзьями и семьей.", []string{"Парк", "Выставка", "Семья"}, []string{"Иммерсивное", "Семейное"}, "https://example.com/futurione.jpg", "place-vdnh", "ВДНХ", "Москва, проспект Мира, 119", city, author, 0, nil, time.Date(2026, time.April, 12, 10, 0, 0, 0, time.UTC), time.Date(2026, time.April, 12, 22, 0, 0, 0, time.UTC), 1200),
		newSeedEvent("navka-show", "Ледовое шоу Татьяны Навки", "10 февраля - 21 марта, Navka arena, Москва", "Большое ледовое шоу Татьяны Навки с постановочными номерами и театральной драматургией. Формат подходит для вечернего досуга и семейного похода.", []string{"Шоу", "Семья"}, []string{"Ледовое шоу", "Семейное"}, "https://example.com/navka.jpg", "place-navka", "Navka Arena", "Москва, Navka arena", city, author, 6, nil, time.Date(2026, time.April, 13, 19, 0, 0, 0, time.UTC), time.Date(2026, time.April, 13, 21, 30, 0, 0, time.UTC), 2500),
		newSeedEvent("rock-concert", "Рок-концерт", "20 марта, Атмосфера, Москва", "Большой рок-концерт с живым звуком и вечерней программой. Подходит для любителей концертного формата и активного отдыха.", []string{"Музыка"}, []string{"Рок"}, "https://example.com/rock.jpg", "place-atmosphere", "Атмосфера", "Москва, Атмосфера", city, author, 12, strPtr("https://cityhawk.local/events/rock-concert"), time.Date(2026, time.April, 14, 19, 30, 0, 0, time.UTC), time.Date(2026, time.April, 14, 22, 30, 0, 0, time.UTC), 2200),
	}
}

func newSeedEvent(id, title, shortDescription, fullDescription string, categoryNames, tagNames []string, imageURL, placeID, placeName, address string, city placemodel.EventSessionPlaceCityView, author placemodel.EventAuthorView, ageLimit int, sourceURL *string, startAt, endAt time.Time, price int) placemodel.EventDetailsView {
	now := time.Date(2026, time.April, 10, 10, 0, 0, 0, time.UTC)
	return placemodel.EventDetailsView{
		ID:               id,
		Title:            title,
		ShortDescription: shortDescription,
		FullDescription:  fullDescription,
		AgeLimit:         ageLimit,
		SourceURL:        sourceURL,
		Author:           author,
		Categories:       taxonomyFromNames(categoryNames),
		Tags:             taxonomyFromNames(tagNames),
		Images: []placemodel.EventImageView{
			{ID: "img-" + id, ImageURL: imageURL},
		},
		Sessions: []placemodel.EventSessionView{
			{
				ID:      "session-" + id,
				StartAt: startAt,
				EndAt:   endAt,
				Price:   price,
				Place: placemodel.EventSessionPlaceView{
					ID:          placeID,
					Name:        placeName,
					AddressLine: address,
					Latitude:    55.75,
					Longitude:   37.61,
					City:        city,
				},
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func taxonomyFromNames(names []string) []placemodel.EventTaxonomyItem {
	items := make([]placemodel.EventTaxonomyItem, 0, len(names))
	for _, name := range names {
		items = append(items, placemodel.EventTaxonomyItem{
			ID:   slugify(name),
			Name: name,
			Slug: slugify(name),
		})
	}
	return items
}

func strPtr(value string) *string {
	return &value
}
