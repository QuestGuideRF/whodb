package azure
const (
	GeographyAmericas   = "Americas"
	GeographyEurope     = "Europe"
	GeographyAsiaPac    = "Asia Pacific"
	GeographyMiddleEast = "Middle East"
	GeographyAfrica     = "Africa"
)
type Region struct {
	ID          string
	DisplayName string
	Geography   string
}
func GetRegions() []Region {
	return []Region{
		{ID: "eastus", DisplayName: "East US", Geography: GeographyAmericas},
		{ID: "eastus2", DisplayName: "East US 2", Geography: GeographyAmericas},
		{ID: "centralus", DisplayName: "Central US", Geography: GeographyAmericas},
		{ID: "northcentralus", DisplayName: "North Central US", Geography: GeographyAmericas},
		{ID: "southcentralus", DisplayName: "South Central US", Geography: GeographyAmericas},
		{ID: "westcentralus", DisplayName: "West Central US", Geography: GeographyAmericas},
		{ID: "westus", DisplayName: "West US", Geography: GeographyAmericas},
		{ID: "westus2", DisplayName: "West US 2", Geography: GeographyAmericas},
		{ID: "westus3", DisplayName: "West US 3", Geography: GeographyAmericas},
		{ID: "canadacentral", DisplayName: "Canada Central", Geography: GeographyAmericas},
		{ID: "canadaeast", DisplayName: "Canada East", Geography: GeographyAmericas},
		{ID: "brazilsouth", DisplayName: "Brazil South", Geography: GeographyAmericas},
		{ID: "brazilsoutheast", DisplayName: "Brazil Southeast", Geography: GeographyAmericas},
		{ID: "mexicocentral", DisplayName: "Mexico Central", Geography: GeographyAmericas},
		{ID: "northeurope", DisplayName: "North Europe (Ireland)", Geography: GeographyEurope},
		{ID: "westeurope", DisplayName: "West Europe (Netherlands)", Geography: GeographyEurope},
		{ID: "uksouth", DisplayName: "UK South", Geography: GeographyEurope},
		{ID: "ukwest", DisplayName: "UK West", Geography: GeographyEurope},
		{ID: "francecentral", DisplayName: "France Central", Geography: GeographyEurope},
		{ID: "francesouth", DisplayName: "France South", Geography: GeographyEurope},
		{ID: "germanywestcentral", DisplayName: "Germany West Central", Geography: GeographyEurope},
		{ID: "germanynorth", DisplayName: "Germany North", Geography: GeographyEurope},
		{ID: "switzerlandnorth", DisplayName: "Switzerland North", Geography: GeographyEurope},
		{ID: "switzerlandwest", DisplayName: "Switzerland West", Geography: GeographyEurope},
		{ID: "norwayeast", DisplayName: "Norway East", Geography: GeographyEurope},
		{ID: "norwaywest", DisplayName: "Norway West", Geography: GeographyEurope},
		{ID: "swedencentral", DisplayName: "Sweden Central", Geography: GeographyEurope},
		{ID: "polandcentral", DisplayName: "Poland Central", Geography: GeographyEurope},
		{ID: "italynorth", DisplayName: "Italy North", Geography: GeographyEurope},
		{ID: "spaincentral", DisplayName: "Spain Central", Geography: GeographyEurope},
		{ID: "austriaeast", DisplayName: "Austria East", Geography: GeographyEurope},
		{ID: "eastasia", DisplayName: "East Asia (Hong Kong)", Geography: GeographyAsiaPac},
		{ID: "southeastasia", DisplayName: "Southeast Asia (Singapore)", Geography: GeographyAsiaPac},
		{ID: "japaneast", DisplayName: "Japan East", Geography: GeographyAsiaPac},
		{ID: "japanwest", DisplayName: "Japan West", Geography: GeographyAsiaPac},
		{ID: "australiaeast", DisplayName: "Australia East", Geography: GeographyAsiaPac},
		{ID: "australiasoutheast", DisplayName: "Australia Southeast", Geography: GeographyAsiaPac},
		{ID: "australiacentral", DisplayName: "Australia Central", Geography: GeographyAsiaPac},
		{ID: "koreacentral", DisplayName: "Korea Central", Geography: GeographyAsiaPac},
		{ID: "koreasouth", DisplayName: "Korea South", Geography: GeographyAsiaPac},
		{ID: "centralindia", DisplayName: "Central India", Geography: GeographyAsiaPac},
		{ID: "southindia", DisplayName: "South India", Geography: GeographyAsiaPac},
		{ID: "westindia", DisplayName: "West India", Geography: GeographyAsiaPac},
		{ID: "newzealandnorth", DisplayName: "New Zealand North", Geography: GeographyAsiaPac},
		{ID: "indonesiacentral", DisplayName: "Indonesia Central", Geography: GeographyAsiaPac},
		{ID: "malaysiawest", DisplayName: "Malaysia West", Geography: GeographyAsiaPac},
		{ID: "taiwannorth", DisplayName: "Taiwan North", Geography: GeographyAsiaPac},
		{ID: "uaenorth", DisplayName: "UAE North", Geography: GeographyMiddleEast},
		{ID: "uaecentral", DisplayName: "UAE Central", Geography: GeographyMiddleEast},
		{ID: "qatarcentral", DisplayName: "Qatar Central", Geography: GeographyMiddleEast},
		{ID: "israelcentral", DisplayName: "Israel Central", Geography: GeographyMiddleEast},
		{ID: "saudiarabiacentral", DisplayName: "Saudi Arabia Central", Geography: GeographyMiddleEast},
		{ID: "southafricanorth", DisplayName: "South Africa North", Geography: GeographyAfrica},
		{ID: "southafricawest", DisplayName: "South Africa West", Geography: GeographyAfrica},
	}
}