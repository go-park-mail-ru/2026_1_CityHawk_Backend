package usecase_test

import (
	"context"
	"errors"
	"testing"

	photonintegration "cityhawk/backend/internal/integration/photon"
	"cityhawk/backend/internal/mocks"
	placemodel "cityhawk/backend/internal/place/model"
	placerepo "cityhawk/backend/internal/place/repository"
	placeusecase "cityhawk/backend/internal/place/usecase"
	"github.com/golang/mock/gomock"
)

func TestPlaceLookupServiceSuggestAndResolve(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	photon := mocks.NewMockPhotonSearcher(ctrl)
	repo := mocks.NewMockPlaceLookupRepository(ctrl)
	svc := placeusecase.NewPlaceLookupService(photon, repo, []byte("secret"), "Russia", "Europe/Moscow")

	photon.EXPECT().Search(gomock.Any(), "vdnh", 5).Return([]photonintegration.SearchResult{
		{
			Name:         "ВДНХ",
			Street:       "Проспект Мира",
			HouseNumber:  "119",
			City:         "Москва",
			Country:      "Россия",
			DisplayLabel: "ВДНХ, Москва",
			Latitude:     55.0,
			Longitude:    37.0,
			OSMID:        10,
			OSMType:      "W",
			OSMKey:       "tourism",
			OSMValue:     "attraction",
		},
		{
			Name: "",
		},
	}, nil)

	suggestions, err := svc.Suggest(context.Background(), "vdnh", 5)
	if err != nil {
		t.Fatalf("Suggest() error = %v", err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("len(suggestions) = %d, want 1", len(suggestions))
	}
	if suggestions[0].Token == "" || suggestions[0].AddressLine != "Проспект Мира 119" {
		t.Fatalf("unexpected suggestion: %+v", suggestions[0])
	}

	repo.EXPECT().EnsurePlace(gomock.Any(), placerepo.PlaceLookupWriteInput{
		CityName:    "Москва",
		CountryName: "Россия",
		Timezone:    "Europe/Moscow",
		Name:        "ВДНХ",
		AddressLine: "Проспект Мира 119",
		Latitude:    55.0,
		Longitude:   37.0,
	}).Return(placemodel.PlaceResolved{ID: "place-1"}, nil)

	resolved, err := svc.Resolve(context.Background(), placemodel.PlaceResolveInput{Token: suggestions[0].Token})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if resolved.ID != "place-1" {
		t.Fatalf("unexpected resolved place: %+v", resolved)
	}
}

func TestPlaceLookupServiceErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	photon := mocks.NewMockPhotonSearcher(ctrl)
	repo := mocks.NewMockPlaceLookupRepository(ctrl)
	svc := placeusecase.NewPlaceLookupService(photon, repo, []byte("secret"), "Russia", "UTC")

	photon.EXPECT().Search(gomock.Any(), "bad", 3).Return(nil, errors.New("search failed"))
	if _, err := svc.Suggest(context.Background(), "bad", 3); err == nil {
		t.Fatal("expected search error")
	}

	if _, err := svc.Resolve(context.Background(), placemodel.PlaceResolveInput{Token: "bad.token"}); !errors.Is(err, placeusecase.ErrInvalidPlaceSuggestion) {
		t.Fatalf("error = %v, want ErrInvalidPlaceSuggestion", err)
	}
}
