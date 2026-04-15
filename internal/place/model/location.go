package model

type PlaceSuggestion struct {
	Token        string
	Label        string
	Name         string
	AddressLine  string
	CityName     string
	CountryName  string
	Timezone     string
	Latitude     float64
	Longitude    float64
	Postcode     string
	District     string
	PhotonSource PlaceSuggestionSource
}

type PlaceSuggestionSource struct {
	OSMID    int64
	OSMType  string
	OSMKey   string
	OSMValue string
}

type PlaceResolveInput struct {
	Token string
}

type PlaceResolved struct {
	ID          string
	CityID      string
	Name        string
	AddressLine string
	CityName    string
	CountryName string
	Timezone    string
	Latitude    float64
	Longitude   float64
}
