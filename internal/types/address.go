package types

type Address struct {
	Found    bool
	CacheHit bool

	Latitude     float64
	Longitude    float64
	Altitude     float64
	DisplayName  string
	Country      string
	State        string
	Municipality string
	CityDistrict string
	Postcode     string
	City         string
	Suburb       string
	Street       string
	HouseNumber  string
}
