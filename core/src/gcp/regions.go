package gcp
type Region struct {
	ID          string
	Description string
}
func GetRegions() []Region {
	return []Region{
		{ID: "us-central1", Description: "Iowa"},
		{ID: "us-east1", Description: "South Carolina"},
		{ID: "us-east4", Description: "Northern Virginia"},
		{ID: "us-east5", Description: "Columbus"},
		{ID: "us-south1", Description: "Dallas"},
		{ID: "us-west1", Description: "Oregon"},
		{ID: "us-west2", Description: "Los Angeles"},
		{ID: "us-west3", Description: "Salt Lake City"},
		{ID: "us-west4", Description: "Las Vegas"},
		{ID: "northamerica-northeast1", Description: "Montréal"},
		{ID: "northamerica-northeast2", Description: "Toronto"},
		{ID: "southamerica-east1", Description: "São Paulo"},
		{ID: "southamerica-west1", Description: "Santiago"},
		{ID: "europe-central2", Description: "Warsaw"},
		{ID: "europe-north1", Description: "Finland"},
		{ID: "europe-southwest1", Description: "Madrid"},
		{ID: "europe-west1", Description: "Belgium"},
		{ID: "europe-west2", Description: "London"},
		{ID: "europe-west3", Description: "Frankfurt"},
		{ID: "europe-west4", Description: "Netherlands"},
		{ID: "europe-west6", Description: "Zurich"},
		{ID: "europe-west8", Description: "Milan"},
		{ID: "europe-west9", Description: "Paris"},
		{ID: "europe-west10", Description: "Berlin"},
		{ID: "europe-west12", Description: "Turin"},
		{ID: "asia-east1", Description: "Taiwan"},
		{ID: "asia-east2", Description: "Hong Kong"},
		{ID: "asia-northeast1", Description: "Tokyo"},
		{ID: "asia-northeast2", Description: "Osaka"},
		{ID: "asia-northeast3", Description: "Seoul"},
		{ID: "asia-south1", Description: "Mumbai"},
		{ID: "asia-south2", Description: "Delhi"},
		{ID: "asia-southeast1", Description: "Singapore"},
		{ID: "asia-southeast2", Description: "Jakarta"},
		{ID: "australia-southeast1", Description: "Sydney"},
		{ID: "australia-southeast2", Description: "Melbourne"},
		{ID: "me-central1", Description: "Doha"},
		{ID: "me-central2", Description: "Dammam"},
		{ID: "me-west1", Description: "Tel Aviv"},
		{ID: "africa-south1", Description: "Johannesburg"},
	}
}